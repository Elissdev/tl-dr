# Changelog

## PR #20 — Fase 8: Documentação e Release — README, licença, primeiro release (2026-07-08)

### 🟢 Novas Funcionalidades

#### Licença MIT

- **Arquivo `LICENSE` adicionado**: licença MIT com copyright 2026 Elissandra Santos da Silva
- README atualizado com badge e link para a licença

#### Release via Makefile

- **`make release v=v0.1.0`**: cria tag SemVer e faz push (dispara o job `release` no CI)
- Target valida que o parâmetro `v` foi fornecido

#### Badges no README

- **Go Version**: badge da versão do Go (1.22.5)
- **License**: badge MIT com link para LICENSE
- **CI**: badge de status do GitHub Actions (build/test/lint)
- **Release**: badge da última versão publicada

### 📝 Documentação

#### README expandido

- **Instalação**: 3 métodos — binário pre-compilado (curl), `go install`, e compilação do fonte
- **Seção de Contribuição**: guia passo a passo com fork, branch, testes, lint e abertura de PR
- **Convenções**: prefixos semânticos de commit, padrão de nomenclatura de branches
- **Licença**: atualizada de "A definir" para MIT com link para o arquivo LICENSE
- **Makefile**: comando `make release` adicionado à lista de comandos de desenvolvimento

---

## PR #19 — Fase 7: CI/CD com GitHub Actions + Correções de Code Review (2026-07-08)

### 🟢 Novas Funcionalidades

#### CI/CD com GitHub Actions

- **4 jobs independentes**: `test`, `lint`, `build`, `release` com dependências claras (`needs`)
- **CI em todas as branches**: triggers `push` e `pull_request` sem restrição de branch
- **Release automática**: disparada em tags `v*`, com `github.event_name != 'pull_request'` para evitar release acidental em PRs
- **Race detector**: `go test -race -count=1` em todo o codebase
- **Gosec security scan**: `gosec -exclude-dir=internal/integration ./...`
- **Upload de artifact**: binário compilado disponível como artifact dos PRs
- **Cross-compilação**: `GOOS=linux GOARCH=amd64 go build` para consistência

#### Constantes nomeadas para UTF-8 e ANSI

- **`cmd/root.go`**: números mágicos substituídos por constantes nomeadas:
  - `utf8Lead2`/`utf8Lead2End`, `utf8Lead3`/`utf8Lead3End`, `utf8Lead4`/`utf8Lead4End`
  - `utf8ContStart`/`utf8ContEnd`, `c1Start`/`c1End`
  - `csi8bit`, `osc8bit`, `dcs8bit`, `sos8bit`, `pm8bit`, `apc8bit`, `st8bit`, `sci8bit`
  - `bel`, `esc`, `c0Min`/`c0Max`, `csiFinalMin`/`csiFinalMax`

#### Case-insensitive locale lookup

- **`getLocale()`** agora normaliza o idioma com `strings.ToLower()` + `strings.ReplaceAll("_", "-")`, aceitando `"PT-BR"`, `"pt_BR"`, `"Pt-br"` como equivalentes a `"pt-br"`

### 🔧 Melhorias

- **Validação de `--timeout < 0`**: retorna `ExitArgs` com mensagem clara em vez de comportamento indefinido
- **Fail fast**: validação de `--timeout` movida para antes de `ReadInput` — feedback instantâneo sem esperar I/O
- **Help do `--timeout`**: descrição atualizada para "deve ser > 0; default: 30"
- **`finish_reason` com switch**: tratamento explícito de `stop`, `length`, `content_filter` + fallback para valores desconhecidos (`tool_calls`, `function_call`)
- **`apiKeyCopy` zerada após uso**: `defer` com loop de zeroing em `Summarize()` minimiza janela de exposição da chave
- **CI juda `push`** já não mais restrito a `main`

### 🧪 Testes

- `TestGetLocaleCaseInsensitive` — 6 subtestes cobrindo `PT-BR`, `pt_BR`, `Pt-br`, `PT`, equivalência entre variantes, `EN` maiúsculo
- `TestExecute/--timeout_negativo_(fail_fast_antes_de_IO)` — validação sem necessidade de API key ou arquivo

---

## PR #18 — Fase 6: Testes Unitários e de Integração com Cassete (2026-07-08)

