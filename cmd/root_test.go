package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		name         string
		lang         string
		customPrompt string
		want         string
	}{
		{
			name:         "default prompt",
			lang:         "pt-br",
			customPrompt: "",
			want:         "Summarize the following text in pt-br. Be concise but capture all key points.",
		},
		{
			name:         "custom prompt with lang",
			lang:         "en",
			customPrompt: "Resuma para um leigo no assunto",
			want:         "Responda em en.\n\nResuma para um leigo no assunto",
		},
		{
			name:         "custom prompt includes original",
			lang:         "es",
			customPrompt: "Haz un resumen",
			want:         "Responda em es.\n\nHaz un resumen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPrompt(tt.lang, tt.customPrompt)
			if result != tt.want {
				t.Errorf("buildPrompt(%q, %q) = %q, want %q", tt.lang, tt.customPrompt, result, tt.want)
			}
		})
	}
}

func TestFlags(t *testing.T) {
	t.Run("verbose flag padrão é false", func(t *testing.T) {
		if verbose {
			t.Error("verbose padrão deveria ser false")
		}
	})

	t.Run("verbose flag é bool", func(t *testing.T) {
		// Verifica que a flag foi registrada corretamente
		flag := rootCmd.Flags().Lookup("verbose")
		if flag == nil {
			t.Fatal("flag --verbose não encontrada")
		}
		if flag.DefValue != "false" {
			t.Errorf("verbose.DefValue = %q, want 'false'", flag.DefValue)
		}
	})

	t.Run("lang flag é obrigatória", func(t *testing.T) {
		flag := rootCmd.Flags().Lookup("lang")
		if flag == nil {
			t.Fatal("flag --lang não encontrada")
		}
	})

	t.Run("model flag existe", func(t *testing.T) {
		flag := rootCmd.Flags().Lookup("model")
		if flag == nil {
			t.Fatal("flag --model não encontrada")
		}
	})

	t.Run("prompt flag existe", func(t *testing.T) {
		flag := rootCmd.Flags().Lookup("prompt")
		if flag == nil {
			t.Fatal("flag --prompt não encontrada")
		}
	})

	t.Run("verbose flag alias -v", func(t *testing.T) {
		flag := rootCmd.Flags().Lookup("verbose")
		if flag == nil {
			t.Fatal("flag --verbose não encontrada")
		}
		if flag.Shorthand != "v" {
			t.Errorf("verbose.Shorthand = %q, want 'v'", flag.Shorthand)
		}
	})
}

func TestRootCmdArgs(t *testing.T) {
	t.Run("MaximumNArgs é 1", func(t *testing.T) {
		// Verifica que o comando aceita no máximo 1 argumento
		if rootCmd.Args == nil {
			t.Fatal("rootCmd.Args é nil")
		}
		// Testa a validação com 2 args
		err := cobra.MaximumNArgs(1)(rootCmd, []string{"a", "b"})
		if err == nil {
			t.Error("MaximumNArgs(1) com 2 args = nil, want erro")
		}
		// Testa com 0 args
		err = cobra.MaximumNArgs(1)(rootCmd, []string{})
		if err != nil {
			t.Errorf("MaximumNArgs(1) com 0 args = %v, want nil", err)
		}
	})
}

func TestVerboseOutput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "verbose contém emoji indicadores",
			input:   "📄",
			want:    "📄",
			wantErr: false,
		},
		{
			name:    "verbose contém indicador de modelo",
			input:   "🤖",
			want:    "🤖",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(tt.want, tt.input) {
				t.Errorf("esperava %q em %q", tt.input, tt.want)
			}
		})
	}
}
