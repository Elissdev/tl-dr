package config

import (
	"fmt"
	"os"
)

// Config armazena as configurações lidas de variáveis de ambiente.
type Config struct {
	APIKey       string
	BaseURL      string
	DefaultModel string
	DefaultLang  string
}

// Load lê as variáveis de ambiente e retorna um Config.
func Load() Config {
	return Config{
		APIKey:       os.Getenv("TLDR_API_KEY"),
		BaseURL:      getEnv("TLDR_BASE_URL", "https://apiario.dev/v1"),
		DefaultModel: getEnv("TLDR_DEFAULT_MODEL", "deepseek/deepseek-v4-flash"),
		DefaultLang:  os.Getenv("TLDR_DEFAULT_LANG"),
	}
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
