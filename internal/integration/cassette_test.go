//go:build integration

// Package integration_test contém testes de integração com a API real,
// gravados e reproduzidos via go-vcr (cassete).
//
// Modos de operação:
//   - Replay (padrão): usa cassetes gravados. Funciona offline.
//   - Record: TLDR_CASSETE_MODE=record + TLDR_API_KEY definida
//
// Uso:
//
//	# Replay (padrão, offline)
//	go test -tags=integration -v ./internal/integration/
//
//	# Record (requer chave de API)
//	TLDR_API_KEY=sk-... TLDR_CASSETE_MODE=record go test -tags=integration -v ./internal/integration/
package integration_test

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/dnaeon/go-vcr.v3/recorder"

	"github.com/Elissdev/tl-dr/internal/summarizer"
)

// cassetteDir é o diretório onde os cassetes são armazenados.
var cassetteDir = filepath.Join("testdata", "cassettes")

// newRecorder cria um novo recorder go-vcr para o cassete especificado.
// Em modo record (TLDR_CASSETE_MODE=record), grava uma nova interação.
// Caso contrário, reproduz do cassete existente.
// O caller deve chamar Stop() no recorder ao finalizar.
func newRecorder(t *testing.T, cassetteName string) *recorder.Recorder {
	t.Helper()

	cassettePath := filepath.Join(cassetteDir, cassetteName)

	// Verifica o modo de operação
	mode := os.Getenv("TLDR_CASSETE_MODE")
	isRecord := strings.EqualFold(mode, "record")

	var rec *recorder.Recorder
	var err error

	if isRecord {
		// Modo record: cria ou sobrescreve o cassete
		rec, err = recorder.NewWithOptions(&recorder.Options{
			CassetteName:       cassettePath,
			Mode:               recorder.ModeRecordOnly,
			RealTransport:      http.DefaultTransport,
			SkipRequestLatency: false,
		})
		if err != nil {
			t.Fatalf("erro ao criar recorder (record): %v", err)
		}
		t.Logf("📼 Gravando cassete: %s", cassettePath)
	} else {
		// Modo replay: reproduz do cassete existente
		rec, err = recorder.NewWithOptions(&recorder.Options{
			CassetteName:       cassettePath,
			Mode:               recorder.ModeReplayOnly,
			RealTransport:      http.DefaultTransport,
			SkipRequestLatency: true,
		})
		if err != nil {
			t.Fatalf("erro ao criar recorder (replay): %v", err)
		}
		t.Logf("📼 Reproduzindo cassete: %s", cassettePath)
	}

	return rec
}

// getTestAPIKey retorna a chave de API para o teste.
// Em modo record, usa a variável de ambiente.
// Em modo replay, usa uma chave fictícia (não fará chamadas reais).
func getTestAPIKey(t *testing.T) string {
	t.Helper()

	mode := os.Getenv("TLDR_CASSETE_MODE")
	if strings.EqualFold(mode, "record") {
		key := os.Getenv("TLDR_API_KEY")
		if key == "" {
			t.Fatal("TLDR_API_KEY é obrigatória em modo record")
		}
		return key
	}

	// Em modo replay, a chave não é usada (interações são simuladas)
	return "sk-replay-key-not-used"
}

// newSummarizerWithCassette cria um summarizer.Client configurado com o
// recorder go-vcr para o cassete especificado.
func newSummarizerWithCassette(t *testing.T, cassetteName, apiKey string) *summarizer.Client {
	t.Helper()

	rec := newRecorder(t, cassetteName)
	t.Cleanup(func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("recorder.Stop() erro: %v", err)
		}
	})

	httpClient := rec.GetDefaultClient()

	s, err := summarizer.New(summarizer.Config{
		APIKey:     apiKey,
		BaseURL:    "https://api.apiario.dev/v1",
		Model:      "deepseek/deepseek-v4-flash",
		Timeout:    30 * time.Second,
		HTTPClient: httpClient,
	})
	if err != nil {
		t.Fatalf("summarizer.New() erro: %v", err)
	}

	return s
}

