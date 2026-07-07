# Pull Request #14 — Conclusão das 11 Code Reviews

> **Branch:** `feat/fase-3-input-refactor`
> **Base:** `main`
> **Status:** 🟢 Pronto para merge

---

## 📋 Resumo

Esta PR consolida **11 rodadas de code review** sobre a Fase 3, resultando em melhorias de **segurança**, **robustez**, **testabilidade** e **qualidade de código** no tl;dr. As principais áreas afetadas são: proteção contra prompt injection, sanitização de saída, redação estendida de credenciais, timeout configurável em stdin, eliminação de TOCTOU, e refatoração do comando raiz para testabilidade isolada.

---

## 🔴 Breaking Changes na API Pública

### `summarizer.New()` agora retorna `(*Client, error)`

```go
// Antes
s := summarizer.New(cfg)    // *Summarizer (nunca nil)

// Depois
s, err := summarizer.New(cfg) // (*Client, error)
if err != nil { /* APIKey/Model/BaseURL vazio */ }
```

### Struct `Summarizer` renomeada para `Client`

Qualquer referência direta a `summarizer.Summarizer` quebra.

### Constantes de exit code renomeadas

| Antes | Depois |
|-------|--------|
| `ExitSuccess` | `ExitOK` |
| `ExitGenericError` | `ExitInternal` |
| `ExitAPIError` | `ExitAPI` |
| `ExitArgumentError` | `ExitArgs` |
| *(nova)* | `ExitTimeout = 4` |

### Renomeações no pacote `input`

| Antes | Depois |
|-------|--------|
| `input.ReadFile()` | `input.ReadFromFile()` |
| `input.ReadStdin()` | `input.ReadFromStdin()` |
| `input.IsStdinAvailable()` | `input.IsStdinRedirected()` |

---

## 🟠 Breaking Changes Comportamentais

- **`ReadFromStdin()` rejeita stdin vazio**: antes retornava `("", nil)`, agora retorna erro
- **`ReadFromStdin()` rejeita terminal interativo**: se stdin for um terminal sem pipe/redirect, erro imediato
- **Timeout de 30s no stdin**: se o pipe não enviar dados em 30s, a leitura falha com erro de timeout
- **`buildPrompt()` agora inclui prefixo de segurança** imutável contra prompt injection, antes de qualquer prompt customizado ou padrão
- **Saída sanitizada por padrão**: ANSI escape codes (CSI, OSC, DCS, SOS, PM, APC) são removidos automaticamente. Use `--no-sanitize` para desabilitar
- **`cfg.Clear()` movido para `defer`**: a chave de API é zerada apenas no retorno do comando
- **Modelo default hardcoded**: se `TLDR_DEFAULT_MODEL` não for definido, usa `deepseek/deepseek-v4-flash`

---

## 🟢 Novas Funcionalidades

### Segurança

| Funcionalidade | Descrição |
|----------------|-----------|
| 🔒 **Safety prefix anti-prompt injection** | Prefixo imutável em pt (`SafetyPrefixPT`) e en (`SafetyPrefixEN`) inserido antes de todo prompt |
| 🧹 **`sanitizeOutput()`** | Remove CSI, OSC, DCS, SOS, PM, APC sequences e caracteres de controle (exceto `\t`, `\n`, `\r`) |
| 🚫 **Flag `--no-sanitize`** | Desabilita a sanitização de escape codes (útil se o terminal já processa cores) |
| 🕵️ **Redação estendida de credenciais** | Agora cobre: OpenAI (`sk-`, `sk-proj-`), DeepSeek, Anthropic (`sk-ant-`), GitHub PAT (`ghp_`), JWT, `api_key=`, `token=`, fallback genérico 60+ chars |
| 🔑 **Redação da própria chave** | A chave da API configurada no `Client` também é redigida em mensagens de erro |
| 🔗 **`redactedError` preserva cadeia** | Erros redigidos preservam o erro original via `Unwrap()` para `errors.Is`/`errors.As` |

### CLI