### 🔴 Breaking Changes

#### `ProtectedAPIKey.Bytes()` — agora retorna cópia defensiva (não referência direta)

- **Antes:** Retornava referência direta ao slice interno — mutação externa afetava o original
- **Depois:** Retorna cópia defensiva via `make([]byte, len) + copy()` — mutação externa não afeta interno
- **Ação:** Código que dependia da mutação do slice para modificar a chave precisa ser revisado

#### `Config.APIKeyBytes()` — renomeado para `apiKeyBytes()` (não exportado)

- **Antes:** `cfg.APIKeyBytes()` — exportado, disponível publicamente
- **Depois:** `cfg.apiKeyBytes()` — não exportado, uso interno apenas
- **Ação:** Código externo deve acessar chave via `cfg.APIKey` (string) ou `secrets.ProtectedAPIKey.Bytes()`

#### `checkEnvPermissions()` — renomeado e com assinatura alterada

- **Antes:** `checkEnvPermissions()` — fixo para `.env`, sem parâmetros
- **Depois:** `checkFilePermissions(path string)` — genérico, recebe caminho do arquivo
- **Ação:** Código que chamava `checkEnvPermissions()` deve usar `checkFilePermissions(".env")`

#### `envPermsWarn` → `filePermsWarn` (constante não exportada)

- Mensagem alterada de português (`"⚠️  AVISO:"`) para inglês (`"⚠️  WARNING:"`)
- Mensagem agora inclui o caminho do arquivo específico

---

### 🟢 Novas Funcionalidades

#### Testes de Integração com Cassete (go-vcr)

- **`gopkg.in/dnaeon/go-vcr.v3 v3.2.0`**: dependência para gravação/reprodução de interações HTTP
- **5 cassetes YAML**: `summarize_success`, `summarize_short`, `summarize_unauthorized`, `summarize_rate_limited`, `summarize_server_error`
- **Hook `BeforeSaveHook`**: redige header `Authorization` e body com credenciais antes de persistir
- **Build tag `integration`**: `go test -tags=integration ./internal/integration/`
- **Modo Record/Replay**: `TLDR_CASSETE_MODE=record` grava; padrão é replay offline

#### Segurança

- **`validateHTTPClientTLS()`**: rejeita `http.Client` com `InsecureSkipVerify=true`
- **`RedactCredentials()` exportada**: função pública para redigir credenciais em qualquer contexto
- **`Client.mu sync.Mutex`**: proteção thread-safe entre `Summarize()` e `Clear()`
- **`Client.cleared` flag**: detecta uso após Clear() e retorna erro (em vez de panic)
- **Cópia defensiva da apiKey em `Summarize()`**: copia chave antes do unlock para `classifyAPIError` seguro
- **Lone continuation bytes**: `sanitizeOutput()` descarta bytes 0x80-0xBF isolados (sem lead byte)

#### Performance

- **Pré-filtro de keywords**: `sanitizePrompt()` verifica `injectionKeywords` primeiro; se ausentes, pula regexes
- **`injectionKeywords` expandido**: `system`, `user`, `assistant` adicionados como keywords (não patterns)

#### Engine

- **`Config.HTTPClient`**: campo opcional para injetar `http.Client` customizado (ex: go-vcr)
- **`classifyAPIError` reordenado**: erros HTTP (401, 429, 500) detectados **antes** de fallback por substring
- **Release job no CI**: `github.event_name != 'pull_request'` impede release acidental em PRs

---

### 🔧 Melhorias

- `ProtectedAPIKey.Bytes()` retorna cópia defensiva (make + copy)
- `sanitizePrompt()`: pré-filtro de keywords evita scan de regex em prompts inocentes
- `sanitizeOutput()`: lone continuation bytes descartados; tracking UTF-8 só atualizado para bytes escritos
- `classifyAPIError()`: ordem corrigida — status HTTP antes de substring (evita falsos timeout)
- `checkFilePermissions(path)`: genérico, aceita qualquer caminho; erros de stat logados no stderr
- Aviso de permissão combinado: se arquivo for legível para grupo E outros, ambos são mencionados
- `TLDR_API_KEY_FILE` também tem permissões verificadas após o Load()
- Lint SA9003 corrigido: branch vazia removida em `validateHTTPClientTLS`

