package summarizer

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	s := New(Config{
		APIKey:  "sk-test",
		BaseURL: "https://api.example.com/v1",
		Model:   "test-model",
		Timeout: 10 * time.Second,
	})
	if s == nil {
		t.Fatal("New() retornou nil")
	}
	if s.model != "test-model" {
		t.Errorf("model = %q, want %q", s.model, "test-model")
	}
}

func TestNewDefaultTimeout(t *testing.T) {
	// Timeout zero deve usar o padrão de 30s
	s := New(Config{
		APIKey:  "sk-test",
		BaseURL: "https://api.example.com/v1",
		Model:   "test-model",
		Timeout: 0,
	})
	if s == nil {
		t.Fatal("New() retornou nil")
	}
}

func TestSummarize(t *testing.T) {
	t.Run("resposta com sucesso", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{
				"id": "chatcmpl-test",
				"object": "chat.completion",
				"created": 1234567890,
				"model": "test-model",
				"choices": [{
					"index": 0,
					"message": {
						"role": "assistant",
						"content": "Este é o resumo."
					},
					"finish_reason": "stop"
				}]
			}`)
		}))
		defer server.Close()

		s := New(Config{
			APIKey:  "sk-test",
			BaseURL: server.URL,
			Model:   "test-model",
			Timeout: 5 * time.Second,
		})

		result, err := s.Summarize(context.Background(), "Sistema", "Texto do usuário")
		if err != nil {
			t.Fatalf("Summarize() erro inesperado: %v", err)
		}
		if result != "Este é o resumo." {
			t.Errorf("Summarize() = %q, want %q", result, "Este é o resumo.")
		}
	})

	t.Run("choices vazio", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{
				"id": "chatcmpl-test",
				"object": "chat.completion",
				"created": 1234567890,
				"model": "test-model",
				"choices": []
			}`)
		}))
		defer server.Close()

		s := New(Config{
			APIKey:  "sk-test",
			BaseURL: server.URL,
			Model:   "test-model",
			Timeout: 5 * time.Second,
		})

		_, err := s.Summarize(context.Background(), "Sistema", "Texto")
		if err == nil {
			t.Fatal("Summarize() com choices vazio = nil, want erro")
		}
		if !strings.Contains(err.Error(), "vazia") {
			t.Errorf("erro = %q, want contendo 'vazia'", err.Error())
		}
	})

	t.Run("finish_reason = length", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{
				"id": "chatcmpl-test",
				"object": "chat.completion",
				"created": 1234567890,
				"model": "test-model",
				"choices": [{
					"index": 0,
					"message": {
						"role": "assistant",
						"content": "Resumo parcial..."
					},
					"finish_reason": "length"
				}]
			}`)
		}))
		defer server.Close()

		s := New(Config{
			APIKey:  "sk-test",
			BaseURL: server.URL,
			Model:   "test-model",
			Timeout: 5 * time.Second,
		})

		_, err := s.Summarize(context.Background(), "Sistema", "Texto")
		if err == nil {
			t.Fatal("Summarize() com length = nil, want erro")
		}
		if !strings.Contains(err.Error(), "truncado") {
			t.Errorf("erro = %q, want contendo 'truncado'", err.Error())
		}
	})

	t.Run("finish_reason = content_filter", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{
				"id": "chatcmpl-test",
				"object": "chat.completion",
				"created": 1234567890,
				"model": "test-model",
				"choices": [{
					"index": 0,
					"message": {
						"role": "assistant",
						"content": ""
					},
					"finish_reason": "content_filter"
				}]
			}`)
		}))
		defer server.Close()

		s := New(Config{
			APIKey:  "sk-test",
			BaseURL: server.URL,
			Model:   "test-model",
			Timeout: 5 * time.Second,
		})

		_, err := s.Summarize(context.Background(), "Sistema", "Texto")
		if err == nil {
			t.Fatal("Summarize() com content_filter = nil, want erro")
		}
		if !strings.Contains(err.Error(), "bloqueado") {
			t.Errorf("erro = %q, want contendo 'bloqueado'", err.Error())
		}
	})

	t.Run("contexto cancelado", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Aguarda o cancelamento do contexto ao invés de bloquear para sempre
			<-r.Context().Done()
		}))
		defer server.Close()

		s := New(Config{
			APIKey:  "sk-test",
			BaseURL: server.URL,
			Model:   "test-model",
			Timeout: 5 * time.Second,
		})

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancela imediatamente

		_, err := s.Summarize(ctx, "Sistema", "Texto")
		if err == nil {
			t.Fatal("Summarize() com contexto cancelado = nil, want erro")
		}
	})

	t.Run("erro 401 da API", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{
				"error": {
					"message": "Incorrect API key provided",
					"type": "authentication_error",
					"code": "invalid_api_key"
				}
			}`)
		}))
		defer server.Close()

		s := New(Config{
			APIKey:  "sk-test",
			BaseURL: server.URL,
			Model:   "test-model",
			Timeout: 5 * time.Second,
		})

		_, err := s.Summarize(context.Background(), "Sistema", "Texto")
		if err == nil {
			t.Fatal("Summarize() com 401 = nil, want erro")
		}
		if !strings.Contains(err.Error(), "credenciais") {
			t.Errorf("erro = %q, want 'credenciais'", err.Error())
		}
	})

	t.Run("erro 429 da API", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `{
				"error": {
					"message": "Rate limit exceeded",
					"type": "rate_limit_error"
				}
			}`)
		}))
		defer server.Close()

		s := New(Config{
			APIKey:  "sk-test",
			BaseURL: server.URL,
			Model:   "test-model",
			Timeout: 5 * time.Second,
		})

		_, err := s.Summarize(context.Background(), "Sistema", "Texto")
		if err == nil {
			t.Fatal("Summarize() com 429 = nil, want erro")
		}
		if !strings.Contains(err.Error(), "limite") {
			t.Errorf("erro = %q, want 'limite'", err.Error())
		}
	})

	t.Run("erro 500 da API", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{
				"error": {
					"message": "Internal server error",
					"type": "server_error"
				}
			}`)
		}))
		defer server.Close()

		s := New(Config{
			APIKey:  "sk-test",
			BaseURL: server.URL,
			Model:   "test-model",
			Timeout: 5 * time.Second,
		})

		_, err := s.Summarize(context.Background(), "Sistema", "Texto")
		if err == nil {
			t.Fatal("Summarize() com 500 = nil, want erro")
		}
		if !strings.Contains(err.Error(), "indisponível") {
			t.Errorf("erro = %q, want 'indisponível'", err.Error())
		}
	})
}

