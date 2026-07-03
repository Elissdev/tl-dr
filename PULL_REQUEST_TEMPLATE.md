# Pull Request — Fase 2: Configuração via Variáveis de Ambiente

> **Branch:** `feat/fase-2-config`
> **Base:** `main`
> **Status:** 🟡 Em revisão

---

## 📋 Resumo

Esta PR conclui a **Fase 2** do projeto tl;dr, trazendo melhorias na configuração via variáveis de ambiente, validação mais robusta, suporte a arquivo `.env`, suporte a prompts em português, e diversas melhorias de segurança e qualidade.

---

## 🚀 Novas Funcionalidades

### 🌐 Suporte a arquivo `.env`
- Adicionada dependência `github.com/joho/godotenv v1.5.1`
- `config.Load()` tenta carregar automaticamente um arquivo `.env` na raiz do projeto
- Se o arquivo não existir, ignora silenciosamente (compatibilidade retroativa)
- Se existir mas tiver erro de parsing, retorna erro explicitamente

### 🇧🇷 Prompt padrão em português
- Quando o idioma é `pt-br` ou `pt`, o prompt padrão usa template em português:
  - **Antes:** `"Summarize the following text in pt-br. Be concise but capture all key points."`
  - **Agora:** `"Resuma o texto a seguir em pt-br. Seja conciso mas capture todos os pontos-chave."`
- Demais idiomas continuam usando o template em inglês

---

## 🔄 Mudanças na API Interna

### `config.Load()` agora retorna `(Config, error)` — **BREAKING CHANGE**
```go
// Antes (Fase 1)
cfg := config.Load()

// Depois (Fase 2)
cfg, err := config.Load()
if err != nil {
    // tratar erro
}
```

### Método `Validate()` removido
- A validação agora é feita durante o `Load()`, eliminando o passo extra
- Código cliente: `config.Load()` → `cfg.Clear()` (sem `cfg.Validate()` intermediário)

---

## ✨ Novas Validações

| Variável | Validação | Comportamento Anterior |
|----------|-----------|----------------------|
| `TLDR_BASE_URL` | Validada como URL via `url.ParseRequestURI()` | Aceita qualquer valor |
| `TLDR_TIMEOUT` | Valor não numérico agora retorna erro | Ignorado silenciosamente (usava padrão) |
| Chave de API | Mensagem de erro mais genérica | Culpava exclusivamente `TLDR_API_KEY` |

---

## 🔒 Segurança

- **`ProtectedAPIKey.Clear()` nil-safe**: chamar `Clear()` em ponteiro `nil` não causa pânico
- **Double-clear seguro**: chamar `Clear()` duas vezes consecutivas não causa pânico
- **Teste byte a byte**: verificação de que todos os bytes do slice interno são zerados
- **Documentação da sanitização**: regex de redação de chaves tem escopo explicitamente documentado (aplicada **apenas** em mensagens de erro, nunca no conteúdo do usuário)
- **Ordem do `cfg.Clear()` documentada**: comentário de segurança em `root.go` alertando que `Clear()` deve permanecer após a cópia da chave para `summarizer.Config`

---

## 🧪 Testes

### Novos testes adicionados
- `TestBuildPrompt` — casos `pt-br` e `pt` com template em português
- `TestLoad` — URL base inválida (`://invalida`) retorna erro
- `TestLoad` — `TLDR_TIMEOUT` inválido retorna erro (antes usava default silenciosamente)
- `TestLoad` — sem API key agora verifica campos do `Config` retornado (APIKey vazia, BaseURL default, DefaultModel default)
- `TestClear` — double-clear não causa pânico
- `TestIsStdinRedirected` — simula pipe real com `os.Pipe()` (true) e terminal (false, sem panic)

### Testes melhorados
- `TestProtectedAPIKeyClear` — verificação byte a byte do buffer interno

---

## 📝 Documentação

- Comentários em `NewExitError`/`WrapExitError` com exemplos de uso:
  ```go
  // Exemplo: return NewExitError(ExitArgs, "idioma é obrigatório")
  // Exemplo: return WrapExitError(ExitInternal, err)
  ```
- Comentário de segurança sobre a ordem do `cfg.Clear()` em `root.go`
- Documentação da dupla validação de timeout em `summarizer.New()` (redundância segura)
- Comentário de fallback seguro em `main.go` para casos de pânico recuperado

---

## 🔧 Manutenção

- Função auxiliar `cfgZeroed()` removida (substituída por inicialização inline)
- `go.mod`/`go.sum` atualizados com dependência `joho/godotenv`

---

## 📦 Arquivos Modificados (13 arquivos)

| Arquivo | Tipo de Mudança |
|---------|----------------|
| `cmd/errors.go` | 📝 Documentação |
| `cmd/root.go` | 🔄 Refatoração + 🇧🇷 Prompts pt |
| `cmd/root_test.go` | 🧪 Novos testes |
| `internal/config/config.go` | ⚡ Validações + .env + error |
| `internal/config/config_test.go` | 🧪 Novos testes |
| `internal/input/input_test.go` | 🧪 Teste de pipe real |
| `internal/secrets/secrets.go` | 🔒 Nil-safe Clear |
| `internal/secrets/secrets_test.go` | 🧪 Byte-level + double-clear |
| `internal/summarizer/summarizer.go` | 📝 Documentação |
| `main.go` | 📝 Documentação fallback |
| `go.mod` | ➕ godotenv |
| `go.sum` | ➕ Checksums |
| `CHANGELOG.md` | 📝 Registro de versão |

---

## ✅ Checklist de Revisão

- [ ] Testes passam (`make test`)
- [ ] Testes com race detector (`make test-race`)
- [ ] Testes de integração (opcional, `make test-integration`)
- [ ] Lint passa (`make lint`)
- [ ] Build passa (`make build`)
- [ ] `config.Load()` agora retorna erro e o chamador trata adequadamente
- [ ] `Validate()` foi removido de todos os callers
- [ ] `.env` é carregado sem quebrar ambientes sem o arquivo
- [ ] Chave de API é limpa da memória (`cfg.Clear()`) após uso
- [ ] Prompt em português funciona para `pt-br` e `pt`
- [ ] `TLDR_BASE_URL` inválida retorna erro
- [ ] `TLDR_TIMEOUT` inválido retorna erro
- [ ] CHANGELOG atualizado

---

## 🔗 Links

- [Especificação do projeto](./projeto.md)
- [CHANGELOG](./CHANGELOG.md)
- [README](./README.md)

---

> **Nota:** Esta PR contém tanto o diff estrutural da Fase 2 (validations, `Load() error`) quanto melhorias incrementais (`.env`, prompts pt, nil-safe, testes).
