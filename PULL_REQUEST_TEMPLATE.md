# Pull Request #15 — Code Review: Fase 4 — Integração com API

> **Branch:** `feat/fase-4-api-integration`
> **Base:** `main`
> **Status:** 🟢 Pronto para merge

---

## 📋 Resumo

Esta PR consolida as **recomendações do code review** sobre a Fase 4 (integração com API OpenAI via Apiário), resultando em melhorias de **segurança**, **robustez**, **performance** e **qualidade de código** no tl;dr. As principais áreas afetadas são: proteção contra prompt injection em prompts customizados, sanitização de bytes C1, suporte a `TLDR_API_KEY_FILE`, limpeza segura de memória, verificação de permissões do `.env`, e tratamento thread-safe do Client.

---

## 🔴 Breaking Changes

### `summarizer.Client.Summarize()` — `panic` substituído por `error`

```go
// Antes
s.Summarize(ctx, prompt, text) // panic se chamado após Clear()

// Depois
s.Summarize(ctx, prompt, text) // retorna error se chamado após Clear()
```

**Ação:** Se seu código capturava o panic com `recover`, agora deve tratar o erro normalmente.

### `summarizer.Client.apiKey` — tipo alterado de `string` para `[]byte`

- Campo não exportado (afeta apenas código interno)
- Função `redactCredentials` teve assinatura atualizada: `(s string, apiKey []byte)`
- Função `classifyAPIError` agora recebe `apiKey []byte` como parâmetro

### `sanitizePrompt()` — flag `--prompt` agora sanitizada

- **Antes:** Prompt customizado passava diretamente para a API
- **Depois:** Padrões de injeção são substituídos por `[REMOVED]`; prompt 100% injetivo causa erro
- **Mitigação:** Regex calibrada; `"now"` não é mais gatilho isolado (evita falso positivo)

### `config.Load()` — novo side effect em stderr

- **Antes:** `Load()` não escrevia em stderr
- **Depois:** Se `.env` tiver permissões >0600, emite warning de segurança no stderr
- **Mitigação:** Use `chmod 600 .env` para silenciar

### `secrets.LoadAPIKey()` — mensagem de erro alterada

- **Antes:** `"TLDR_API_KEY não definida"`
- **Depois:** `"TLDR_API_KEY não definida (defina a variável ou TLDR_API_KEY_FILE)"`

### `config.Clear()` — semântica reforçada

- **Antes:** Usar struct após Clear() era possível (APIKey retornava "")
- **Depois:** Contrato explicita que struct não deve mais ser usada após Clear()

---

## 🟢 Novas Funcionalidades

### Segurança

| Funcionalidade | Descrição |
|----------------|-----------|
| 🔑 **`TLDR_API_KEY_FILE`** | Ler chave de API de arquivo (fallback quando `TLDR_API_KEY` não definida) |
| 🧪 **`secrets.ProtectedAPIKey.Bytes()`** | Retorna cópia defensiva da chave como `[]byte` (mutação não afeta original) |
| 🔍 **`config.Config.APIKeyBytes()`** | Acessa chave como `[]byte` via Config para gerenciamento de memória |
| 🧹 **`summarizer.Client.Clear()`** | Zera a chave de API da memória do Client (thread-safe com mutex) |
| 🛡️ **`sanitizePrompt()`** | Sanitização de prompts customizados contra injeção com pré-filtro de keywords |
| 🌐 **Bytes C1 (0x80-0x9F)** | Tratamento de controles C1 de 8-bit em `sanitizeOutput()` com estado UTF-8 |
| ⚠️ **`checkEnvPermissions()`** | Aviso de segurança se `.env` tiver permissões muito permissivas |

### Performance

| Funcionalidade | Descrição |
|----------------|-----------|
| ⚡ **Pré-filtro de keywords** | `sanitizePrompt()` verifica keywords primeiro; se ausentes, pula scan com regexes |
| 🧩 **`injectionKeywords`** | Lista de 15 palavras-chave para fast-path: `ignore`, `reveal`, `show`, `system`, `user`, `assistant`, etc. |