func TestClassifyAPIErrorSanitization(t *testing.T) {
	// Testa a função classifyAPIError diretamente com erros que têm
	// padrões de chave no texto (caminho do erro não-API)

	t.Run("chave sk- no erro", func(t *testing.T) {
		err := classifyAPIError(fmt.Errorf("timeout with key sk-abcdefghijklmnopqrstuvwxyz123456"))
		if strings.Contains(err.Error(), "sk-abcdefghijklmnopqrstuvwxyz") {
			t.Errorf("chave sk- não deveria aparecer: %q", err.Error())
		}
		if !strings.Contains(err.Error(), "***REDACTED***") {
			t.Errorf("erro deveria conter REDACTED: %q", err.Error())
		}
	})

	t.Run("chave sk-proj- no erro", func(t *testing.T) {
		err := classifyAPIError(fmt.Errorf("error: sk-proj-abcdefghijklmnopqrstuvwxyz123456"))
		if strings.Contains(err.Error(), "sk-proj-") {
			t.Errorf("chave sk-proj- não deveria aparecer: %q", err.Error())
		}
	})

	t.Run("api_key= no erro", func(t *testing.T) {
		err := classifyAPIError(fmt.Errorf("invalid api_key=sk-test-key-here-12345"))
		if strings.Contains(err.Error(), "sk-test-key") {
			t.Errorf("api_key não deveria aparecer: %q", err.Error())
		}
	})

	t.Run("token= no erro", func(t *testing.T) {
		err := classifyAPIError(fmt.Errorf("invalid token=ghp_12345678901234567890"))
		if strings.Contains(err.Error(), "ghp_123456") {
			t.Errorf("token não deveria aparecer: %q", err.Error())
		}
	})

	t.Run("erro de rede comum", func(t *testing.T) {
		err := classifyAPIError(fmt.Errorf("connection refused"))
		if err == nil {
			t.Fatal("classifyAPIError(nil-like) = nil, want erro")
		}
		if !strings.Contains(err.Error(), "erro na chamada da API") {
			t.Errorf("erro = %q, want 'erro na chamada da API'", err.Error())
		}
	})
}
