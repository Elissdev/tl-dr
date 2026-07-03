# Changelog

## Fase 2 — Streaming, Verbose e Multi-Plataforma (2026-07-03)

### 🚀 Funcionalidades Novas

- **Streaming de resposta**: O resumo é exibido token por token em tempo real (via `NewStreaming` do SDK OpenAI)
- **Modo verbose (`--verbose` / `-v`)**: Exibe informações de depuração no stderr:
  - Fonte de entrada (arquivo/stdin)
  - Idioma e modelo selecionados
  - Tamanho do texto
  - Prompt enviado à API
  - Indicador de progresso "⚡ Gerando resumo..."
  - Resumo final com contagem de caracteres
- **Multi-plataforma**: Builds para Linux (amd64), macOS Intel (amd64), macOS Apple Silicon (arm64) e Windows (amd64)

### 🔒 Melhorias no Tratamento de Erros

- **Detecção de contexto excedido (HTTP 400/413)**: Mensagem específica quando o texto excede o limite do modelo, sugerindo usar `--model` com contexto maior
- Função `isContextLengthError()` que detecta padrões como `context_length_exceeded`, `too many tokens`, `request too large`, etc.

### 🧪 Novos Testes

- **Testes de streaming**: Sucesso, erro 401, finish_reason `length` — cobrindo o fluxo completo de `SummarizeStream`
- **Testes de `extractContent`**: Normal, choices vazio, `length`, `content_filter`
- **Testes de `isContextLengthError`**: 12 casos incluindo todos os padrões de contexto excedido e falsos positivos
- **Testes de erro 400 (context)**: Verifica mensagem específica para contexto excedido
- **Testes de erro 413**: Verifica detecção de "request too large"
- **Testes de erro 400 genérico**: Verifica que erros 400 sem relação com contexto são tratados genericamente
- **Testes de flags**: Verifica existência e valores padrão de `--verbose`, `--lang`, `--model`, `--prompt`, e alias `-v`
- **Teste de `MaximumNArgs`**: Valida regra de argumentos do Cobra

### 📦 Build & Distribuição

- `make build-all`: Compila para todas as 4 plataformas
- CI agora faz upload de artefatos para todas as plataformas
- Release publica binários para Linux, macOS (Intel + Apple Silicon) e Windows

### 📝 Documentação

- README atualizado com flag `--verbose` e exemplo de uso
- README atualizado com exemplo de saída do modo verbose

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
