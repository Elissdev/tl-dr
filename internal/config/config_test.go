package config

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// setEnv configura uma variável de ambiente e retorna uma função de cleanup
// que restaura o valor anterior.
func setEnv(t *testing.T, key, value string) {
	t.Helper()
	prev, existed := os.LookupEnv(key)
	os.Setenv(key, value)
	t.Cleanup(func() {
		if existed {
			os.Setenv(key, prev)
		} else {
			os.Unsetenv(key)
		}
	})
}

// unsetEnv remove uma variável de ambiente e restaura no cleanup.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	prev, existed := os.LookupEnv(key)
	os.Unsetenv(key)
	t.Cleanup(func() {
		if existed {
			os.Setenv(key, prev)
		}
	})
}

func TestLoad(t *testing.T) {
	t.Run("config com todas as variáveis", func(t *testing.T) {
		setEnv(t, "TLDR_API_KEY", "sk-test-key-1234567890")
		setEnv(t, "TLDR_BASE_URL", "https://custom.api.com/v1")
		setEnv(t, "TLDR_DEFAULT_MODEL", "gpt-4")
		setEnv(t, "TLDR_DEFAULT_LANG", "en")
		setEnv(t, "TLDR_TIMEOUT", "60")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() erro inesperado: %v", err)
		}
		if cfg.APIKey != "sk-test-key-1234567890" {
			t.Errorf("APIKey = %q, want %q", cfg.APIKey, "sk-test-key-1234567890")
		}
		if cfg.BaseURL != "https://custom.api.com/v1" {
			t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, "https://custom.api.com/v1")
		}
		if cfg.DefaultModel != "gpt-4" {
			t.Errorf("DefaultModel = %q, want %q", cfg.DefaultModel, "gpt-4")
		}
		if cfg.DefaultLang != "en" {
			t.Errorf("DefaultLang = %q, want %q", cfg.DefaultLang, "en")
		}
		if cfg.Timeout != 60*time.Second {
			t.Errorf("Timeout = %v, want %v", cfg.Timeout, 60*time.Second)
		}
	})

	t.Run("config com valores padrão", func(t *testing.T) {
		unsetEnv(t, "TLDR_DEFAULT_MODEL")
		unsetEnv(t, "TLDR_DEFAULT_LANG")
		unsetEnv(t, "TLDR_BASE_URL")
		unsetEnv(t, "TLDR_TIMEOUT")
		setEnv(t, "TLDR_API_KEY", "sk-test-key")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() erro inesperado: %v", err)
		}
		if cfg.BaseURL != "https://api.apiario.dev/v1" {
			t.Errorf("BaseURL padrão = %q, want %q", cfg.BaseURL, "https://api.apiario.dev/v1")
		}
		if cfg.DefaultModel != "deepseek/deepseek-v4-flash" {
			t.Errorf("DefaultModel padrão = %q, want %q", cfg.DefaultModel, "deepseek/deepseek-v4-flash")
		}
		if cfg.DefaultLang != "" {
			t.Errorf("DefaultLang padrão = %q, want %q", cfg.DefaultLang, "")
		}
		if cfg.Timeout != 30*time.Second {
			t.Errorf("Timeout padrão = %v, want %v", cfg.Timeout, 30*time.Second)
		}
	})

	t.Run("config sem API key retorna erro", func(t *testing.T) {
		unsetEnv(t, "TLDR_API_KEY")

		cfg, err := Load()
		if err == nil {
			t.Fatal("Load() sem API key = nil, want erro")
		}
		if cfg.APIKey != "" {
			t.Errorf("cfg.APIKey = %q, want vazio", cfg.APIKey)
		}
		if cfg.BaseURL != "https://api.apiario.dev/v1" {
			t.Errorf("cfg.BaseURL = %q, want default", cfg.BaseURL)
		}
		if cfg.DefaultModel != "deepseek/deepseek-v4-flash" {
			t.Errorf("cfg.DefaultModel = %q, want default", cfg.DefaultModel)
		}
	})

	t.Run("timeout inválido retorna erro", func(t *testing.T) {
		setEnv(t, "TLDR_API_KEY", "sk-test-key")
		setEnv(t, "TLDR_TIMEOUT", "invalido")

		_, err := Load()
		if err == nil {
			t.Fatal("Load() com TLDR_TIMEOUT inválido = nil, want erro")
		}
	})

	t.Run("timeout zero usa padrão", func(t *testing.T) {
		setEnv(t, "TLDR_API_KEY", "sk-test-key")
		setEnv(t, "TLDR_TIMEOUT", "0")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() erro inesperado: %v", err)
		}
		if cfg.Timeout != 30*time.Second {
			t.Errorf("Timeout com valor zero = %v, want %v", cfg.Timeout, 30*time.Second)
		}
	})

	t.Run("URL base inválida retorna erro", func(t *testing.T) {
		setEnv(t, "TLDR_API_KEY", "sk-test-key")
		setEnv(t, "TLDR_BASE_URL", "://invalida")

		_, err := Load()
		if err == nil {
			t.Fatal("Load() com TLDR_BASE_URL inválida = nil, want erro")
		}
	})
}

