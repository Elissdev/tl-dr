# Changelog

## Working Tree — Breaking Changes (não publicado)

### 🔴 Breaking Changes na API Pública

- **`summarizer.New()` agora retorna `(*Client, error)`**: a função agora valida campos
  obrigatórios (`APIKey`, `Model`, `BaseURL`) e retorna erro se ausentes. O retorno mudou
  de `*Summarizer` para `(*Client, error)`.
- **Struct `Summarizer` renomeada para `Client`**: qualquer referência direta a
  `summarizer.Summarizer` quebra.
- **Constantes de exit code renomeadas**: `ExitSuccess` → `ExitOK`, `ExitGenericError` →
  `ExitInternal`, `ExitAPIError` → `ExitAPI`, `ExitArgumentError` → `ExitArgs`.
  Nova constante: `ExitTimeout = 4`.
- **`input.ReadFile()` renomeado para `input.ReadFromFile()`**
- **`input.ReadStdin()` renomeado para `input.ReadFromStdin()`**
- **`input.IsStdinAvailable()` renomeado para `input.IsStdinRedirected()`**

### 🟠 Breaking Changes Comportamentais

- **`ReadFromStdin()` agora rejeita stdin vazio com erro**: antes retornava `("", nil)`,
  agora retorna erro `"nenhum texto recebido via stdin"`.
- **`ReadFromStdin()` agora rejeita terminal interativo**: se stdin for um terminal
  (sem pipe/redirecionamento), retorna erro imediato.
- **Timeout de 30s no stdin**: se o pipe não enviar dados em 30s, a leitura falha com
  erro de timeout e o stdin é fechado.
- **`buildPrompt()` mudou formato de saída**: prefixo de segurança contra prompt injection
  adicionado antes de todo prompt customizado ou padrão.
- **Saída agora é sanitizada**: ANSI escape codes são removidos da resposta do modelo.
  Use a nova flag `--no-sanitize` para desabilitar.
- **`cfg.Clear()` movido para `defer`**: a chave de API agora é zerada apenas no retorno
  do comando, não imediatamente após copiar para o client.
- **Modelo default hardcoded**: se `TLDR_DEFAULT_MODEL` não for definido, usa
  `deepseek/deepseek-v4-flash` como fallback.

### 🟢 Novas Funcionalidades

- **Flag `--no-sanitize`**: desabilita a remoção de ANSI escape codes da saída.
- **Validação de idioma**: formato do `--lang` é validado (ex: `pt-br`, `en`, `zh-CN`).
- **Redação estendida de credenciais**: cobre DeepSeek, Anthropic, GitHub PAT, JWT,
  e fallback genérico para strings de 40+ caracteres.
- **Erros preservam cadeia original**: `classifyAPIError` agora retorna `*redactedError`
  que preserva a cadeia de erros para `errors.Is`/`errors.As`.

### 🔧 Melhorias Técnicas

- TOCTOU eliminado em `ReadFromFile`: usa `os.Open` + `f.Stat()` em vez de `os.Stat` + `os.ReadFile`.
- TOCTOU eliminado em `ReadFromStdin`: verificação de terminal movida para dentro da função.
- Goroutine leak evitado em timeout: `os.Stdin.Close()` desbloqueia a leitura.
- Teste de timeout de stdin adicionado.
- `summarizer.New()` valida todos os campos obrigatórios.

---

## Fase 2 — Configuração via variáveis de ambiente (2026-07-03)

### 🔄 Mudanças na API Interna

- **`config.Load()` agora retorna `(Config, error)`**: validação acontece no ponto de carga,
  eliminando a necessidade do método `Validate()` separado
- **Método `Validate()` removido**: a validação agora é feita durante o `Load()`

### ✨ Novas Validações

- **`TLDR_BASE_URL` é validada como URL**: se for inválida, `Load()` retorna erro
- **`TLDR_TIMEOUT` inválido retorna erro**: valores não numéricos agora falham explicitamente
  (antes eram silenciosamente ignorados)
- **Erro mais descritivo ao falhar `secrets.LoadAPIKey()`**: mensagem genérica em vez de
  culpar exclusivamente `TLDR_API_KEY`

### 🔒 Segurança

- **`ProtectedAPIKey.Clear()` agora é nil-safe**: chamar `Clear()` em ponteiro nil é seguro
- **Documentação da sanitização**: regex de redação de chaves tem escopo explicado
  (aplicada apenas em mensagens de erro)

### 🧪 Testes

