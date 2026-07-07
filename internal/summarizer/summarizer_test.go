package summarizer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestClient é um helper para criar um Client nos testes.
// Panica se houver erro na criação (já que testes devem sempre usar configs válidas).
func newTestClient(t *testing.T, apiKey, baseURL, model string, timeout time.Duration) *Client {
	t.Helper()
	s, err := New(Config{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
		Timeout: timeout,
	})
	if err != nil {
		t.Fatalf("New() erro inesperado: %v", err)
	}
	return s
}

func TestNew(t *testing.T) {
	t.Run("config válida", func(t *testing.T) {
		s, err := New(Config{
			APIKey:  "sk-test",
			BaseURL: "https://api.example.com/v1",
			Model:   "test-model",
			Timeout: 10 * time.Second,
		})
		if err != nil {
			t.Fatalf("New() erro inesperado: %v", err)
		}
		if s == nil {
			t.Fatal("New() retornou nil")
		}
		if s.model != "test-model" {
			t.Errorf("model = %q, want %q", s.model, "test-model")
		}
		if string(s.apiKey) != "sk-test" {
			t.Errorf("apiKey = %q, want %q", string(s.apiKey), "sk-test")
		}
	})

	t.Run("API key vazia", func(t *testing.T) {
		_, err := New(Config{
			APIKey:  "",
			BaseURL: "https://api.example.com/v1",
			Model:   "test-model",
		})
		if err == nil {
			t.Fatal("New() com API key vazia = nil, want erro")
		}
	})

	t.Run("modelo vazio", func(t *testing.T) {
		_, err := New(Config{
			APIKey:  "sk-test",
			BaseURL: "https://api.example.com/v1",
			Model:   "",
		})
		if err == nil {
			t.Fatal("New() com modelo vazio = nil, want erro")
		}
	})

	t.Run("base URL vazia", func(t *testing.T) {
		_, err := New(Config{
			APIKey:  "sk-test",
			BaseURL: "",
			Model:   "test-model",
		})
		if err == nil {
			t.Fatal("New() com base URL vazia = nil, want erro")
		}
	})

	t.Run("timeout zero usa padrão", func(t *testing.T) {
		s, err := New(Config{
			APIKey:  "sk-test",
			BaseURL: "https://api.example.com/v1",
			Model:   "test-model",
			Timeout: 0,
		})
		if err != nil {
			t.Fatalf("New() erro inesperado: %v", err)
		}
		if s == nil {
			t.Fatal("New() retornou nil")
		}
	})
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

		s := newTestClient(t, "sk-test", server.URL, "test-model", 5*time.Second)

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

		s := newTestClient(t, "sk-test", server.URL, "test-model", 5*time.Second)

		_, err := s.Summarize(context.Background(), "Sistema", "Texto")
		if err == nil {
			t.Fatal("Summarize() com choices vazio = nil, want erro")
		}
		if !strings.Contains(err.Error(), "vazia") {
			t.Errorf("erro = %q, want contendo 'vazia'", err.Error())
		}
	})

	t.Run("finish_reason = length retorna conteúdo parcial", func(t *testing.T) {
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

		s := newTestClient(t, "sk-test", server.URL, "test-model", 5*time.Second)

		result, err := s.Summarize(context.Background(), "Sistema", "Texto")
		if err == nil {
			t.Fatal("Summarize() com length = nil, want erro")
		}
		if !errors.Is(err, ErrTruncated) {
			t.Errorf("errors.Is(err, ErrTruncated) = false, want true")
		}
		if result != "Resumo parcial..." {
			t.Errorf("Summarize() = %q, want %q", result, "Resumo parcial...")
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

		s := newTestClient(t, "sk-test", server.URL, "test-model", 5*time.Second)

		_, err := s.Summarize(context.Background(), "Sistema", "Texto")
		if err == nil {
			t.Fatal("Summarize() com content_filter = nil, want erro")
		}
		if !strings.Contains(err.Error(), "bloqueado") {
			t.Errorf("erro = %q, want contendo 'bloqueado'", err.Error())
		}
	})

	t.Run("finish_reason = stop com conteúdo vazio", func(t *testing.T) {
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
					"finish_reason": "stop"
				}]
			}`)
		}))
		defer server.Close()

		s := newTestClient(t, "sk-test", server.URL, "test-model", 5*time.Second)

		_, err := s.Summarize(context.Background(), "Sistema", "Texto")
		if err == nil {
			t.Fatal("Summarize() com stop e conteúdo vazio = nil, want erro")
		}
		if !strings.Contains(err.Error(), "vazio") {
			t.Errorf("erro = %q, want contendo 'vazio'", err.Error())
		}
	})

	t.Run("contexto cancelado", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Aguarda o cancelamento do contexto ao invés de bloquear para sempre
			<-r.Context().Done()
		}))
		defer server.Close()

		s := newTestClient(t, "sk-test", server.URL, "test-model", 5*time.Second)

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

		s := newTestClient(t, "sk-test", server.URL, "test-model", 5*time.Second)

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

		s := newTestClient(t, "sk-test", server.URL, "test-model", 5*time.Second)

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

		s := newTestClient(t, "sk-test", server.URL, "test-model", 5*time.Second)

		_, err := s.Summarize(context.Background(), "Sistema", "Texto")
		if err == nil {
			t.Fatal("Summarize() com 500 = nil, want erro")
		}
		if !strings.Contains(err.Error(), "indisponível") {
			t.Errorf("erro = %q, want 'indisponível'", err.Error())
		}
	})

	t.Run("erro 400 da API (default)", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{
				"error": {
					"message": "Invalid request parameters",
					"type": "invalid_request_error"
				}
			}`)
		}))
		defer server.Close()

		s := newTestClient(t, "sk-test", server.URL, "test-model", 5*time.Second)

		_, err := s.Summarize(context.Background(), "Sistema", "Texto")
		if err == nil {
			t.Fatal("Summarize() com 400 = nil, want erro")
		}
		if !strings.Contains(err.Error(), "400") {
			t.Errorf("erro = %q, want contendo '400'", err.Error())
		}
	})
}

