package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Elissdev/tl-dr/internal/secrets"
	"github.com/joho/godotenv"
)

// filePermsWarn é o prefixo para o aviso de permissões de arquivos.
// Usar stderr evita misturar com a saída do resumo no stdout.
// Mantido em inglês por ser um aviso de segurança universal.
const filePermsWarn = "⚠️  WARNING: "

// checkFilePermissions verifica as permissões do arquivo no caminho
// informado e emite um aviso no stderr se estiver legível para outros
// usuários (group/others).
//
// Recomendação: chmod 600 <arquivo>
//
// ATENÇÃO: Esta verificação é BEST-EFFORT. Entre o stat() e o uso do
// arquivo, um atacante com acesso ao sistema de arquivos pode alterar
// as permissões (TOCTOU). Esta função serve como alerta preventivo,
// não como garantia de segurança. A proteção real vem de:
//   - Configuração correta de permissões no momento da criação do arquivo
//   - Uso de variáveis de ambiente em vez de arquivos, quando possível
//   - Sistemas de secrets gerenciados (ex: HashiCorp Vault, Docker secrets)
func checkFilePermissions(path string) {
	info, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			// Erro inesperado ao acessar o arquivo (ex: permissão negada no
			// diretório pai, broken symlink, filesystem offline). Avisamos no
			// stderr para debug, sem interromper o fluxo.
			msg := fmt.Sprintf("%s%s: não foi possível verificar permissões: %v\n",
				filePermsWarn, path, err)
			_, _ = os.Stderr.WriteString(msg)
		}
		return
	}
	if info.IsDir() {
		return
	}

	// ModePerm = bits de permissão (Unix: 0x1FF = 0777)
	perm := info.Mode().Perm()

	// Verifica se o arquivo é legível por "group" (0o040) ou "others" (0o004)
	// A permissão segura recomendada é 0o600 (apenas owner)
	const (
		groupRead  = 0o040
		othersRead = 0o004
	)

	var warnings []string
	if perm&othersRead != 0 {
		warnings = append(warnings, "legível para outros usuários")
	}
	if perm&groupRead != 0 {
		warnings = append(warnings, "legível para o grupo")
	}

	if len(warnings) > 0 {
		details := strings.Join(warnings, " e ")
		msg := fmt.Sprintf("%s%s tem permissões %04o — %s. Recomendado: chmod 600\n",
			filePermsWarn, path, perm, details)
		_, _ = os.Stderr.WriteString(msg)
	}
}

// Config armazena as configurações lidas de variáveis de ambiente.
type Config struct {
	APIKey       string
	BaseURL      string
	DefaultModel string
	DefaultLang  string
	Timeout      time.Duration

	protectedKey *secrets.ProtectedAPIKey // guarda referência para limpeza posterior
}

// Load lê as variáveis de ambiente e retorna um Config.
// Retorna erro se a chave de API não puder ser carregada ou se alguma
// variável de ambiente obrigatória for inválida.
func Load() (Config, error) {
	// Tenta carregar .env; se houver erro diferente de "arquivo não encontrado", reporta
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{APIKey: ""}, fmt.Errorf("erro ao carregar .env: %w", err)
	} else if err == nil {
		// .env foi carregado com sucesso — verifica permissões do arquivo
		checkFilePermissions(".env")
	}

	cfg := Config{
		BaseURL:      envOr("TLDR_BASE_URL", "https://api.apiario.dev/v1"),
		DefaultModel: envOr("TLDR_DEFAULT_MODEL", "deepseek/deepseek-v4-flash"),
		DefaultLang:  os.Getenv("TLDR_DEFAULT_LANG"),
		Timeout:      30 * time.Second,
	}

	// Valida a URL base (deve ter scheme http ou https)
	u, err := url.ParseRequestURI(cfg.BaseURL)
	if err != nil {
		return cfg, fmt.Errorf("TLDR_BASE_URL inválida: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return cfg, fmt.Errorf("TLDR_BASE_URL deve começar com http:// ou https://, got %q", cfg.BaseURL)
	}

	// Timeout configurável via TLDR_TIMEOUT (em segundos)
	if t := os.Getenv("TLDR_TIMEOUT"); t != "" {
		secs, err := strconv.Atoi(t)
		if err != nil {
			return cfg, fmt.Errorf("TLDR_TIMEOUT inválido: %q — deve ser um número inteiro de segundos", t)
		}
		if secs > 0 {
			cfg.Timeout = time.Duration(secs) * time.Second
		}
	}

	// Carrega a chave de API — obrigatória
	key, err := secrets.LoadAPIKey()
	if err != nil {
		return cfg, fmt.Errorf("falha ao carregar chave de API: %w", err)
	}
	cfg.APIKey = key.Get()
	cfg.protectedKey = key

	// Verifica permissões do arquivo de chave (TLDR_API_KEY_FILE)
	// após o Load para evitar TOCTOU entre stat e leitura.
	if path := os.Getenv("TLDR_API_KEY_FILE"); path != "" {
		checkFilePermissions(path)
	}

	return cfg, nil
}

// Clear zera a chave de API da memória. Deve ser chamado assim que a chave
// não for mais necessária (após criar o cliente da API).
// Após chamar Clear, a struct não deve mais ser usada.
//
// NOTA: A string c.APIKey é uma cópia do []byte interno do ProtectedAPIKey,
// e o seu backing array NÃO pode ser zerado de forma portável em Go.
// No entanto, ao substituir a string por "" e chamar o GC, o backing array
// original fica elegível para coleta. A proteção principal vem de:
//  1. ProtectedAPIKey.Clear() — zera o []byte original
//  2. s.Clear() no summarizer.Client — zera a cópia interna em []byte
func (c *Config) Clear() {
	if c.protectedKey != nil {
		c.protectedKey.Clear()
		c.protectedKey = nil
	}
	c.APIKey = ""
}

// apiKeyBytes retorna uma cópia da chave de API como []byte para uso em
// contextos onde o caller pode gerenciar o ciclo de vida da memória.
// O caller pode zerar a cópia sem afetar o estado interno.
// Retorna nil se o ProtectedAPIKey não estiver mais disponível.
func (c *Config) apiKeyBytes() []byte {
	if c.protectedKey == nil {
		return nil
	}
	return c.protectedKey.Bytes()
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