---

### 🧪 Testes

#### Testes de Integração com Cassete (+332 linhas, novo arquivo)

- `TestSummarizeWithCassette` — sucesso pt-br (200)
- `TestSummarizeWithCassetteShortText` — sucesso en (200)
- `TestSummarizeAuthError` — 401 para credenciais inválidas
- `TestSummarizeRateLimitError` — 429 rate limit
- `TestSummarizeServerError` — 500 servidor indisponível
- `TestSummarizeContextCanceled` — contexto cancelado sem panic
- `TestSummarizeContextDeadline` — deadline expirado como timeout

#### Novos Testes Unitários

- `TestCheckFilePermissions` — 8 casos: inexistente, 0600, 0640 (grupo), 0644 (outros), subdiretório, diretório, 000, EACCES
- `TestExecuteGlobal` — Execute() lazy-init sync.Once, idempotência, --help
- `TestInitRootCommand` — flags --lang/--model/--prompt registradas
- `TestSanitizePrompt` — +6 casos: system tag total/residual, assistant, user, pré-filtro, keyword sem pattern
- `TestSanitizeOutputC1Bytes` — +5 casos: UTF-8 0x80, 0xA0-0xBF, 3-byte, lone continuation byte

---

## PR #15 — Fase 4: Integração com API (2026-07-07)

### 🔴 Breaking Changes

#### `summarizer.Client.Summarize()` — `panic` substituído por `error`

- **Antes:** Chamar `Summarize()` após `Clear()` causava `panic`
- **Depois:** Retorna erro `"summarizer: Client usado após Clear()"`
- **Ação:** Se seu código capturava o panic com `recover`, agora deve tratar o erro normalmente

#### `summarizer.Client.apiKey` — tipo alterado de `string` para `[]byte`

- Campo não exportado do pacote (afeta apenas código interno)
- Função `redactCredentials` teve assinatura atualizada: `(s string, apiKey []byte)`

#### `sanitizePrompt()` — flag `--prompt` agora sanitizada

- **Antes:** Prompt customizado passava diretamente para a API
- **Depois:** Padrões de injeção são substituídos por `[REMOVED]`; prompt 100% injetivo causa erro
- **Mitigação:** Regex calibrada; `"now"` não é mais gatilho isolado (evita falso positivo)

#### `config.Load()` — novo side effect em stderr

- **Antes:** `Load()` não escrevia em stderr
- **Depois:** Se `.env` tiver permissões >0600, emite warning no stderr
- **Mitigação:** Use `chmod 600 .env` para silenciar

#### `secrets.LoadAPIKey()` — mensagem de erro alterada

- **Antes:** `"TLDR_API_KEY não definida"`
- **Depois:** `"TLDR_API_KEY não definida (defina a variável ou TLDR_API_KEY_FILE)"`

#### `config.Clear()` — semântica reforçada

- **Antes:** Usar struct após Clear() era possível (APIKey retornava "")
- **Depois:** Contrato explicita que struct não deve mais ser usada após Clear()

### 🟢 Novas Funcionalidades

- **`TLDR_API_KEY_FILE`**: ler chave de API de arquivo (fallback quando `TLDR_API_KEY` não definida)
- **`secrets.ProtectedAPIKey.Bytes()`**: retorna cópia da chave como `[]byte` (defensiva)
- **`config.Config.APIKeyBytes()`**: acessa chave como `[]byte` via Config
- **`summarizer.Client.Clear()`**: zera a chave de API da memória do Client
- **`sanitizePrompt()`**: sanitização de prompts customizados contra injeção
- **Tratamento de bytes C1 (0x80-0x9F)** em `sanitizeOutput()` com estado UTF-8

### 🔧 Melhorias

- `ProtectedAPIKey.Bytes()` retorna cópia defensiva (não mais referência direta ao slice interno)
- Regex de prompt injection: `now` removido como gatilho (falso positivo)
- `sanitizeOutput()`: corrigido bug onde `0x1b` (ESC) podia ser escrito na saída quando `utf8Remaining > 0`
- `Summarize()` pós-Clear retorna erro em vez de panic
- Warning de permissão do `.env` em inglês
- `cfg.Clear()` agora é chamado imediatamente (sem defer); `s.Clear()` mantém `defer`

