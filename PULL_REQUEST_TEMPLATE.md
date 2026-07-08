# Pull Request #18 — Fase 6: Testes Unitários e de Integração com Cassete

> **Branch:** `feat/fase-6-testes`
> **Base:** `main`
> **Status:** 🟢 Pronto para merge

---

## 📋 Resumo

Esta PR implementa a **Fase 6** do projeto tl;dr, adicionando **testes de integração com gravação/reprodução via cassete (go-vcr)**, além de correções de segurança, performance e robustez identificadas no code review. As principais áreas afetadas são: testes de integração offline com HTTP recording/replay, thread-safety do Client, validação TLS, redação de credenciais exportada, pré-filtro de keywords em sanitizePrompt, e correção da ordem de classificação de erros da API.

---

## 🔴 Breaking Changes

### `ProtectedAPIKey.Bytes()` — agora retorna cópia defensiva (não referência direta)

```go
// Antes
b := key.Bytes()
b[0] = 'X'        // Modificava o slice interno da chave!
key.Get() == "Xk-..." // true — vazamento via mutação

// Depois
b := key.Bytes()
b[0] = 'X'        // Não afeta o original
key.Get() == "sk-..." // true — cópia defensiva
```

**Ação:** Código que dependia da mutação do slice retornado por `Bytes()` para modificar a chave precisa ser revisado. A nova semântica é de cópia defensiva.

### `Config.APIKeyBytes()` — renomeado para `apiKeyBytes()` (não exportado)

- **Antes:** `cfg.APIKeyBytes()` — exportado, disponível para callers externos
- **Depois:** `cfg.apiKeyBytes()` — não exportado, uso interno apenas
- **Ação:** Código externo que chamava `cfg.APIKeyBytes()` agora deve acessar a chave via `cfg.APIKey` (string) ou via `secrets.ProtectedAPIKey.Bytes()`

### `checkEnvPermissions()` — renomeado e com assinatura alterada

- **Antes:** `checkEnvPermissions()` — fixo para `.env`, sem parâmetros
- **Depois:** `checkFilePermissions(path string)` — genérico, recebe caminho do arquivo
- **Ação:** Código que chamava `checkEnvPermissions()` precisa ser atualizado para `checkFilePermissions(".env")`

### `envPermsWarn` — renomeado para `filePermsWarn`

- Constante não exportada — afeta apenas código interno
- Mensagem de warning alterada de português (`"⚠️  AVISO:"`) para inglês (`"⚠️  WARNING:"`)
- Mensagem agora inclui o caminho do arquivo específico nas permissões

---

## 🟢 Novas Funcionalidades

### Testes de Integração com Cassete (go-vcr)

| Funcionalidade | Descrição |
|----------------|-----------|
| 🎭 **`go-vcr` v3.2.0** | Dependência adicionada para gravação/reprodução de interações HTTP |
| 📼 **5 cassetes YAML** | `summarize_success`, `summarize_short`, `summarize_unauthorized`, `summarize_rate_limited`, `summarize_server_error` |
| 🔒 **Hook BeforeSaveHook** | Redige header `Authorization` e body com credenciais antes de persistir |
| 🏷️ **Build tag `integration`** | Testes isolados: `go test -tags=integration ./internal/integration/` |
| 🌐 **Modo Record/Replay** | `TLDR_CASSETE_MODE=record` para gravar; padrão é replay offline |
| 🔁 **`TLDR_CASSETE_MODE`** | Variável de ambiente que controla modo de operação dos cassetes |

### Segurança