func TestClassifyAPIErrorSanitization(t *testing.T) {
	// Cria um client com chave conhecida para testar redação da própria chave
	s := newTestClient(t, "sk-my-secret-key-12345", "https://api.example.com/v1", "test-model", 5*time.Second)

	t.Run("chave sk- no erro", func(t *testing.T) {
		err := s.classifyAPIError(fmt.Errorf("timeout with key sk-abcdefghijklmnopqrstuvwxyz123456"), s.apiKey)
		if strings.Contains(err.Error(), "sk-abcdefghijklmnopqrstuvwxyz") {
			t.Errorf("chave sk- não deveria aparecer: %q", err.Error())
		}
		if !strings.Contains(err.Error(), "***REDACTED***") {
			t.Errorf("erro deveria conter REDACTED: %q", err.Error())
		}
	})

	t.Run("chave sk-proj- no erro", func(t *testing.T) {
		err := s.classifyAPIError(fmt.Errorf("error: sk-proj-abcdefghijklmnopqrstuvwxyz123456"), s.apiKey)
		if strings.Contains(err.Error(), "sk-proj-") {
			t.Errorf("chave sk-proj- não deveria aparecer: %q", err.Error())
		}
	})

	t.Run("api_key= no erro", func(t *testing.T) {
		err := s.classifyAPIError(fmt.Errorf("invalid api_key=sk-test-key-here-12345"), s.apiKey)
		if strings.Contains(err.Error(), "sk-test-key") {
			t.Errorf("api_key não deveria aparecer: %q", err.Error())
		}
	})

	t.Run("token= no erro", func(t *testing.T) {
		err := s.classifyAPIError(fmt.Errorf("invalid token=ghp_12345678901234567890"), s.apiKey)
		if strings.Contains(err.Error(), "ghp_123456") {
			t.Errorf("token não deveria aparecer: %q", err.Error())
		}
	})

	t.Run("chave da própria API no erro", func(t *testing.T) {
		// A chave configurada no client é "sk-my-secret-key-12345"
		err := s.classifyAPIError(fmt.Errorf("error: authentication failed for key 'sk-my-secret-key-12345'"), s.apiKey)
		if strings.Contains(err.Error(), "sk-my-secret-key-12345") {
			t.Errorf("chave configurada não deveria aparecer: %q", err.Error())
		}
		if !strings.Contains(err.Error(), "***REDACTED***") {
			t.Errorf("erro deveria conter REDACTED: %q", err.Error())
		}
	})

	t.Run("erro de rede comum", func(t *testing.T) {
		err := s.classifyAPIError(fmt.Errorf("connection refused"), s.apiKey)
		if err == nil {
			t.Fatal("classifyAPIError(nil-like) = nil, want erro")
		}
		if !strings.Contains(err.Error(), "erro na chamada da API") {
			t.Errorf("erro = %q, want 'erro na chamada da API'", err.Error())
		}
	})

	t.Run("erro de timeout preserva cadeia via errors.Is", func(t *testing.T) {
		original := fmt.Errorf("connection timeout: %w", context.DeadlineExceeded)
		err := s.classifyAPIError(original, s.apiKey)
		// A mensagem exibida ao usuário (Error()) deve ser a versão tratada
		if !strings.Contains(err.Error(), "tempo limite") {
			t.Errorf("mensagem deveria conter 'tempo limite': %q", err.Error())
		}
		// A cadeia deve ser preservada para errors.Is
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Error("errors.Is(err, context.DeadlineExceeded) = false, want true")
		}
	})

	t.Run("ErrTimeout detectável via errors.Is", func(t *testing.T) {
		original := fmt.Errorf("request timeout: %w", context.DeadlineExceeded)
		err := s.classifyAPIError(original, s.apiKey)
		if !errors.Is(err, ErrTimeout) {
			t.Error("errors.Is(err, ErrTimeout) = false, want true")
		}
	})
}

