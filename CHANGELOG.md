# Changelog

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
