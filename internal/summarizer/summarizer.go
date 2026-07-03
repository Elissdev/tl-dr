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

// ErrTruncated é retornado quando o modelo atinge o limite de tokens e o resumo
// ficou incompleto. O conteúdo parcial ainda está disponível no retorno da função.
var ErrTruncated = errors.New("resumo truncado: o modelo atingiu o limite de tokens")

// Patterns para redação de credenciais em mensagens de erro.
// ATENÇÃO: aplicados APENAS em mensagens de erro (nunca no conteúdo do usuário).
// Cada pattern cobre um formato de chave/token conhecido.
//
// A lista prioriza patterns específicos e conhecidos. O fallback genérico
// (strings com 60+ caracteres) existe para capturar tokens proprietários
// não conhecidos, mas pode gerar falsos positivos com hashes longos (ex: SHA-512)
// ou conteúdo legítimo do usuário. Por isso, é aplicado APENAS em mensagens
// de erro retornadas pela API, nunca no conteúdo processado do usuário.
var apiKeyRedactors = []*regexp.Regexp{
	regexp.MustCompile(`sk-proj-[a-z0-9]{20,}`),             // OpenAI (novo formato)
	regexp.MustCompile(`sk-[a-z0-9]{20,}`),                   // OpenAI (legado)
	regexp.MustCompile(`deepseek-[a-z0-9]{20,}`),             // DeepSeek
	regexp.MustCompile(`sk-ant-[a-z0-9]{20,}`),               // Anthropic
	regexp.MustCompile(`ghp_[a-zA-Z0-9]{36}`),                // GitHub PAT
	regexp.MustCompile(`eyJ[a-zA-Z0-9_\-]+\.[a-zA-Z0-9_\-]+\.[a-zA-Z0-9_\-]+`), // JWT
	regexp.MustCompile(`(?i)api[_-]?key[=:]\s*\S{8,}`),   // api_key= / api-key:
	regexp.MustCompile(`(?i)token[=:]\s*\S{8,}`),          // token= / token:
	regexp.MustCompile(`[A-Za-z0-9_\-]{60,}`),              // Fallback genérico: strings com 60+ chars alfanuméricos
	// ATENÇÃO: Este fallback pode capturar SHA-512 (128 chars hex) ou hashes
	// legítimos em mensagens de erro. É aceitável pois a redação ocorre APENAS
	// em mensagens de erro (nunca no conteúdo do usuário), e é melhor redigir
	// um falso positivo do que vazar uma credencial desconhecida.
}

// Config define a configuração para o summarizer.
type Config struct {
	APIKey  string
	BaseURL string
	Model   string
	Timeout time.Duration
}

// Client gerencia a comunicação com a API de sumarização.
type Client struct {
	client *openai.Client
	model  string
	apiKey string // armazenada para redação em mensagens de erro
}

// New cria um novo Client com a configuração fornecida.
// Retorna erro se campos obrigatórios estiverem ausentes ou inválidos.
func New(cfg Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("API key é obrigatória")
	}
	if cfg.Model == "" {
		return nil, errors.New("modelo é obrigatório")
	}
	if cfg.BaseURL == "" {
		return nil, errors.New("base URL é obrigatória")
	}

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
	return &Client{
		client: &client,
		model:  cfg.Model,
		apiKey: cfg.APIKey,
	}, nil
}

// Summarize envia um prompt e texto para a API e retorna o resumo.
// O prompt é enviado como mensagem de sistema (reduz risco de injeção de prompt)
// e o texto do usuário como mensagem de usuário.
func (s *Client) Summarize(ctx context.Context, systemPrompt, userText string) (string, error) {
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
		return "", s.classifyAPIError(err)
	}

	if len(chat.Choices) == 0 {
		return "", fmt.Errorf("API retornou resposta vazia")
	}

	choice := chat.Choices[0]
	if choice.FinishReason == "length" {
		return choice.Message.Content, ErrTruncated
	}
	if choice.FinishReason == "content_filter" {
		return "", fmt.Errorf("resumo bloqueado pelo filtro de conteúdo da API")
	}
	if choice.FinishReason == "stop" && choice.Message.Content == "" {
		return "", fmt.Errorf("API retornou conteúdo vazio com finish_reason=stop")
	}

	return choice.Message.Content, nil
}

// redactedError envolve um erro original e apresenta uma mensagem redigida,
// mas preserva a cadeia de erros original para errors.Is/errors.As.
type redactedError struct {
	msg   string
	cause error
}

func (e *redactedError) Error() string { return e.msg }
func (e *redactedError) Unwrap() error { return e.cause }

// ErrTimeout indica que a requisição excedeu o tempo limite.
// O caller (ex: cmd/root.go) pode usar errors.Is para detectar
// este erro e mapear para ExitTimeout.
var ErrTimeout = errors.New("a requisição excedeu o tempo limite")

// classifyAPIError mapeia erros da API para mensagens mais amigáveis e
// remove qualquer vazamento acidental de credenciais ou tokens.
// Preserva a cadeia de erros original via redactedError para permitir
// errors.Is/errors.As sem expor credenciais na mensagem.
func (s *Client) classifyAPIError(err error) error {
	// Detecta timeout antes de tentar interpretar como erro da API,
	// pois timeouts podem manifestar-se como erros de rede antes mesmo
	// de uma resposta HTTP ser recebida.
	if errors.Is(err, context.DeadlineExceeded) ||
		strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "deadline") {
		return &redactedError{
			msg:   ErrTimeout.Error(),
			cause: err,
		}
	}

	var apiErr *openai.Error
	if errors.As(err, &apiErr) {
		msg := redactCredentials(apiErr.Error(), s.apiKey)

		switch apiErr.StatusCode {
		case 401:
			return &redactedError{
				msg:   "credenciais inválidas — verifique TLDR_API_KEY",
				cause: err,
			}
		case 429:
			return &redactedError{
				msg:   "limite de requisições excedido — tente novamente mais tarde",
				cause: err,
			}
		case 500, 502, 503:
			return &redactedError{
				msg:   "serviço temporariamente indisponível — tente novamente",
				cause: err,
			}
		default:
			return &redactedError{
				msg:   fmt.Sprintf("erro da API (HTTP %d): %s", apiErr.StatusCode, msg),
				cause: err,
			}
		}
	}

	msg := redactCredentials(err.Error(), s.apiKey)
	return &redactedError{
		msg:   fmt.Sprintf("erro na chamada da API: %s", msg),
		cause: err,
	}
}

// redactCredentials substitui quaisquer padrões de credenciais
// encontrados em s por "***REDACTED***".
// Também redige a chave de API fornecida (qualquer formato).
func redactCredentials(s, apiKey string) string {
	if apiKey != "" {
		s = strings.ReplaceAll(s, apiKey, "***REDACTED***")
	}
	for _, re := range apiKeyRedactors {
		s = re.ReplaceAllString(s, "***REDACTED***")
	}
	return s
}
