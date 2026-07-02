package secrets

import (
	"os"
	"testing"
)

func TestLoadAPIKey(t *testing.T) {
	t.Run("chave presente", func(t *testing.T) {
		os.Setenv("TLDR_API_KEY", "sk-test-key-12345")
		defer os.Unsetenv("TLDR_API_KEY")

		key, err := LoadAPIKey()
		if err != nil {
			t.Fatalf("LoadAPIKey() erro inesperado: %v", err)
		}
		if key == nil {
			t.Fatal("LoadAPIKey() retornou nil")
		}
		if key.Get() != "sk-test-key-12345" {
			t.Errorf("Get() = %q, want %q", key.Get(), "sk-test-key-12345")
		}
	})

	t.Run("chave ausente", func(t *testing.T) {
		os.Unsetenv("TLDR_API_KEY")

		_, err := LoadAPIKey()
		if err == nil {
			t.Fatal("LoadAPIKey() sem chave = nil, want erro")
		}
	})

	t.Run("chave vazia", func(t *testing.T) {
		os.Setenv("TLDR_API_KEY", "")
		defer os.Unsetenv("TLDR_API_KEY")

		_, err := LoadAPIKey()
		if err == nil {
			t.Fatal("LoadAPIKey() com chave vazia = nil, want erro")
		}
	})
}

func TestProtectedAPIKeyClear(t *testing.T) {
	os.Setenv("TLDR_API_KEY", "sk-sensitive-key")
	defer os.Unsetenv("TLDR_API_KEY")

	key, err := LoadAPIKey()
	if err != nil {
		t.Fatal(err)
	}

	if key.Get() != "sk-sensitive-key" {
		t.Fatalf("Get() = %q, want %q", key.Get(), "sk-sensitive-key")
	}

	key.Clear()

	// Após Clear, o buffer interno deve estar zerado
	if key.Get() != "" {
		t.Error("Clear() não zerou o buffer interno")
	}
}

func TestProtectedAPIKeyGetCopy(t *testing.T) {
	os.Setenv("TLDR_API_KEY", "sk-copy-test")
	defer os.Unsetenv("TLDR_API_KEY")

	key, err := LoadAPIKey()
	if err != nil {
		t.Fatal(err)
	}

	// Get() retorna uma cópia; modificar a cópia não afeta o original
	got := key.Get()
	// Não podemos modificar strings em Go, mas verificamos que o valor está correto
	if got != "sk-copy-test" {
		t.Errorf("Get() = %q, want %q", got, "sk-copy-test")
	}
}
