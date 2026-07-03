package cmd

import (
	"testing"
)

func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		name         string
		lang         string
		customPrompt string
		want         string
	}{
		{
			name:         "default prompt en",
			lang:         "en",
			customPrompt: "",
			want:         "Summarize the following text in en. Be concise but capture all key points.",
		},
		{
			name:         "default prompt pt-br",
			lang:         "pt-br",
			customPrompt: "",
			want:         "Resuma o texto a seguir em pt-br. Seja conciso mas capture todos os pontos-chave.",
		},
		{
			name:         "default prompt pt",
			lang:         "pt",
			customPrompt: "",
			want:         "Resuma o texto a seguir em pt. Seja conciso mas capture todos os pontos-chave.",
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