func TestClientClear(t *testing.T) {
	t.Run("Clear zera apiKey e marca como limpo", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"id":"x","object":"chat.completion","created":0,"model":"m","choices":[]}`)
		}))
		defer server.Close()

		s := newTestClient(t, "sk-secret-to-clear", server.URL, "test-model", 5*time.Second)

		// Verifica estado inicial
		if s.cleared {
			t.Error("cleared = true antes de Clear()")
		}
		if string(s.apiKey) != "sk-secret-to-clear" {
			t.Errorf("apiKey = %q, want %q", string(s.apiKey), "sk-secret-to-clear")
		}

		s.Clear()

		// Verifica que o slice foi zerado byte a byte
		if s.apiKey != nil {
			t.Error("apiKey não foi setado para nil após Clear()")
		}
		if !s.cleared {
			t.Error("cleared = false após Clear(), want true")
		}
	})

	t.Run("Summarize após Clear retorna erro", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{}`)
		}))
		defer server.Close()

		s := newTestClient(t, "sk-test", server.URL, "test-model", 5*time.Second)
		s.Clear()

		_, err := s.Summarize(context.Background(), "Sistema", "Texto")
		if err == nil {
			t.Error("Summarize() após Clear() deveria retornar erro, mas retornou nil")
		}
		if !strings.Contains(err.Error(), "usado após Clear") {
			t.Errorf("erro = %q, want contendo 'usado após Clear'", err.Error())
		}
	})

	t.Run("Clear duplo não panica", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{}`)
		}))
		defer server.Close()

		s := newTestClient(t, "sk-test", server.URL, "test-model", 5*time.Second)
		s.Clear()
		s.Clear() // não deve panicar
		if !s.cleared {
			t.Error("cleared = false após double Clear()")
		}
	})
}

func TestRedactCredentials(t *testing.T) {
	t.Run("redige chave configurada", func(t *testing.T) {
		result := redactCredentials("my-api-key-12345 is secret", []byte("my-api-key-12345"))
		if strings.Contains(result, "my-api-key-12345") {
			t.Errorf("chave não redigida: %q", result)
		}
		if !strings.Contains(result, "***REDACTED***") {
			t.Errorf("resultado não contém REDACTED: %q", result)
		}
	})

	t.Run("apiKey vazia não quebra", func(t *testing.T) {
		result := redactCredentials("some error message", nil)
		if result != "some error message" {
			t.Errorf("resultado inesperado: %q", result)
		}
	})

	t.Run("string vazia não quebra", func(t *testing.T) {
		result := redactCredentials("", []byte("sk-test-key"))
		if result != "" {
			t.Errorf("resultado inesperado: %q", result)
		}
	})

	t.Run("slice vazio não quebra", func(t *testing.T) {
		result := redactCredentials("some error message", []byte{})
		if result != "some error message" {
			t.Errorf("resultado inesperado: %q", result)
		}
	})
}
