package summarizer

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

// Config defines the configuration for the summarizer.
type Config struct {
	APIKey  string
	BaseURL string
	Model   string
}

// Summarizer handles communication with the API.
type Summarizer struct {
	client *openai.Client
	model  string
}

// New creates a new Summarizer with the given config.
func New(cfg Config) *Summarizer {
	client := openai.NewClient(
		option.WithAPIKey(cfg.APIKey),
		option.WithBaseURL(cfg.BaseURL),
	)
	return &Summarizer{
		client: &client,
		model:  cfg.Model,
	}
}

// Summarize sends a prompt and text to the API and returns the summary.
func (s *Summarizer) Summarize(ctx context.Context, prompt, text string) (string, error) {
	chat, err := s.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.String(fmt.Sprintf("%s\n\n%s", prompt, text)),
				},
			},
		}},
		Model: shared.ChatModel(s.model),
	})
	if err != nil {
		return "", classifyAPIError(err)
	}

	if len(chat.Choices) == 0 {
		return "", fmt.Errorf("API retornou resposta vazia")
	}

	return chat.Choices[0].Message.Content, nil
}

// classifyAPIError mapeia erros da API para erros amigáveis com exit codes apropriados.
func classifyAPIError(err error) error {
	return fmt.Errorf("erro na chamada da API: %w", err)
}
