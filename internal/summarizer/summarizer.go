package summarizer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

// apiKeyPattern detecta possíveis chaves de API em mensagens de erro para
// fazer redação antes de exibir ao usuário.
// ATENÇÃO: esta regex é aplicada APENAS em mensagens de erro (nunca no
// conteúdo do usuário), então o risco de redigir conteúdo legítimo é baixo.
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
// Nota: config.Load() já garante timeout > 0, mas mantemos o fallback
// como segurança para chamadas diretas a esta função.
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

// classifyAPIError mapeia erros da API para mensagens mais amigáveis e
// remove qualquer vazamento acidental de credenciais ou tokens.
func classifyAPIError(err error) error {
	var apiErr *openai.Error
	if errors.As(err, &apiErr) {
		// Sanitiza a mensagem original removendo possíveis chaves
		msg := apiKeyPattern.ReplaceAllString(apiErr.Error(), "***REDACTED***")

		switch apiErr.StatusCode {
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