func TestClear(t *testing.T) {
	setEnv(t, "TLDR_API_KEY", "sk-test-key-clear")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() erro inesperado: %v", err)
	}
	if cfg.APIKey == "" {
		t.Fatal("Load() deveria ter carregado a chave")
	}

	// Verifica que Clear() zera a chave visível e que a struct pode ser
	// usada novamente sem pânico (proteção contra double-clear).
	cfg.Clear()
	if cfg.APIKey != "" {
		t.Errorf("Clear() não zerou APIKey: %q", cfg.APIKey)
	}

	// Double-clear não deve causar pânico
	cfg.Clear()
	if cfg.APIKey != "" {
		t.Errorf("Clear() após double-clear não zerou APIKey: %q", cfg.APIKey)
	}
}

func TestAPIKeyBytes(t *testing.T) {
	setEnv(t, "TLDR_API_KEY", "sk-test-key-bytes")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() erro inesperado: %v", err)
	}

	t.Run("APIKeyBytes retorna slice com a chave", func(t *testing.T) {
		b := cfg.APIKeyBytes()
		if b == nil {
			t.Fatal("APIKeyBytes() = nil, want non-nil")
		}
		if string(b) != "sk-test-key-bytes" {
			t.Errorf("APIKeyBytes() = %q, want %q", string(b), "sk-test-key-bytes")
		}
	})

	t.Run("APIKeyBytes após Clear retorna nil", func(t *testing.T) {
		cfg.Clear()
		b := cfg.APIKeyBytes()
		if b != nil {
			t.Errorf("APIKeyBytes() após Clear = %q, want nil", string(b))
		}
	})
}

func TestCheckEnvPermissions(t *testing.T) {
	t.Run("sem .env não emite warning", func(t *testing.T) {
		// Remove .env temporariamente se existir
		_, err := os.ReadFile(".env")
		hadEnv := err == nil
		if hadEnv {
			if err := os.Rename(".env", ".env.bak"); err != nil {
				t.Fatalf("erro ao renomear .env: %v", err)
			}
			defer func() {
				if err := os.Rename(".env.bak", ".env"); err != nil {
					t.Errorf("erro ao restaurar .env: %v", err)
				}
			}()
		}

		// Captura stderr
		stderr := captureStderr(t, func() {
			checkEnvPermissions()
		})
		if stderr != "" {
			t.Errorf("sem .env não deveria emitir warning, got: %s", stderr)
		}
	})

	t.Run(".env com permissão 0600 não emite warning", func(t *testing.T) {
		err := os.WriteFile(".env_test_perms", []byte("TEST=1"), 0o600)
		if err != nil {
			t.Fatalf("erro ao criar arquivo temporário: %v", err)
		}
		defer os.Remove(".env_test_perms")

		// Renomeia .env temporário
		_, err = os.ReadFile(".env")
		hadEnv := err == nil
		if hadEnv {
			if err := os.Rename(".env", ".env.bak"); err != nil {
				t.Fatalf("erro ao renomear .env: %v", err)
			}
			defer func() {
				if err := os.Rename(".env.bak", ".env"); err != nil {
					t.Errorf("erro ao restaurar .env: %v", err)
				}
			}()
		}
		if err := os.Rename(".env_test_perms", ".env"); err != nil {
			t.Fatalf("erro ao renomear .env_test_perms para .env: %v", err)
		}
		defer func() {
			if err := os.Rename(".env", ".env_test_perms"); err != nil {
				t.Errorf("erro ao restaurar .env_test_perms: %v", err)
			}
		}()

		stderr := captureStderr(t, func() {
			checkEnvPermissions()
		})
		if stderr != "" {
			t.Errorf("0600 não deveria emitir warning, got: %s", stderr)
		}
	})

	t.Run(".env com permissão 0644 emite warning", func(t *testing.T) {
		err := os.WriteFile(".env_test_perms", []byte("TEST=1"), 0o644)
		if err != nil {
			t.Fatalf("erro ao criar arquivo temporário: %v", err)
		}
		defer os.Remove(".env_test_perms")

		_, err = os.ReadFile(".env")
		hadEnv := err == nil
		if hadEnv {
			if err := os.Rename(".env", ".env.bak"); err != nil {
				t.Fatalf("erro ao renomear .env: %v", err)
			}
			defer func() {
				if err := os.Rename(".env.bak", ".env"); err != nil {
					t.Errorf("erro ao restaurar .env: %v", err)
				}
			}()
		}
		if err := os.Rename(".env_test_perms", ".env"); err != nil {
			t.Fatalf("erro ao renomear .env_test_perms para .env: %v", err)
		}
		defer func() {
			if err := os.Rename(".env", ".env_test_perms"); err != nil {
				t.Errorf("erro ao restaurar .env_test_perms: %v", err)
			}
		}()

		stderr := captureStderr(t, func() {
			checkEnvPermissions()
		})
		if !strings.Contains(stderr, "WARNING") {
			t.Errorf("0644 deveria emitir warning, got: %q", stderr)
		}
		if !strings.Contains(stderr, "600") {
			t.Errorf("warning deveria recomendar chmod 600, got: %q", stderr)
		}
	})
}

// captureStderr captura a saída de os.Stderr durante a execução de f.
func captureStderr(t *testing.T, f func()) string {
	t.Helper()

	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("erro ao criar pipe: %v", err)
	}
	os.Stderr = w

	ch := make(chan string)
	go func() {
		data, _ := io.ReadAll(r)
		ch <- string(data)
	}()

	f()

	w.Close()
	os.Stderr = orig

	return <-ch
}