### Engine

| Funcionalidade | Descrição |
|----------------|-----------|
| 🔒 **`Client.mu sync.Mutex`** | Proteção thread-safe contra race condition entre `Summarize()` e `Clear()` |
| 📋 **`Client.cleared`** | Flag booleana para detectar uso após limpeza (em vez de panic) |
| 🧪 **`allRemoved` regex** | Detecta se prompt contém apenas marcações `[REMOVED]` (bloqueio total) |
| 🧪 **`spaceRun` regex** | Colapsa espaços/tabs duplicados (preserva newlines) no resultado sanitizado |

---

## 🔧 Melhorias Técnicas

### Correções de segurança/qualidade

| Issue | Solução |
|-------|---------|
| **Prompt injection via `--prompt`** | `sanitizePrompt()` com 4 regexes de injection + pré-filtro de keywords |
| **Vazamento de bytes C1 na saída** | `sanitizeOutput()` rastreia estado UTF-8 (`utf8Remaining`) para distinguir continuation bytes legítimos de controles C1 |
| **Chave de API no arquivo** | `LoadAPIKey()` lê de `TLDR_API_KEY_FILE` (com `filepath.Clean` para G304), lê arquivo, remove trailing newline/CR |
| **Permissões do .env** | `checkEnvPermissions()` verifica bits `0o040` (group) e `0o004` (others); recomenda `chmod 600` |
| **`Summarize()` pós-Clear** | Retorna erro em vez de panic — thread-safe com mutex |
| **`cfg.Clear()` imediato** | `cfg.Clear()` agora é chamado imediatamente (sem defer); `s.Clear()` mantém `defer` |
| **`redactCredentials` com `[]byte`** | Tipo alterado de `string` para `[]byte` — permite gerenciamento de memória |
| **`classifyAPIError` recebe `[]byte`** | Cópia defensiva da chave antes do unlock para evitar race |

### Falsos positivos eliminados

| Palavra | Motivo | Ação |
|---------|--------|------|
| `now` | Aparece em prompts legítimos ("Resuma agora", "Agora explique") | Removido do `injectionPatterns` |
| `system` | "O sistema está rodando" é legítimo | Mantido como keyword, mas sem pattern que case sozinho |

### Padrões de código

| Melhoria | Detalhe |
|----------|---------|
| **`sanitizePrompt()` separada** | Função isolada e testável (18 casos de teste) |
| **`sanitizeOutput()` reescrita** | Código reestruturado com estado UTF-8, tratamento de C1 8-bit |
| **`ProtectedAPIKey.Bytes()` cópia defensiva** | `make([]byte, len)` + `copy()` — mutação externa não afeta interno |
| **`Config.APIKeyBytes()` nil-safe** | Retorna nil se `protectedKey` for nil |
| **`CheckEnvPermissions` testável** | `captureStderr()` helper para verificar warnings |

---

## 🧪 Testes

### Testes de unidade adicionados/modificados (+261 linhas)

| Teste | O que cobre |
|-------|-------------|
| `TestSanitizePrompt` | 18 casos: normal, vazio, injeção total/parcial, `im_start`, `system`, `user`, `assistant`, pré-filtro com/sem keyword |
| `TestSanitizeOutputC1Bytes` | 15 casos: CSI 8-bit (0x9B), OSC (0x9D), DCS (0x90), SOS (0x98), PM (0x9E), APC (0x9F), ST (0x9C), C1 genérico, C1 + ESC misturados, UTF-8 válido (0x80, 0xA0-0xBF, 3-byte), lone continuation byte |
| `TestAPIKeyBytes` | Slice correto, nil após Clear |
| `TestProtectedAPIKeyBytes` | Retorna cópia (mutação não afeta original), nil após Clear |
| `TestCheckEnvPermissions` | Sem .env (sem warning), 0600 (sem warning), 0644 (com warning) |
| `TestClientClear` | Clear zera apiKey e marca `cleared`, Summarize após Clear retorna erro, double Clear seguro |
| `TestRedactCredentials` | Atualizado: `[]byte`, slice vazio não quebra |