| Funcionalidade | Descrição |
|----------------|-----------|
| 🛡️ **`validateHTTPClientTLS()`** | Rejeita `http.Client` com `InsecureSkipVerify=true` no `summarizer.New()` |
| 🔑 **`RedactCredentials()` exportada** | Função pública para redigir credenciais em qualquer contexto (reuso nos hooks do cassete) |
| 🔒 **`Client.mu sync.Mutex`** | Thread-safety entre `Summarize()` e `Clear()` — sem race condition |
| 📋 **`Client.cleared` flag** | Detecta uso após Clear() e retorna erro (em vez de panic) |
| 🧪 **Cópia defensiva da apiKey** | `Summarize()` copia a chave antes do unlock para uso seguro em `classifyAPIError` |
| 🗑️ **Lone continuation bytes** | `sanitizeOutput()` descarta bytes 0x80-0xBF isolados (sem lead byte) |

### Performance

| Funcionalidade | Descrição |
|----------------|-----------|
| ⚡ **Pré-filtro de keywords** | `sanitizePrompt()` verifica keywords primeiro; se ausentes, pula scan completo com regexes |
| 🧩 **`injectionKeywords`** | Lista expandida: `ignore`, `reveal`, `show`, `system`, `user`, `assistant`, etc. |

### Engine

| Funcionalidade | Descrição |
|----------------|-----------|
| 🔧 **`Config.HTTPClient`** | Campo opcional para injetar `http.Client` customizado (ex: go-vcr recorder) |
| 🔄 **`classifyAPIError` reordenado** | Erros HTTP (401, 429, 500) detectados **antes** de fallback por substring |
| 🚫 **Release job no CI** | Impedido de rodar em eventos de `pull_request` (`github.event_name != 'pull_request'`) |

---

## 🔧 Melhorias Técnicas

### Correções de segurança/qualidade

| Issue | Solução |
|-------|---------|
| **`ProtectedAPIKey.Bytes()` retornava referência direta** | Agora retorna cópia defensiva (`make + copy`) — mutação externa não afeta interno |
| **Race condition entre Summarize() e Clear()** | Mutex (`s.mu.Lock()`) protege acesso a `apiKey` e `cleared` |
| **Timeouts falsos em erros HTTP** | `classifyAPIError` agora verifica status HTTP antes de fallback por substring |
| **Lint SA9003** | Branch vazia removida em `validateHTTPClientTLS` |
| **Permissões do `TLDR_API_KEY_FILE`** | `checkFilePermissions` agora também verifica permissões do arquivo de chave |
| **Erros de `os.Stat` não eram logados** | Erros inesperados (EACCES, broken symlink) agora são reportados no stderr |
| **Aviso combinado de permissões** | Se arquivo for legível para grupo E outros, ambos são mencionados |
| **Release em PR** | Job de release agora pula quando `github.event_name == 'pull_request'` |

### Falsos positivos eliminados em sanitizePrompt