### 🧪 Testes

- `TestSanitizePrompt` — 13 casos incluindo normal, vazio, injeção total/parcial, padrões mistos
- `TestSanitizeOutputC1Bytes` — CSI, OSC, DCS, SOS, PM, APC, ST, C1 genérico, misto com ESC
- `TestAPIKeyBytes` — slice correto, nil após Clear
- `TestProtectedAPIKeyBytes` — retorna cópia (mutação não afeta original), nil após Clear
- `TestCheckEnvPermissions` — sem .env, 0600, 0644 com warning
- `TestClientClear` — Clear zera apiKey, Summarize após Clear retorna erro, double Clear seguro

---

## PR #14 — Conclusão das 11 Code Reviews (2026-07-03)

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
- **Flag `--timeout` / `-t`**: timeout customizável via CLI (sobrescreve `TLDR_TIMEOUT`).
- **Flag `--version` / `-v`**: exibe versão do binário (injetada via ldflags).
- **Validação de idioma**: formato do `--lang` é validado (ex: `pt-br`, `en`, `zh-CN`).
- **Redação estendida de credenciais**: cobre DeepSeek, Anthropic, GitHub PAT, JWT,
  e fallback genérico para strings de 60+ caracteres.
- **Erros preservam cadeia original**: `classifyAPIError` agora retorna `*redactedError`
  que preserva a cadeia de erros para `errors.Is`/`errors.As`.
- **Sentinel errors**: `ErrTruncated` (conteúdo parcial + truncamento) e `ErrTimeout`
  (detectável via `errors.Is`).
- **Safety prefix anti-prompt injection**: prefixo imutável em pt e en inserido antes
  de todo prompt.
- **Feedback visual no stderr**: exibe idioma, modelo e progresso ("📝 Resumindo...").
- **Expansão de `~` em caminhos de arquivo**: `tldr ~/documento.txt` funciona.

### 🔧 Melhorias Técnicas

- TOCTOU eliminado em `ReadFromFile`: usa `os.Open` + `f.Stat()` em vez de `os.Stat` + `os.ReadFile`.
- TOCTOU eliminado em `ReadFromStdin`: verificação de terminal movida para dentro da função.
- Goroutine leak evitado em timeout: leitura em goroutine com `context.WithTimeout`.
- `unsafe` removido de `secrets.go`: zerar buffer de string com `unsafe.Pointer` não é confiável.
- `summarizer.New()` valida todos os campos obrigatórios (`APIKey`, `Model`, `BaseURL`).
- `newRootCommand()`: comando raiz sem estado global/`init()` — testável isoladamente.
- `localeConfig` + `supportedLocales`: adicionar idioma = adicionar entrada no mapa.
- `apiKeyRedactors` como slice de regexes específicas (em vez de uma regex gigante).
- `redactedError` preserva cadeia de erros original para `errors.Is`/`errors.As`.
- Validação de URL agora exige `http://` ou `https://` no scheme.
- `envOr()` renomeado de `getEnv()` para clareza.
- CI: testes com race detector + gosec security scan.
- Makefile: versão injetada via ldflags.

### 🧪 Testes

- `TestBuildPrompt` — verifica prefixo de segurança + sufixo correto
- `TestBuildPromptSafetyPrefixAlwaysPresent` — todas as combinações de idioma
- `TestSanitizeOutput` — CSI, OSC, DCS, SOS, PM, APC, newlines, tabs, CR
- `TestSanitizeOutputEdgeCases` — ESC isolado, incompleto, múltiplo, unicode
- `TestGetLocale` — pt-br, pt, en, fallback, não quebra existentes
- `TestFirstNonEmpty` — vários cenários de precedência
- `TestExecute` — langPattern, flags, env vars, sem API key, idioma inválido
- `TestNew` (summarizer) — validação de campos obrigatórios, timeout zero
- `TestSummarize` — `finish_reason=stop` vazio, erro 400
- `TestClassifyAPIErrorSanitization` — sk-, sk-proj-, api_key=, token=, própria chave, timeout
- `TestRedactCredentials` — chave configurada, vazia, string vazia
- Stdin timeout com `ReadFromStdinWithTimeout`
- `TestIsStdinRedirected` determinístico (pipe vs terminal)

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
