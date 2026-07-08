# Pull Request #19 — Fase 7: CI/CD com GitHub Actions + Correções de Code Review

> **Branch:** `feat/fase-7-cicd`
> **Base:** `main`
> **Status:** 🟢 Pronto para merge

---

## 📋 Resumo

Esta PR implementa a **Fase 7** do projeto tl;dr, adicionando **CI/CD completo com GitHub Actions** (4 jobs independentes: test, lint, build, release), além de melhorias de código identificadas no code review. As principais áreas afetadas são: pipeline de CI desacoplado de branch específica, cross-compilação, release automática via tags, substituição de números mágicos por constantes nomeadas em `sanitizeOutput()`, case-insensitive locale lookup, validação de `--timeout` com fail fast, tratamento explícito de `finish_reason` via switch, e zeroing da cópia da API key após uso.

---

## 🟢 Novas Funcionalidades

### CI/CD com GitHub Actions

| Funcionalidade | Descrição |
|----------------|-----------|
| 🧪 **Job `test`** | `go test ./... -v -count=1` + `go test ./... -race -count=1` + gosec security scan |
| 🔍 **Job `lint`** | `golangci-lint` com timeout de 5 minutos |
| 📦 **Job `build`** | Cross-compilação `GOOS=linux GOARCH=amd64` + upload de artifact (`tldr-linux-amd64`) |
| 🚀 **Job `release`** | Disparado apenas em tags `v*`, com `github.event_name != 'pull_request'` para evitar release acidental em PRs |
| 🔄 **CI em todas as branches** | Triggers `push` e `pull_request` sem restrição de branch (antes era restrito a `main`) |

### Constantes nomeadas para UTF-8 e ANSI

- **`cmd/root.go`**: Números mágicos substituídos por constantes nomeadas em `sanitizeOutput()`:
  - **UTF-8 lead bytes**: `utf8Lead2`/`utf8Lead2End` (0xC0-0xDF), `utf8Lead3`/`utf8Lead3End` (0xE0-0xEF), `utf8Lead4`/`utf8Lead4End` (0xF0-0xF7)
  - **Continuation bytes**: `utf8ContStart`/`utf8ContEnd` (0x80-0xBF)
  - **C1 control characters**: `c1Start`/`c1End` (0x80-0x9F)
  - **ANSI 8-bit sequences**: `csi8bit` (0x9B), `osc8bit` (0x9D), `dcs8bit` (0x90), `sos8bit` (0x98), `pm8bit` (0x9E), `apc8bit` (0x9F), `st8bit` (0x9C), `sci8bit` (0x9A)
  - **Controles diversos**: `bel` (0x07), `esc` (0x1B), `c0Min`/`c0Max` (0x00-0x1F), `csiFinalMin`/`csiFinalMax` (0x40-0x7E)

### Case-insensitive locale lookup

- **`getLocale()`** agora normaliza o idioma com `strings.ToLower()` + `strings.ReplaceAll("_", "-")`, aceitando `"PT-BR"`, `"pt_BR"`, `"Pt-br"` como equivalentes a `"pt-br"`

---

## 🔧 Melhorias Técnicas

### Qualidade de código

| Melhoria | Detalhe |
|----------|---------|
| **Números mágicos eliminados** | 20+ constantes nomeadas para faixas UTF-8 e ANSI C1/C0 |
| **`finish_reason` com switch** | Tratamento explícito de `stop`, `length`, `content_filter` + fallback para valores desconhecidos (`tool_calls`, `function_call`) com mensagens descritivas |
| **`apiKeyCopy` zerada após uso** | `defer` com loop de zeroing em `Summarize()` minimiza janela de exposição da chave na memória |
| **Variáveis de ambiente preservadas** | Testes que setam `TLDR_API_KEY` restauram o valor original via `defer` |

### Validação de entrada

| Melhoria | Detalhe |
|----------|---------|
| **Fail fast para `--timeout < 0`** | Validação movida para **antes** de `ReadInput` — feedback instantâneo sem esperar I/O |
| **Mensagem de erro clara** | Retorna `ExitArgs` com `"--timeout deve ser um número positivo em segundos, got %d"` |
| **Help do `--timeout`** | Descrição atualizada para `"deve ser > 0; default: 30"` |

### CI/CD

| Melhoria | Detalhe |
|----------|---------|
| **Trigger desacoplado** | `push` e `pull_request` sem `branches: [main]` — CI roda em qualquer branch |
| **Cross-compilação explícita** | `GOOS=linux GOARCH=amd64` garante binário consistente independente do runner |
| **Release condicional** | `github.event_name != 'pull_request'` impede release acidental em PRs |
| **Artifact naming** | `tldr-linux-amd64` — nome descritivo com platforma |

---

## 🧪 Testes

### Novos testes unitários

| Teste | O que cobre |
|-------|-------------|
| `TestGetLocaleCaseInsensitive` | 6 subtestes: `PT-BR`, `pt_BR`, `Pt-br`, `PT`, equivalência entre 5 variantes, `EN` maiúsculo |
| `TestExecute/--timeout_negativo_(fail_fast_antes_de_IO)` | Timeout -5 falha com erro contendo "positivo" e retorna `ExitArgs` (exit code 3) |

---

## 📦 Arquivos Modificados (5 arquivos)

| Arquivo | Status | Mudanças |
|---------|--------|----------|
| `.github/workflows/ci.yml` | 🔄 | Removido `branches: [main]` dos triggers `push` e `pull_request`; build agora usa `GOOS=linux GOARCH=amd64` |
| `CHANGELOG.md` | 📝 | Seção PR #19 adicionada |
| `PULL_REQUEST_TEMPLATE.md` | 📝 | Este documento |
| `cmd/root.go` | 🔄 | Constantes nomeadas UTF-8/ANSI; validação `--timeout < 0` com fail fast; `getLocale()` case-insensitive; help do timeout atualizado |
| `cmd/root_test.go` | 🧪 | `TestGetLocaleCaseInsensitive` (6 subtestes); `TestExecute/--timeout negativo` |
| `internal/summarizer/summarizer.go` | 🔄 | `finish_reason` com switch; `apiKeyCopy` zerada via defer |

---

## ✅ Checklist de Revisão

- [ ] Testes passam (`make test`)
- [ ] Testes com race detector (`make test-race`)
- [ ] Lint passa (`make lint`)
- [ ] Build passa (`make build`)
- [ ] CI roda em branches não-main (push e pull_request sem restrição)
- [ ] `--timeout -5` retorna erro fail fast (antes de ReadInput)
- [ ] `--timeout -5` retorna ExitArgs (código 3)
- [ ] `getLocale("PT-BR")` retorna configuração pt-br (case-insensitive)
- [ ] `getLocale("pt_BR")` retorna configuração pt-br (underscore)
- [ ] Números mágicos substituídos por constantes em `sanitizeOutput()`
- [ ] `finish_reason` desconhecido (ex: `tool_calls`) tem fallback seguro
- [ ] `apiKeyCopy` é zerada após uso em Summarize()
- [ ] Release job não roda em pull_request
- [ ] Release job roda apenas em tags `v*`
- [ ] Cross-compilação produz binário `tldr` (linux amd64)
- [ ] Artifact `tldr-linux-amd64` é carregado no build job
- [ ] CHANGELOG e PULL_REQUEST_TEMPLATE atualizados

---

## 🔗 Links

- [Issue #7 — CI/CD com GitHub Actions](https://github.com/Elissdev/tl-dr/issues/7)
- [CHANGELOG](./CHANGELOG.md)
- [README](./README.md)
- [Especificação do projeto](./projeto.md)
