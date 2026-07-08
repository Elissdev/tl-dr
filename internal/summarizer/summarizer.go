package summarizer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
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

	// HTTPClient permite injetar um cliente HTTP customizado (ex: para testes com go-vcr).
	// Se nil, um cliente padrão com Timeout é criado.
	//
	// SEGURANÇA: o cliente HTTP informado NÃO deve utilizar
	// tls.Config{InsecureSkipVerify: true} ou transportes que
	// desabilitem a verificação de certificados TLS, pois isso
	// abriria uma vulnerabilidade de MITM (Man-In-The-Middle).
	// Também não devem ser usados transportes que loguem ou
	// desviem requisições para destinos não autorizados.
	HTTPClient *http.Client
}

// Client gerencia a comunicação com a API de sumarização.
type Client struct {
	mu     sync.Mutex
	client *openai.Client
	model  string
	apiKey []byte  // armazenada como []byte para permitir limpeza na memória
	cleared bool   // true após Clear() — uso posterior causa panic
}

// New cria um novo Client com a configuração fornecida.
// Retorna erro se campos obrigatórios estiverem ausentes ou inválidos.
// ATENÇÃO: A string cfg.APIKey é copiada internamente como []byte.
// O caller deve zerar a string original via cfg.Clear() após criar o Client.
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

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	} else if httpClient.Timeout == 0 {
		httpClient.Timeout = timeout
	}

	// Validação de segurança: rejeita clientes HTTP com verificação TLS desabilitada
	if err := validateHTTPClientTLS(httpClient); err != nil {
		return nil, err
	}

	client := openai.NewClient(
		option.WithAPIKey(cfg.APIKey),
		option.WithBaseURL(cfg.BaseURL),
		option.WithHTTPClient(httpClient),
		option.WithMaxRetries(0), // sem retry automático — tratamos erros no classifyAPIError
	)
	// Faz uma cópia em []byte para permitir zerar a memória posteriormente
	apiKeyBytes := make([]byte, len(cfg.APIKey))
	copy(apiKeyBytes, cfg.APIKey)

	return &Client{
		client: &client,
		model:  cfg.Model,
		apiKey: apiKeyBytes,
	}, nil
}

// validateHTTPClientTLS verifica se o http.Client fornecido não possui
// configurações TLS inseguras, como InsecureSkipVerify ativado.
// Retorna erro se encontrar uma configuração que comprometa a segurança
// da comunicação com a API.
func validateHTTPClientTLS(c *http.Client) error {
	if c == nil {
		return nil
	}

	transport := c.Transport
	if transport == nil {
		return nil // http.DefaultTransport é seguro
	}

	t, ok := transport.(*http.Transport)
	if !ok {
		// Transport customizado não é um *http.Transport padrão;
		// não podemos inspecionar, mas confiamos no caller.
		return nil
	}

	if t.TLSClientConfig != nil && t.TLSClientConfig.InsecureSkipVerify {
		return errors.New("HTTPClient com InsecureSkipVerify=true rejeitado: " +
			"verificação TLS desabilitada expõe a comunicação a ataques MITM")
	}

	// NOTA: RootCAs personalizado não é bloqueado — assumimos que é intencional.

	return nil
}

// Clear zera a chave de API da memória do Client.
// Deve ser chamado assim que o Client não for mais necessário.
// Após Clear, o Client não deve mais ser usado para chamadas de API
// — qualquer tentativa resultará em erro.
func (s *Client) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cleared {
		return // já foi limpo — seguro chamar múltiplas vezes
	}
	for i := range s.apiKey {
		s.apiKey[i] = 0
	}
	s.apiKey = nil
	s.cleared = true
}

// Summarize envia um prompt e texto para a API e retorna o resumo.
// O prompt é enviado como mensagem de sistema (reduz risco de injeção de prompt)
// e o texto do usuário como mensagem de usuário.
// Retorna erro se o Client já tiver sido limpo via Clear().
//
// Thread-safe: copia a apiKey internamente para evitar race condition
// com Clear() durante a chamada HTTP.
func (s *Client) Summarize(ctx context.Context, systemPrompt, userText string) (string, error) {
	s.mu.Lock()
	if s.cleared {
		s.mu.Unlock()
		return "", errors.New("summarizer: Client usado após Clear()")
	}
	// Copia a chave para uso seguro em classifyAPIError após o unlock,
	// evitando segurar o lock durante toda a chamada HTTP.
	apiKeyCopy := make([]byte, len(s.apiKey))
	copy(apiKeyCopy, s.apiKey)
	s.mu.Unlock()

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
		return "", s.classifyAPIError(err, apiKeyCopy)
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
//
// O parâmetro apiKey é uma cópia do slice interno, evitando race condition
// com Clear() durante a classificação do erro.
func (s *Client) classifyAPIError(err error, apiKey []byte) error {
	// Redige credenciais primeiro para evitar vazamento antes de qualquer
	// classificação do erro.
	redactedMsg := redactCredentials(err.Error(), apiKey)

	// 1. Timeout real por contexto expirado — detecta via errors.Is,
	// que é a forma confiável. O Go padrão da biblioteca (http.Client,
	// openai-go) propaga context.DeadlineExceeded para timeouts de rede.
	// context.Canceled NÃO entra aqui (cancelamento não é timeout).
	if errors.Is(err, context.DeadlineExceeded) {
		return &redactedError{
			msg:   fmt.Sprintf("%s: %s", ErrTimeout.Error(), redactedMsg),
			cause: fmt.Errorf("%w: %w", ErrTimeout, err),
		}
	}

	// 2. Erro da API com status HTTP — classifica por código.
	// IMPORTANTE: Isso vem ANTES da detecção por substring para evitar que
	// mensagens de erro da API que contenham "timeout" ou "deadline"
	// (ex: erro 400 "timeout parameter invalid") sejam falsamente
	// classificadas como timeout.
	var apiErr *openai.Error
	if errors.As(err, &apiErr) {
		msg := redactCredentials(apiErr.Error(), apiKey)

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

	// 3. Fallback para erros de rede não capturados por errors.Is:
	// Alguns transportes HTTP custom ou proxies podem não propagar
	// context.DeadlineExceeded corretamente. Neste caso, usamos
	// substring como heurística apenas para erros que NÃO são da API.
	if strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "deadline") {
		return &redactedError{
			msg:   fmt.Sprintf("%s: %s", ErrTimeout.Error(), redactedMsg),
			cause: fmt.Errorf("%w: %w", ErrTimeout, err),
		}
	}

	return &redactedError{
		msg:   fmt.Sprintf("erro na chamada da API: %s", redactedMsg),
		cause: err,
	}
}

// RedactCredentials substitui quaisquer padrões de credenciais
// encontrados em s por "***REDACTED***".
// Também redige a chave de API fornecida (qualquer formato).
//
// Esta função é exportada para permitir reuso em outras camadas
// (ex: testes de integração que precisam redigir dados antes de
// persistir cassetes HTTP).
func RedactCredentials(s, apiKey string) string {
	return redactCredentials(s, []byte(apiKey))
}

// redactCredentials é a implementação interna de RedactCredentials.
// O parâmetro apiKey é []byte para permitir que o caller gerencie
// o ciclo de vida da memória da chave.
func redactCredentials(s string, apiKey []byte) string {
	if len(apiKey) > 0 {
		s = strings.ReplaceAll(s, string(apiKey), "***REDACTED***")
	}
	for _, re := range apiKeyRedactors {
		s = re.ReplaceAllString(s, "***REDACTED***")
	}
	return s
}