- Teste "sem API key" agora verifica os campos do `Config` retornado (não só o erro)
- Teste de URL base inválida adicionado
- `TestIsStdinAvailable` agora simula pipe real (não apenas loga o resultado)
- `TestClear` testa double-clear (não causa pânico)
- Testes de `buildPrompt` incluem casos `pt-br` e `pt`

### 📝 Documentação

- Comentários em `NewExitError`/`WrapExitError` com exemplos de uso
- Comentário de segurança sobre a ordem do `cfg.Clear()` em `root.go`
- Prompt padrão agora adapta ao idioma: `pt-br`/`pt` usam template em português

### 🔧 Manutenção

- Função auxiliar `cfgZeroed()` removida (substituída por inline)
- Dupla validação de timeout documentada em `summarizer.New()`

---

## Fase 1 — MVP (2026-07-02)

### 🚀 Funcionalidades Implementadas

- **CLI com Cobra**: Comando `tldr` com flags `--lang` (obrigatória), `--model`, `--prompt` e `--help`
- **Leitura de entrada**: Arquivo (argumento posicional) ou stdin (pipe/redirect), com limite de 10 MB
- **Codificação UTF-8**: Validação rigorosa — apenas UTF-8 é aceito
- **Integração com API OpenAI**: Via SDK oficial (`github.com/openai/openai-go`), compatível com provedor Apiário
- **Separação system/user message**: Prompt de sistema + texto do usuário (anti prompt injection)
- **Streams separadas**: Resumo no stdout, erros e logs no stderr
- **Exit codes**: 0 (sucesso), 1 (erro genérico), 2 (erro de API), 3 (argumento inválido)
- **Modelo configurável**: Flag `--model` > env var `TLDR_DEFAULT_MODEL` > hardcoded `deepseek/deepseek-v4-flash`
- **Timeout configurável**: Nova env var `TLDR_TIMEOUT` (segundos, default: 30)

### 🔒 Segurança

- `ProtectedAPIKey`: Wrapper que permite zerar a chave de API da memória após o uso
- Sanitização automática de credenciais em mensagens de erro (regex cobre `sk-`, `sk-proj-`, `api_key=`, `token=`)
- Separação de mensagens (sistema vs usuário) para mitigar prompt injection
- `.env.example` adicionado ao repositório (sem credenciais reais)

### 🧪 Cobertura de Testes

- **31 testes unitários** distribuídos em 6 pacotes
- Testes com `httptest.Server` simulando respostas da API (sucesso, 401, 429, 500, choices vazio, length, content_filter)
- Teste de integração real (opcional, com tag `integration`)
- Testes com race detector habilitados
- Zero dependências de API externa para testes unitários

### 📝 Documentação

- README completo com exemplos de uso, flags, env vars e exit codes
- `.env.example` com todas as variáveis de ambiente documentadas
- Spec (`projeto.md`) alinhada com a implementação

### 🔧 Melhorias Técnicas

- `SilenceErrors`/`SilenceUsage` para evitar duplicação de mensagens de erro
- `WithMaxRetries(0)` — sem retry automático da SDK, fail fast
- Detecção de truncamento em entrada stdin (>10MB)
- Mensagens de erro descritivas em português
- Makefile com targets: `build`, `test`, `test-race`, `test-integration`, `lint`, `clean`

### 📦 Estrutura do Repositório

```
tl-dr/
├── .env.example              # Template de ambiente (sem credenciais)
├── .github/workflows/ci.yml  # CI/CD com GitHub Actions
├── .gitignore
├── CHANGELOG.md              # Este arquivo
├── Makefile                  # Comandos auxiliares
├── README.md                 # Documentação do usuário
├── cmd/
│   ├── errors.go             # Tipos de erro com exit codes
│   ├── errors_test.go
│   ├── root.go               # Comando Cobra principal
│   └── root_test.go
├── internal/
│   ├── config/
│   │   ├── config.go         # Leitura de env vars
│   │   └── config_test.go
│   ├── input/
│   │   ├── input.go          # Leitura de arquivo e stdin
│   │   └── input_test.go
│   ├── integration/
│   │   └── summarizer_test.go # Teste de integração (opcional)
│   ├── secrets/
│   │   ├── secrets.go        # API key com limpeza de memória
│   │   └── secrets_test.go
│   └── summarizer/
│       ├── summarizer.go     # Chamada à API via OpenAI SDK
│       └── summarizer_test.go
├── main.go                   # Entry point
├── go.mod
├── go.sum
└── projeto.md                # Especificação do projeto
```
