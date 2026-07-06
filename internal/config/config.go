package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/Elissdev/tl-dr/internal/secrets"
	"github.com/joho/godotenv"
)

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

	return cfg, nil
}

// Clear zera a chave de API da memória. Deve ser chamado assim que a chave
// não for mais necessária (após criar o cliente da API).
func (c *Config) Clear() {
	if c.protectedKey != nil {
		c.protectedKey.Clear()
		c.protectedKey = nil
	}
	c.APIKey = ""
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
