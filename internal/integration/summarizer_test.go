//go:build integration

// Package integration_test contém testes de integração com a API real.
//
// Estes testes compilam apenas com a tag "integration" ativa, evitando
// que sejam executados acidentalmente durante go test ./... sem a chave de API.
//
// Uso:
//
//	TLDR_API_KEY=sk-... go test -tags=integration -v ./internal/integration/
package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Elissdev/tl-dr/internal/summarizer"
)

// TestSummarizeRealAPI é um teste de integração que chama a API real.
// Só executa se a variável de ambiente TLDR_API_KEY estiver definida
// E a tag de build "integration" estiver ativa.
//
// Uso:
//
//	TLDR_API_KEY=sk-... go test -tags=integration -run TestSummarizeRealAPI ./internal/integration/
func TestSummarizeRealAPI(t *testing.T) {
	apiKey := os.Getenv("TLDR_API_KEY")
	if apiKey == "" {
		t.Skip("TLDR_API_KEY não definida — pulando teste de integração com API real")
	}

	baseURL := os.Getenv("TLDR_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.apiario.dev/v1"
	}

	model := os.Getenv("TLDR_DEFAULT_MODEL")
	if model == "" {
		model = "deepseek/deepseek-v4-flash"
	}

	s, err := summarizer.New(summarizer.Config{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
		Timeout: 60 * time.Second,
	})
	if err != nil {
		t.Fatalf("summarizer.New() erro: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := s.Summarize(ctx,
		"Summarize the following text in pt-br. Be concise but capture all key points.",
		"O tl;dr é uma ferramenta de linha de comando que recebe um texto "+
			"(de arquivo ou stdin) e produz um resumo conciso no idioma especificado. "+
			"Utiliza uma API compatível com a OpenAI para gerar os resumos.")
	if err != nil {
		t.Fatalf("Summarize() erro na API real: %v", err)
	}

	if result == "" {
		t.Fatal("Summarize() retornou resumo vazio")
	}

	t.Logf("Resumo gerado (%d caracteres): %s", len(result), result)
}
