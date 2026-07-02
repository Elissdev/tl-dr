# tl;dr — Resumidor de texto via CLI

CLI de sumarização de texto com suporte a múltiplos idiomas, via API compatível com OpenAI.

## Instalação

```bash
# Compilar a partir do código-fonte
git clone https://github.com/Elissdev/tl-dr
cd tl-dr
make build
```

O binário será gerado em `build/tldr`.

## Uso

```bash
# Exporte sua chave de API
export TLDR_API_KEY="sua-chave-aqui"

# Resume e traduz para português
cat texto.txt | tldr --lang pt-br

# Resume a partir de um arquivo
tldr arquivo.txt --lang en

# Usando um modelo específico
cat texto.txt | tldr --lang pt-br --model deepseek/deepseek-v4-flash

# Com prompt customizado
cat texto.txt | tldr --prompt "Resuma para um leigo no assunto" --lang pt-br
```

## Flags

| Flag            | Alias | Obrigatória | Descrição                                      |
|-----------------|-------|-------------|------------------------------------------------|
| `--lang`        | `-l`  | ✅ Sim      | Idioma do resumo (ex: `pt-br`, `en`, `es`)     |
| `--model`       | `-m`  | ❌ Não      | Modelo a usar (default: `deepseek/deepseek-v4-flash`) |
| `--prompt`      | `-p`  | ❌ Não      | Prompt customizado para o resumo               |
| `--help`        |       | ❌ Não      | Exibe ajuda                                    |

## Variáveis de Ambiente

| Variável             | Obrigatória | Padrão                            | Descrição                              |
|----------------------|-------------|-----------------------------------|----------------------------------------|
| `TLDR_API_KEY`       | ✅ Sim      | —                                 | Chave de API do provedor               |
| `TLDR_BASE_URL`      | ❌ Não      | `https://api.apiario.dev/v1`      | URL base da API (compatível OpenAI)    |
| `TLDR_DEFAULT_MODEL` | ❌ Não      | `deepseek/deepseek-v4-flash`      | Modelo padrão                          |
| `TLDR_DEFAULT_LANG`  | ❌ Não      | —                                 | Idioma padrão                          |
| `TLDR_TIMEOUT`       | ❌ Não      | `30`                              | Timeout da requisição em segundos      |

## Exit Codes

| Código | Significado            |
|--------|------------------------|
| `0`    | Sucesso                |
| `1`    | Erro genérico          |
| `2`    | Erro de API            |
| `3`    | Erro de argumento      |

## Exemplos

### 1. Resumo básico

```bash
echo "Lorem ipsum dolor sit amet..." | tldr --lang pt-br
```

### 2. Resumo de arquivo com modelo específico

```bash
tldr artigo.txt --lang en --model gpt-4
```

### 3. Prompt customizado

```bash
cat relatorio.txt | tldr --lang pt-br --prompt "Extraia apenas os números e dados estatísticos"
```

### 4. Pipe para outras ferramentas

```bash
tldr --lang en < document.txt | grep -i importante
```

## Desenvolvimento

```bash
# Compilar
make build

# Rodar testes
make test

# Lint
make lint

# Limpar artefatos
make clean
```

## Licença

A definir.