func TestSummarizeWithCassette(t *testing.T) {
	if testing.Short() {
		t.Skip("pulado em modo short")
	}

	apiKey := getTestAPIKey(t)
	s := newSummarizerWithCassette(t, "summarize_success", apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.Summarize(ctx,
		"Summarize the following text in pt-br. Be concise but capture all key points.",
		"O tl;dr é uma ferramenta de linha de comando que recebe um texto "+
			"(de arquivo ou stdin) e produz um resumo conciso no idioma especificado. "+
			"Utiliza uma API compatível com a OpenAI para gerar os resumos.")
	if err != nil {
		t.Fatalf("Summarize() erro: %v", err)
	}

	if result == "" {
		t.Fatal("Summarize() retornou resumo vazio")
	}

	t.Logf("Resumo gerado (%d caracteres): %s", len(result), result)
}

func TestSummarizeWithCassetteShortText(t *testing.T) {
	if testing.Short() {
		t.Skip("pulado em modo short")
	}

	apiKey := getTestAPIKey(t)
	s := newSummarizerWithCassette(t, "summarize_short", apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := s.Summarize(ctx,
		"Summarize the following text in en. Be concise.",
		"Go is a statically typed, compiled programming language designed at Google.")
	if err != nil {
		t.Fatalf("Summarize() erro: %v", err)
	}

	if result == "" {
		t.Fatal("Summarize() retornou resumo vazio")
	}

	t.Logf("Resumo gerado (%d caracteres): %s", len(result), result)
}

func TestSummarizeAuthError(t *testing.T) {
	if testing.Short() {
		t.Skip("pulado em modo short")
	}

	s := newSummarizerWithCassette(t, "summarize_unauthorized", "sk-replay-invalid")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := s.Summarize(ctx, "System prompt", "User text")
	if err == nil {
		t.Fatal("Summarize() com credenciais inválidas = nil, want erro")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "credenciais") {
		t.Errorf("erro = %q, want contendo 'credenciais'", errMsg)
	}
	t.Logf("Erro de autenticação capturado: %s", errMsg)
}

func TestSummarizeRateLimitError(t *testing.T) {
	if testing.Short() {
		t.Skip("pulado em modo short")
	}

	s := newSummarizerWithCassette(t, "summarize_rate_limited", "sk-replay-rate-limited")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := s.Summarize(ctx, "System prompt", "User text")
	if err == nil {
		t.Fatal("Summarize() com rate limit = nil, want erro")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "limite") {
		t.Errorf("erro = %q, want contendo 'limite'", errMsg)
	}
	t.Logf("Erro de rate limit capturado: %s", errMsg)
}

func TestSummarizeServerError(t *testing.T) {
	if testing.Short() {
		t.Skip("pulado em modo short")
	}

	s := newSummarizerWithCassette(t, "summarize_server_error", "sk-replay-server-error")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := s.Summarize(ctx, "System prompt", "User text")
	if err == nil {
		t.Fatal("Summarize() com erro 500 = nil, want erro")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "indisponível") {
		t.Errorf("erro = %q, want contendo 'indisponível'", errMsg)
	}
	t.Logf("Erro de servidor capturado: %s", errMsg)
}

func TestSummarizeContextCanceled(t *testing.T) {
	if testing.Short() {
		t.Skip("pulado em modo short")
	}

	apiKey := getTestAPIKey(t)
	s := newSummarizerWithCassette(t, "summarize_success", apiKey)

	// Contexto cancelado imediatamente — a chamada HTTP nem chega a ser feita
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := s.Summarize(ctx, "System prompt", "User text")
	if err == nil {
		t.Fatal("Summarize() com contexto cancelado = nil, want erro")
	}

	if errors.Is(err, summarizer.ErrTimeout) {
		t.Logf("Erro de timeout detectado via errors.Is: %s", err)
	}
	t.Logf("Erro de contexto capturado: %s", err)
}

func TestSummarizeContextDeadline(t *testing.T) {
	if testing.Short() {
		t.Skip("pulado em modo short")
	}

	apiKey := getTestAPIKey(t)
	s := newSummarizerWithCassette(t, "summarize_success", apiKey)

	// Contexto com deadline já expirado
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Second))
	defer cancel()

	// Dá um tempo para o deadline propagar
	time.Sleep(10 * time.Millisecond)

	_, err := s.Summarize(ctx, "System prompt", "User text")
	if err == nil {
		t.Fatal("Summarize() com deadline expirado = nil, want erro")
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, summarizer.ErrTimeout) {
		t.Logf("Erro de deadline detectado: %s", err)
	} else {
		t.Logf("Erro capturado: %v", err)
	}
}
