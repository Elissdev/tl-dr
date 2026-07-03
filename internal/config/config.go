package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Elissdev/tl-dr/internal/secrets"
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
// Retorna erro se TLDR_API_KEY não estiver definida.
func Load() (Config, error) {
	cfg := Config{
		BaseURL:      getEnv("TLDR_BASE_URL", "https://api.apiario.dev/v1"),
		DefaultModel: getEnv("TLDR_DEFAULT_MODEL", "deepseek/deepseek-v4-flash"),
		DefaultLang:  os.Getenv("TLDR_DEFAULT_LANG"),
		Timeout:      30 * time.Second,
	}

	// Timeout configurável via TLDR_TIMEOUT (em segundos)
	if t := os.Getenv("TLDR_TIMEOUT"); t != "" {
		if secs, err := strconv.Atoi(t); err == nil && secs > 0 {
			cfg.Timeout = time.Duration(secs) * time.Second
		}
	}

	// Carrega a chave de API — obrigatória
	key, err := secrets.LoadAPIKey()
	if err != nil {
		return cfg, fmt.Errorf("TLDR_API_KEY não definida — configure a variável de ambiente")
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

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
