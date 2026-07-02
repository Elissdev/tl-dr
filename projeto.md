# tl;dr — Especificação do Projeto

> CLI de sumarização de texto com suporte a múltiplos idiomas, via API compatível com OpenAI.

---

## 1. Visão Geral

**tl;dr** é uma ferramenta de linha de comando que recebe um texto (de arquivo ou stdin) e produz um resumo conciso no idioma especificado pelo usuário. Utiliza uma API compatível com a OpenAI (provedor [Apiário](https://apiario.dev/)) para gerar os resumos.

- **Linguagem:** Go
- **Framework CLI:** Cobra
- **SDK da API:** OpenAI Go SDK (`github.com/openai/openai-go`)
- **Público:** Uso pessoal
- **Licença:** (a definir)

---

## 2. Exemplos de Uso (MVP)

```bash
# Resume e traduz pra português
cat text.md | tldr --lang pt-br

# Resume usando DeepSeek V4 Flash
cat text.md | tldr --lang pt-br --model deepseek/deepseek-v4-flash

# Resume com prompt customizado
cat text.md | tldr --prompt "Resuma para um leigo no assunto" --lang pt-br

# Resume de arquivo
tldr arquivo.txt --lang en
```

Os comandos acima **devem funcionar** — esse é o MVP.

---

## 3. Entrada de Dados

### 3.1. Fontes

A CLI aceita duas fontes de entrada, com a seguinte ordem de precedência:

1. **Arquivo** — se um argumento posicional (caminho de arquivo) for passado, lê dele.
2. **stdin** — se não houver argumento posicional, lê da entrada padrão (pipe ou redirecionamento).
3. **Erro** — se nenhuma das duas fontes estiver disponível, exibe erro e sai com código 1.

### 3.2. Codificação

Apenas **UTF-8** é suportada. Qualquer outra codificação resultará em erro.

### 3.3. Tamanho Máximo

Se o texto de entrada exceder o limite de contexto do modelo, o programa **envia mesmo assim** e, se a API retornar erro (400, 413, etc.), **avisa o usuário e encerra**.

---

## 4. Saída de Dados

### 4.1. Streams

| Stream | Conteúdo |
|--------|----------|
| **stdout** | Apenas o resumo gerado |
| **stderr** | Logs de debug, progresso, erros |

Isso permite redirecionar ou encadear a saída: `tldr --lang pt-br < text.txt | grep ...`

### 4.2. Exit Codes

| Código | Significado |
|--------|-------------|
| `0` | Sucesso |
| `1` | Erro genérico (entrada inválida, etc.) |
| `2` | Erro de API (rate limit, autenticação, etc.) |
| `3` | Erro de argumento inválido |

---

## 5. Interface da CLI

### 5.1. Comando

```
tldr [flags] [<arquivo>]
```

### 5.2. Flags

| Flag | Alias | Obrigatório | Descrição |
|------|-------|-------------|-----------|
| `--lang` | `-l` | ✅ Sim | Idioma do resumo (ex: `pt-br`, `en`, `es`) |
| `--model` | `-m` | ❌ Não | Modelo a usar (default: `deepseek/deepseek-v4-flash`) |
| `--prompt` | `-p` | ❌ Não | Prompt customizado para o resumo |
| `--help` | | ❌ Não | Exibe ajuda |

### 5.3. Ordem de Precedência — Modelo

`--model` (flag) > `TLDR_DEFAULT_MODEL` (env) > `deepseek/deepseek-v4-flash` (hardcoded)

### 5.4. Prompt Default

Quando `--prompt` não é fornecido, o prompt padrão é algo como:

```
Summarize the following text in {idioma}. Be concise but capture all key points.
```

### 5.5. Prompt Customizado + --lang

Quando ambos são fornecidos, **ambos são respeitados**: o prompt customizado dita o estilo/conteúdo, e o `--lang` dita o idioma da resposta. O prompt final informado à API deve incluir instrução para responder em `{idioma}`.

---

## 6. Configuração (Variáveis de Ambiente)

| Variável | Obrigatória | Padrão | Descrição |
|----------|-------------|--------|-----------|
| `TLDR_API_KEY` | ✅ Sim | — | Chave de API do provedor |
| `TLDR_BASE_URL` | ❌ Não | `https://apiario.dev/v1` | URL base da API (compatível OpenAI) |
| `TLDR_DEFAULT_MODEL` | ❌ Não | `deepseek/deepseek-v4-flash` | Modelo padrão |
| `TLDR_DEFAULT_LANG` | ❌ Não | `""` | Idioma padrão (se não passado via `--lang`) |

Não há arquivo de configuração — tudo é via variáveis de ambiente.

---

## 7. API — Provedor

- **Provedor:** [Apiário](https://apiario.dev/)
- **Compatibilidade:** 100% compatível com a API da OpenAI (formato, endpoints, streaming)
- **SDK usado:** `github.com/openai/openai-go` (com `BaseURL` configurado para `TLDR_BASE_URL`)
- **Modelo padrão:** `deepseek/deepseek-v4-flash`

### Tratamento de Erros de API

- Rate limit (429): avisa o usuário no stderr e sai com código 2
- Auth (401): avisa o usuário no stderr e sai com código 2
- Contexto excedido (400/413): avisa o usuário no stderr e sai com código 2

---

## 8. Estrutura do Repositório

```
tl-dr/
├── cmd/
│   └── root.go              # Cobra root command (flags, execução)
├── internal/
│   ├── config/
│   │   └── config.go        # Leitura de variáveis de ambiente
│   ├── input/
│   │   └── input.go         # Leitura de arquivo e stdin
│   └── summarizer/
│       └── summarizer.go    # Chamada à API via OpenAI SDK
├── .github/
│   └── workflows/
│       └── ci.yml           # CI/CD com GitHub Actions
├── main.go                  # Entry point
├── go.mod
├── go.sum
├── Makefile                 # Comandos auxiliares (build, test, lint)
├── projeto.md               # Este documento
└── README.md                # Documentação do usuário
```

---

## 9. Testes

### 9.1. Estratégia

- **Testes unitários:** para parsing de argumentos, lógica de input, tratamento de erros
- **Testes de integração com API:** usando [Cassete](https://github.com/vhs/cassete) (VHS Cassette)
  - Gravar cassetes com a API real **uma vez**
  - Rodar testes sempre contra os cassetes gravados
  - Cassetes versionados no Git

### 9.2. Cobertura

| Área | Tipo |
|------|------|
| Parsing de flags/args | Unitário |
| Leitura de arquivo | Unitário |
| Leitura de stdin | Unitário |
| Construção do prompt | Unitário |
| Chamada à API | Integração (com cassete) |
| Tratamento de erros da API | Integração (com cassete) |

---

## 10. Build & Distribuição

### 10.1. Plataformas Alvo

| Plataforma | Arquitetura |
|------------|-------------|
| Linux | `amd64` |

*Futuro: darwin/amd64, darwin/arm64, windows/amd64*

### 10.2. Distribuição

Binários pré-compilados disponíveis no **GitHub Releases**.

### 10.3. Versionamento

**SemVer** (semantic versioning).

### 10.4. CI/CD (GitHub Actions)

| Pipeline | Gatilho | Ações |
|----------|---------|-------|
| **test** | `push` / `pull_request` | `go test ./...` (com cassetes) |
| **lint** | `push` / `pull_request` | `golangci-lint` |
| **build** | `push` / `pull_request` | Compila para `linux/amd64` |
| **release** | Tag `v*` | Sobe binário no GitHub Releases |

---

## 11. Dependências Principais

| Pacote | Propósito |
|--------|-----------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/openai/openai-go` | SDK da API OpenAI (via Apiário) |
| `github.com/vhs/cassete` | Testes de integração com cassete HTTP |

---

## 12. MVP (Já é o suficiente)

O MVP consiste nos comandos de exemplo da seção 2 funcionarem corretamente:

- Leitura de arquivo e stdin
- Flag `--lang` obrigatória
- Flag `--model` opcional
- Flag `--prompt` opcional
- Chamada à API do Apiário com OpenAI SDK
- Saída do resumo no stdout
- Erros no stderr com exit codes apropriados

Nada além disso é necessário para o primeiro release.

---

## 13. Decisões de Design (Registradas)

| Decisão | Opção Escolhida |
|---------|-----------------|
| Entrada simultânea (arquivo + stdin) | Arquivo primeiro, depois stdin |
| Sem entrada | Erro (exit 1) |
| Texto excede limite do modelo | Envia e trata erro da API |
| Codificação | Apenas UTF-8 |
| Configuração | Apenas variáveis de ambiente |
| Testes de API | Cassete (grava uma vez, replay sempre) |
| Streaming da resposta | (a decidir — pode ser futuro) |
