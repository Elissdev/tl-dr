package summarizer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

// apiKeyPattern detecta possíveis chaves de API em mensagens de erro para
// fazer redação antes de exibir ao usuário.
// Cobre os formatos:
//   - sk-... (OpenAI legado)
//   - sk-proj-... (OpenAI novo formato)
//   - api[_-]?key=... (qualquer provider)
//   - token=... (qualquer provider)
var apiKeyPattern = regexp.MustCompile(`(?i)(sk-(proj-)?[a-z0-9]{20,}|api[_-]?key[=:]\s*\S{8,}|token[=:]\s*\S{8,})`)

// Config define a configuração para o summarizer.
type Config struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
}

// Summarizer gerencia a comunicação com a API.
type Summarizer struct {
	client *openai.Client
	model  string
}

// New cria um novo Summarizer com a configuração fornecida.
func New(cfg Config) *Summarizer {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	client := openai.NewClient(
		option.WithAPIKey(cfg.APIKey),
		option.WithBaseURL(cfg.BaseURL),
		option.WithHTTPClient(&http.Client{Timeout: timeout}),
		option.WithMaxRetries(0), // sem retry automático — tratamos erros no classifyAPIError
	)
	return &Summarizer{
		client: &client,
		model:  cfg.Model,
	}
}

// Summarize envia um prompt e texto para a API e retorna o resumo.
// O prompt é enviado como mensagem de sistema (reduz risco de injeção de prompt)
// e o texto do usuário como mensagem de usuário.
func (s *Summarizer) Summarize(ctx context.Context, systemPrompt, userText string) (string, error) {
	chat, err := s.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			{
				OfSystem: &openai.ChatCompletionSystemMessageParam{
					Content: openai.ChatCompletionSystemMessageParamContentUnion{
						OfString: openai.String(systemPrompt),
					},
				},
			},
			{
				OfUser: &openai.ChatCompletionUserMessageParam{
					Content: openai.ChatCompletionUserMessageParamContentUnion{
						OfString: openai.String(userText),
					},
				},
			},
		},
		Model: shared.ChatModel(s.model),
	})
	if err != nil {
		return "", classifyAPIError(err)
	}

	return extractContent(chat)
}

// SummarizeStream envia um prompt e texto para a API e retorna um canal
// que recebe chunks do resumo em tempo real (streaming).
// O canal é fechado quando o streaming termina ou ocorre um erro.
func (s *Summarizer) SummarizeStream(ctx context.Context, systemPrompt, userText string) (<-chan StreamChunk, error) {
	stream := s.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			{
				OfSystem: &openai.ChatCompletionSystemMessageParam{
					Content: openai.ChatCompletionSystemMessageParamContentUnion{
						OfString: openai.String(systemPrompt),
					},
				},
			},
			{
				OfUser: &openai.ChatCompletionUserMessageParam{
					Content: openai.ChatCompletionUserMessageParamContentUnion{
						OfString: openai.String(userText),
					},
				},
			},
		},
		Model: shared.ChatModel(s.model),
	})

	ch := make(chan StreamChunk, 64)

	go func() {
		defer close(ch)
		defer stream.Close()

		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) == 0 {
				continue
			}

			choice := chunk.Choices[0]

			// Verifica finish_reason em cada chunk
			if choice.FinishReason == "length" {
				ch <- StreamChunk{Err: fmt.Errorf("resumo truncado: o modelo atingiu o limite de tokens")}
				return
			}
			if choice.FinishReason == "content_filter" {
				ch <- StreamChunk{Err: fmt.Errorf("resumo bloqueado pelo filtro de conteúdo da API")}
				return
			}

			ch <- StreamChunk{Text: choice.Delta.Content}
		}

		if err := stream.Err(); err != nil {
			ch <- StreamChunk{Err: classifyAPIError(err)}
		}
	}()

	return ch, nil
}

// StreamChunk representa um pedaço do resumo recebido via streaming.
type StreamChunk struct {
	Text string
	Err  error
}

// extractContent extrai o conteúdo da resposta da API.
func extractContent(chat *openai.ChatCompletion) (string, error) {
	if len(chat.Choices) == 0 {
		return "", fmt.Errorf("API retornou resposta vazia")
	}

	choice := chat.Choices[0]
	if choice.FinishReason == "length" {
		return "", fmt.Errorf("resumo truncado: o modelo atingiu o limite de tokens")
	}
	if choice.FinishReason == "content_filter" {
		return "", fmt.Errorf("resumo bloqueado pelo filtro de conteúdo da API")
	}

	return choice.Message.Content, nil
}

// isContextLengthError verifica se a mensagem de erro indica que o texto
// excedeu o limite de contexto do modelo.
func isContextLengthError(msg string) bool {
	keywords := []string{
		"context length",
		"maximum context",
		"token limit",
		"context_window",
		"context_length_exceeded",
		"too many tokens",
		"request too large",
		"max_tokens",
	}
	msgLower := strings.ToLower(msg)
	for _, kw := range keywords {
		if strings.Contains(msgLower, kw) {
			return true
		}
	}
	return false
}

// classifyAPIError mapeia erros da API para mensagens mais amigáveis e
// remove qualquer vazamento acidental de credenciais ou tokens.
func classifyAPIError(err error) error {
	var apiErr *openai.Error
	if errors.As(err, &apiErr) {
		// Sanitiza a mensagem original removendo possíveis chaves
		msg := apiKeyPattern.ReplaceAllString(apiErr.Error(), "***REDACTED***")

		switch apiErr.StatusCode {
		case 400, 413:
			if isContextLengthError(msg) {
				return fmt.Errorf("texto muito longo para o contexto do modelo — tente um texto menor ou use --model com contexto maior")
			}
			return fmt.Errorf("erro na requisição (HTTP %d): %s", apiErr.StatusCode, msg)
		case 401:
			return fmt.Errorf("credenciais inválidas — verifique TLDR_API_KEY")
		case 429:
			return fmt.Errorf("limite de requisições excedido — tente novamente mais tarde")
		case 500, 502, 503:
			return fmt.Errorf("serviço temporariamente indisponível — tente novamente")
		default:
			return fmt.Errorf("erro da API (HTTP %d): %s", apiErr.StatusCode, msg)
		}
	}

	// Sanitiza o erro original mesmo sem ser APIError
	msg := apiKeyPattern.ReplaceAllString(err.Error(), "***REDACTED***")
	return fmt.Errorf("erro na chamada da API: %s", msg)
}