| Palavra | Problema | Solução |
|---------|----------|---------|
| `system` | "O sistema está rodando" era falsamente acionado | Removido de `injectionPatterns`, mantido apenas em `injectionKeywords` para pré-filtro |
| `user` | "O usuário informou" era capturado | Mesma abordagem: keyword apenas, sem pattern standalone |
| `assistant` | "Assistente de pesquisa" era bloqueado | Mesma abordagem |
| `now` | "Resuma agora" era capturado (já removido no PR#15) | Mantido removido |

### Padrões de código

| Melhoria | Detalhe |
|----------|---------|
| **`checkFilePermissions` genérico** | Agora recebe `path string` em vez de ser fixo em `.env` |
| **`filePermsWarn` em inglês** | Universal — warning de segurança em inglês para alcançar mais desenvolvedores |
| **`RedactCredentials` exportada** | Reutilizável em hooks de cassete e outros contextos |
| **`classifyAPIError` recebe `[]byte`** | Cópia defensiva evita race com Clear() |
| **Tracking UTF-8 apenas em bytes escritos** | Bytes descartados (ESC, C0, C1) nunca alteram `utf8Remaining` |
| **Remoção de `Now()` em testes** | `time.Now()` substituído por `time.Date()` para determinismo |

### CI/CD

| Melhoria | Detalhe |
|----------|---------|
| **`gosec` exclui `internal/integration`** | Testes de integração com cassetes não são escaneados (chaves fictícias) |
| **Release condicional** | `github.event_name != 'pull_request'` impede release acidental em PRs |

---

## 🧪 Testes

### Testes de Integração com Cassete (+332 linhas)

| Teste | O que cobre |
|-------|-------------|
| `TestSummarizeWithCassette` | Chamada bem-sucedida com texto longo em pt-br |
| `TestSummarizeWithCassetteShortText` | Chamada bem-sucedida com texto curto em en |
| `TestSummarizeAuthError` | Erro 401 — credenciais inválidas mapeado para mensagem amigável |
| `TestSummarizeRateLimitError` | Erro 429 — rate limit mapeado para mensagem amigável |
| `TestSummarizeServerError` | Erro 500 — servidor indisponível mapeado para mensagem amigável |
| `TestSummarizeContextCanceled` | Contexto cancelado — erro tratado sem panic |
| `TestSummarizeContextDeadline` | Deadline expirado — erro de timeout detectado |

### Cassetes Gravados (5 arquivos YAML)

| Cassete | Cenário | Status HTTP |
|---------|---------|-------------|
| `summarize_success.yaml` | Resumo bem-sucedido (pt-br) | 200 |
| `summarize_short.yaml` | Resumo bem-sucedido (en, texto curto) | 200 |
| `summarize_unauthorized.yaml` | Chave inválida | 401 |
| `summarize_rate_limited.yaml` | Rate limit excedido | 429 |
| `summarize_server_error.yaml` | Erro interno do servidor | 500 |

### Testes de Unidade Adicionados/Modificados

| Teste | O que cobre |
|-------|-------------|
| `TestCheckFilePermissions` | 8 casos: arquivo inexistente, 0600, 0640 (grupo), 0644 (outros), subdiretório, diretório, 000 (sem permissão), erro EACCES no stat |
| `TestExecuteGlobal` | Execute() lazy-init com sync.Once, idempotência, --help funciona |
| `TestInitRootCommand` | Execute() não panica, flags --lang/--model/--prompt registradas |
| `TestSanitizePrompt` | +6 casos: system tag injeção total, system tag residual, assistant tag, user tag, pré-filtro sem keywords, keyword sem pattern |
| `TestSanitizeOutputC1Bytes` | +5 casos: UTF-8 válido 0x80, 0xA0-0xBF (à), 3-byte U+0900, lone continuation byte descartado |
| `TestProtectedAPIKeyBytes` | Atualizado: Bytes retorna cópia (mutação não afeta original) |
| `TestAPIKeyBytes` | Renomeado para `apiKeyBytes` (não exportado) |

### Melhorias em testes existentes

| Teste | Mudança |
|-------|---------|
| `TestCheckEnvPermissions` → `TestCheckFilePermissions` | Renomeado e expandido: parâmetro path, casos de erro EACCES, permissão de grupo + others combinadas |
| `TestRedactCredentials` | Função exportada como `RedactCredentials` — testada também via hook do cassete |
| `TestNew` (summarizer) | Validação de `HTTPClient` com TLS inseguro, cliente customizado |

---

## 📦 Arquivos Modificados (19 arquivos)

| Arquivo | Mudanças |
|---------|----------|
| `.github/workflows/ci.yml` | 🔄 Release job: `&& github.event_name != 'pull_request'` |
| `CHANGELOG.md` | 📝 Seção PR #18 adicionada |
| `PULL_REQUEST_TEMPLATE.md` | 📝 Este documento |
| `cmd/root.go` | 🔄 `sanitizePrompt()` com pré-filtro de keywords (performance), `injectionKeywords` expandido com system/user/assistant |
| `cmd/root_test.go` | 🧪 `TestExecuteGlobal`, `TestInitRootCommand`, +6 casos em `TestSanitizePrompt`, +5 casos em `TestSanitizeOutputC1Bytes` |
| `go.mod` | ➕ `gopkg.in/dnaeon/go-vcr.v3 v3.2.0`, `gopkg.in/yaml.v3 v3.0.1` |
| `go.sum` | ➕ Checksums para go-vcr e yaml.v3 |
| `internal/config/config.go` | 🔄 `checkEnvPermissions()` → `checkFilePermissions(path string)`, `APIKeyBytes()` → `apiKeyBytes()` (não exportado), warning combinado group+others, stderr log para erros de stat, verificação de `TLDR_API_KEY_FILE` |
| `internal/config/config_test.go` | 🧪 `TestCheckFilePermissions` (8 casos), `TestAPIKeyBytes` → `apiKeyBytes` |
| `internal/integration/cassette_test.go` | ➕ **Novo**: 332 linhas — testes de integração com go-vcr, 7 cenários |
| `internal/integration/testdata/cassettes/summarize_success.yaml` | ➕ Cassete: sucesso 200 (pt-br) |
| `internal/integration/testdata/cassettes/summarize_short.yaml` | ➕ Cassete: sucesso 200 (en, texto curto) |
| `internal/integration/testdata/cassettes/summarize_unauthorized.yaml` | ➕ Cassete: erro 401 |
| `internal/integration/testdata/cassettes/summarize_rate_limited.yaml` | ➕ Cassete: erro 429 |
| `internal/integration/testdata/cassettes/summarize_server_error.yaml` | ➕ Cassete: erro 500 |
| `internal/secrets/secrets.go` | 🔒 `Bytes()` retorna cópia defensiva (`make + copy`), `Clear()` nil-safe mantido |
| `internal/secrets/secrets_test.go` | 🧪 `TestProtectedAPIKeyBytes` atualizado: mutação não afeta original |
| `internal/summarizer/summarizer.go` | ➕ `Config.HTTPClient`, `validateHTTPClientTLS()`, `Client.mu sync.Mutex`, `Client.cleared`, `RedactCredentials()` exportada, `classifyAPIError` reordenado (HTTP antes de substring), `Summarize()` thread-safe com cópia da apiKey |
| `internal/summarizer/summarizer_test.go` | 🧪 Testes para TLS validation, HTTPClient customizado, Clear thread-safe |

---

## ✅ Checklist de Revisão

- [ ] Testes passam (`make test`)
- [ ] Testes com race detector (`make test-race`)
- [ ] Testes de integração em modo replay (`go test -tags=integration -v ./internal/integration/`)
- [ ] Lint passa (`make lint`)
- [ ] Build passa (`make build`)
- [ ] `ProtectedAPIKey.Bytes()` retorna cópia (mutação externa não afeta interno)
- [ ] `sanitizePrompt()` com pré-filtro não quebra prompts sem keywords
- [ ] `sanitizeOutput()` descarta lone continuation bytes (0x80-0xBF sem lead byte)
- [ ] `classifyAPIError` classifica HTTP 401/429/500 antes de fallback por substring
- [ ] `Summarize()` após Clear retorna erro (não panic)
- [ ] Thread-safety: Summarize() e Clear() não raceiam (mutex + cópia defensiva)
- [ ] `validateHTTPClientTLS()` rejeita `InsecureSkipVerify=true`
- [ ] Cassetes gravados têm Authorization redigido (Bearer ***REDACTED***)
- [ ] `checkFilePermissions()` funciona com caminhos arbitrários
- [ ] Erro inesperado de `os.Stat` é logado no stderr
- [ ] Aviso de permissão combinado (group + others) funciona
- [ ] `TLDR_API_KEY_FILE` tem permissões verificadas
- [ ] Release job não roda em pull_request
- [ ] CHANGELOG e PULL_REQUEST_TEMPLATE atualizados

---

## 🔗 Links

- [Issue #6 — Testes unitários e de integração](https://github.com/Elissdev/tl-dr/issues/6)
- [CHANGELOG](./CHANGELOG.md)
- [README](./README.md)
- [Especificação do projeto](./projeto.md)
