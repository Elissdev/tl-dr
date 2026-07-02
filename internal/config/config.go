package config

import (
	"fmt"
	"os"

	"github.com/Elissdev/tl-dr/internal/secrets"
)

// Config armazena as configurações lidas de variáveis de ambiente.
type Config struct {
	APIKey       string
	BaseURL      string
	DefaultModel string
	DefaultLang  string

	protectedKey *secrets.ProtectedAPIKey // guarda referência para limpeza posterior
}

// Load lê as variáveis de ambiente e retorna um Config.
func Load() Config {
	cfg := Config{
		BaseURL:      getEnv("TLDR_BASE_URL", "https://apiario.dev/v1"),
		DefaultModel: getEnv("TLDR_DEFAULT_MODEL", "deepseek/deepseek-v4-flash"),
		DefaultLang:  os.Getenv("TLDR_DEFAULT_LANG"),
	}

	// Tenta carregar a chave de forma protegida; se falhar, Validate() captura depois.
	key, err := secrets.LoadAPIKey()
	if err == nil {
		cfg.APIKey = key.Get()
		cfg.protectedKey = key
	}

	return cfg
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

// Validate verifica se as configurações essenciais estão presentes.
func (c Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("TLDR_API_KEY não definida — configure a variável de ambiente")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