| Funcionalidade | Descrição |
|----------------|-----------|
| 🏷️ **Flag `--version` / `-v`** | Exibe versão do binário (injetada via ldflags no `Makefile`) |
| ⏱️ **Flag `--timeout` / `-t`** | Timeout customizável via CLI (sobrescreve `TLDR_TIMEOUT`) |
| 🌐 **Feedback visual no stderr** | Exibe idioma e modelo usados, além de "📝 Resumindo..." |
| ⚠️ **Truncamento parcial** | Se a resposta for truncada (`finish_reason=length`), exibe o conteúdo parcial + aviso no stderr |

### Engine

| Funcionalidade | Descrição |
|----------------|-----------|
| 📦 **`ErrTruncated`** | Sentinel error para respostas truncadas — retorna conteúdo parcial + erro |
| ⏰ **`ErrTimeout`** | Sentinel error para timeout — detectável via `errors.Is(err, summarizer.ErrTimeout)` |
| 🗺️ **`getLocale()`** | Sistema de localização com `localeConfig` (SafetyPrefix + DefaultPrompt + RespondIn) |
| ➕ **`firstNonEmpty()`** | Helper para resolver precedência: flag > env > default |
| 🧪 **`newRootCommand()`** | Comando raiz agora é construído por função (sem estado global/`init()`), permitindo testes isolados |
| 🧪 **`ReadFromStdinWithTimeout()`** | Versão testável da leitura de stdin com timeout customizável |

---

## 🔧 Melhorias Técnicas

### Correções de segurança/qualidade

| Issue | Solução |
|-------|---------|
| **TOCTOU em `ReadFromFile`** | `os.Open()` primeiro, `f.Stat()` depois, `io.ReadAll(limitReader)` — eliminado race entre Stat e ReadFile |
| **TOCTOU em `ReadFromStdin`** | Verificação de terminal movida para dentro da função (antes era externa) |
| **Goroutine leak em timeout** | Leitura de stdin em goroutine com `context.WithTimeout` + canal com buffer 1 |
| **`unsafe` removido de `secrets`** | Uso de `unsafe.Pointer` para zerar buffer de string não é confiável em Go moderno; agora apenas copia para `[]byte` controlado |
| **Validação de URL** | Agora exige `http://` ou `https://` no scheme (antes aceitava qualquer URI) |
| **`summarizer.New()` valida campos** | `APIKey`, `Model` e `BaseURL` vazios retornam erro (antes aceitavam silenciosamente) |
| **Content vazio com `finish_reason=stop`** | Novo caso de erro detectado e reportado |

### Padrões de código

| Melhoria | Detalhe |
|----------|---------|
| **Sem estado global/`init()`** | `newRootCommand()` retorna um comando novo a cada chamada, flags via ponteiros |
| **Variáveis globais eliminadas** | `lang`, `model`, `customPromptFlag` removidas — agora são ponteiros locais do `cobra.Command` |
| **`getEnv` renomeado para `envOr`** | Nome mais semântico |
| **`SupportedLocales` como mapa** | Adicionar novo idioma = adicionar entrada no mapa |
| **`apiKeyRedactors` como slice de regexes** | Em vez de uma regex gigante, lista de patterns específicos e documentados |
| **`newTestClient()` helper** | Elimina repetição de `summarizer.New` + `t.Fatalf` nos testes |

---

## 🧪 Testes

### Testes de unidade adicionados

