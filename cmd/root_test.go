package cmd

import (
	"strings"
	"testing"
)

func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		name         string
		lang         string
		customPrompt string
		wantContains string
	}{
		{
			name:         "default prompt",
			lang:         "pt-br",
			customPrompt: "",
			wantContains: "Summarize the following text in pt-br",
		},
		{
			name:         "custom prompt with lang",
			lang:         "en",
			customPrompt: "Resuma para um leigo no assunto",
			wantContains: "Responda em en.",
		},
		{
			name:         "custom prompt includes original",
			lang:         "es",
			customPrompt: "Haz un resumen",
			wantContains: "Haz un resumen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPrompt(tt.lang, tt.customPrompt)
			if !strings.Contains(result, tt.wantContains) {
				t.Errorf("buildPrompt(%q, %q) = %q, want containing %q", tt.lang, tt.customPrompt, result, tt.wantContains)
			}
		})
	}
}
