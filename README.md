# tl;dr — Resumidor de texto via CLI

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.22.5-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![CI](https://github.com/Elissdev/tl-dr/actions/workflows/ci.yml/badge.svg)](https://github.com/Elissdev/tl-dr/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Elissdev/tl-dr)](https://github.com/Elissdev/tl-dr/releases)

</div>

CLI de sumarização de texto com suporte a múltiplos idiomas, via API compatível com OpenAI.

## Instalação

### Via binário (recomendado)

Baixe o binário mais recente para Linux/amd64 na [página de releases](https://github.com/Elissdev/tl-dr/releases):

```bash
# Baixar a última versão
curl -L -o tldr https://github.com/Elissdev/tl-dr/releases/latest/download/tldr-linux-amd64
chmod +x tldr
sudo mv tldr /usr/local/bin/
```

### Via go install

```bash
go install github.com/Elissdev/tl-dr@latest
```

### Compilar a partir do código-fonte

```bash
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
| `--lang`        | `-l`  | ✅ Sim¹     | Idioma do resumo (ex: `pt-br`, `en`, `es`)     |
| `--model`       | `-m`  | ❌ Não      | Modelo a usar (default: `deepseek/deepseek-v4-flash`) |
| `--prompt`      | `-p`  | ❌ Não      | Prompt customizado para o resumo               |
| `--timeout`     | `-t`  | ❌ Não      | Timeout da requisição em segundos (default: 30) |
| `--no-sanitize` |       | ❌ Não      | Desabilita sanitização de escape codes ANSI na saída |
| `--version`     | `-v`  | ❌ Não      | Exibe a versão                                  |
| `--help`        | `-h`  | ❌ Não      | Exibe ajuda                                    |

> ¹ O idioma também pode ser definido via variável de ambiente `TLDR_DEFAULT_LANG`.
> A flag `--lang` tem precedência sobre a variável de ambiente.

## Variáveis de Ambiente

| Variável             | Obrigatória | Padrão                            | Descrição                              |
|----------------------|-------------|-----------------------------------|----------------------------------------|
| `TLDR_API_KEY`       | ✅ Sim      | —                                 | Chave de API do provedor               |
| `TLDR_BASE_URL`      | ❌ Não      | `https://api.apiario.dev/v1`      | URL base da API (compatível OpenAI)    |
| `TLDR_DEFAULT_MODEL` | ❌ Não      | `deepseek/deepseek-v4-flash`      | Modelo padrão                          |
| `TLDR_DEFAULT_LANG`  | ❌ Não      | —                                 | Idioma padrão                          |
| `TLDR_TIMEOUT`       | ❌ Não      | `30`                              | Timeout da requisição em segundos (pode ser sobrescrito com --timeout) |

> **⚠️ Segurança:** O arquivo `.env` contém sua chave de API em texto claro.
> Mantenha-o com permissão `600` (`chmod 600 .env`) e **nunca** o commite no Git.

## Exit Codes

| Código | Significado               |
|--------|---------------------------|
| `0`    | Sucesso                   |
| `1`    | Erro interno/genérico     |
| `2`    | Erro de API               |
| `3`    | Erro de argumento         |
| `4`    | Timeout na requisição     |

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

### 4. Timeout customizado

```bash
cat texto_grande.txt | tldr --lang pt-br --timeout 120
```

### 5. Versionamento

```bash
tldr --version
```

### 6. Pipe para outras ferramentas

```bash
tldr --lang en < document.txt | grep -i importante
```

## Desenvolvimento

```bash
# Compilar
make build

# Rodar testes
make test

# Testes com race detector
make test-race

# Lint
make lint

# Limpar artefatos
make clean

# Criar release (tag + push)
make release v=v0.1.0
```

## Contribuição

Contribuições são bem-vindas! Siga os passos:

1. Faça um fork do repositório
2. Crie uma branch para sua feature: `git checkout -b feat/minha-feature`
3. Faça suas alterações
4. Rode os testes: `make test` e `make test-race`
5. Rode o linter: `make lint`
6. Faça commit e push
7. Abra um Pull Request

### Convenções

- **Commits:** use prefixos semânticos (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`)
- **Branches:** `feat/nome-da-feature` ou `fix/nome-do-fix`
- **Testes:** todo código novo deve ter testes
- **Lint:** o linter deve passar antes do merge

## Licença

Distribuído sob a licença MIT. Veja [LICENSE](LICENSE) para mais informações.