| Teste | O que cobre |
|-------|-------------|
| `TestBuildPrompt` | Prefixo de segurança + sufixo correto para en, pt, pt-br, es |
| `TestBuildPromptSafetyPrefixAlwaysPresent` | Todas as combinações de idioma sempre incluem safety prefix |
| `TestSanitizeOutput` | CSI, OSC, DCS, SOS, PM, APC, newlines, tabs, CR, string vazia |
| `TestSanitizeOutputEdgeCases` | ESC isolado, CSI/OSC incompleto, ESC + controle, múltiplos ESC, unicode |
| `TestGetLocale` | pt-br, pt, en, idioma desconhecido, não quebra existentes |
| `TestFirstNonEmpty` | Vários casos: preenchido, vazios, slice vazio |
| `TestExecute` | langPattern válido/inválido, `--no-sanitize`, sem API key, env vars, idioma inválido, arquivo inexistente |
| `TestNew` (summarizer) | Config válida, API key vazia, modelo vazio, base URL vazia, timeout zero |
| `TestSummarize` | `finish_reason=stop` vazio, erro 400 |
| `TestClassifyAPIErrorSanitization` | Chave sk-, sk-proj-, api_key=, token=, própria chave, timeout, `errors.Is(ErrTimeout)` |
| `TestRedactCredentials` | Redige chave configurada, apiKey vazia, string vazia |
| Stdin timeout | `ReadFromStdinWithTimeout` com pipe lento |
| `TestIsStdinRedirected` | Pipe retorna true (determinístico) |

---

## 📦 Arquivos Modificados (18 arquivos)

| Arquivo | Mudanças |
|---------|----------|
| `.github/workflows/ci.yml` | ➕ Race detector + gosec security scan |
| `CHANGELOG.md` | 📝 Documentação das mudanças |
| `Makefile` | 🔧 Versão via ldflags |
| `PULL_REQUEST_TEMPLATE.md` | 📝 Este documento |
| `README.md` | 📝 Flags `--no-sanitize`, `--timeout`, `--version`; exit codes; env vars |
| `cmd/errors.go` | 🔴 Exit codes renomeados + `ExitTimeout` |
| `cmd/errors_test.go` | 🧪 Atualizado para novos exit codes |
| `cmd/root.go` | 🔄 Refatoração completa (84% rewrite) |
| `cmd/root_test.go` | 🧪 +436 linhas de novos testes |
| `go.mod` | ➕ godotenv como dependência direta |
| `internal/config/config.go` | 🔧 `envOr()`, validação de scheme http/https |
| `internal/input/input.go` | 🔧 TOCTOU, timeout, ~ expansion, `ReadFromStdinWithTimeout` |
| `internal/input/input_test.go` | 🧪 Stdin timeout, `IsStdinRedirected` |
| `internal/integration/summarizer_test.go` | 🧪 `//go:build integration`, `summarizer.New()` error |
| `internal/secrets/secrets.go` | 🔒 `unsafe` removido |
| `internal/summarizer/summarizer.go` | 🔴 `Client` + `(*Client, error)`, `ErrTruncated`, `ErrTimeout`, redação |
| `internal/summarizer/summarizer_test.go` | 🧪 `newTestClient`, novos testes |
| `.gitignore` | 🔧 Ignorar `teste.txt`, `me explique...` |

---

## ✅ Checklist de Revisão

- [ ] Testes passam (`make test`)
- [ ] Testes com race detector (`make test-race`)
- [ ] Lint passa (`make lint`)
- [ ] Build passa (`make build`)
- [ ] `summarizer.New()` trata erro de retorno em todos os callers
- [ ] Prefixo de segurança contra prompt injection está presente em todos os prompts
- [ ] Saída é sanitizada por padrão (ANSI removido)
- [ ] `--no-sanitize` desabilita a sanitização
- [ ] Timeout no stdin funciona (pipe travado não trava o CLI)
- [ ] Redação de credenciais cobre todos os formatos conhecidos
- [ ] `errors.Is(err, summarizer.ErrTimeout)` funciona
- [ ] `errors.Is(err, summarizer.ErrTruncated)` com conteúdo parcial
- [ ] TOCTOU eliminado em `ReadFromFile` e `ReadFromStdin`
- [ ] `cfg.Clear()` via `defer` (não antes)
- [ ] `--version` exibe a versão correta
- [ ] `--timeout` sobrescreve `TLDR_TIMEOUT`
- [ ] CHANGELOG e PULL_REQUEST_TEMPLATE atualizados

---

## 🔗 Links

- [CHANGELOG](./CHANGELOG.md)
- [README](./README.md)
- [Especificação do projeto](./projeto.md)