### Melhorias em testes existentes

| Teste | Mudança |
|-------|---------|
| `TestClassifyAPIErrorSanitization` | Agora passa `s.apiKey` como `[]byte` |
| `TestRedactCredentials` | Assinatura atualizada para `[]byte` |
| `TestNew` (summarizer) | Comparação de `s.apiKey` usa `string(s.apiKey)` |

---

## 📦 Arquivos Modificados (10 arquivos)

| Arquivo | Mudanças |
|---------|----------|
| `.env.example` | ➕ Documentação de `TLDR_API_KEY_FILE` |
| `CHANGELOG.md` | 📝 Seção PR #15 detalhada |
| `PULL_REQUEST_TEMPLATE.md` | 📝 Este documento |
| `cmd/root.go` | 🔄 `sanitizePrompt()` integrado ao fluxo, `cfg.Clear()` imediato + `defer s.Clear()`, comentários atualizados |
| `cmd/root_test.go` | 🧪 +231 linhas: `TestSanitizePrompt`, `TestSanitizeOutputC1Bytes` |
| `internal/config/config.go` | ➕ `checkEnvPermissions()`, `APIKeyBytes()`, warning no stderr |
| `internal/config/config_test.go` | 🧪 +156 linhas: `TestAPIKeyBytes`, `TestCheckEnvPermissions` |
| `internal/secrets/secrets.go` | ➕ `LoadAPIKey()` lê `TLDR_API_KEY_FILE`, `Bytes()` método, `filepath.Clean` |
| `internal/secrets/secrets_test.go` | 🧪 `TestProtectedAPIKeyBytes` |
| `internal/summarizer/summarizer.go` | 🔄 `apiKey` → `[]byte`, mutex thread-safe, `Clear()`, `classifyAPIError` recebe `[]byte` |
| `internal/summarizer/summarizer_test.go` | 🧪 `TestClientClear`, assinaturas atualizadas |

---

## ✅ Checklist de Revisão

- [ ] Testes passam (`make test`)
- [ ] Testes com race detector (`make test-race`)
- [ ] Lint passa (`make lint`)
- [ ] Build passa (`make build`)
- [ ] `TLDR_API_KEY_FILE` funciona (arquivo com permissão restritiva)
- [ ] `ProtectedAPIKey.Bytes()` retorna cópia (mutação externa não afeta interno)
- [ ] `Config.APIKeyBytes()` retorna nil após Clear
- [ ] `sanitizePrompt()` bloqueia prompt 100% injetivo com erro
- [ ] `sanitizePrompt()` preserva prompts normais (sem falsos positivos)
- [ ] `sanitizeOutput()` trata bytes C1 8-bit (0x80-0x9F) corretamente
- [ ] `sanitizeOutput()` preserva UTF-8 válido com continuation bytes
- [ ] `Summarize()` após Clear retorna erro (não panic)
- [ ] `Clear()` duplo não causa panic
- [ ] Race condition entre Summarize() e Clear() protegida por mutex
- [ ] `cfg.Clear()` chamado imediatamente (não defer) em root.go
- [ ] `s.Clear()` chamado via defer em root.go
- [ ] `.env` com permissão 0644 emite warning (0600 não)
- [ ] `checkEnvPermissions()` não quebra se .env não existir
- [ ] CHANGELOG e PULL_REQUEST_TEMPLATE atualizados

---

## 🔗 Links

- [CHANGELOG](./CHANGELOG.md)
- [README](./README.md)
- [Especificação do projeto](./projeto.md)
