package cmd

import (
	"fmt"

	"github.com/Elissdev/tl-dr/internal/config"
	"github.com/Elissdev/tl-dr/internal/input"
	"github.com/Elissdev/tl-dr/internal/summarizer"
	"github.com/spf13/cobra"
)

var (
	lang             string
	model            string
	customPromptFlag string
)

var rootCmd = &cobra.Command{
	Use:   "tldr [flags] [<arquivo>]",
	Short: "tl;dr — Resumidor de texto via CLI",
	Long: `tl;dr é uma ferramenta de linha de comando que recebe um texto
(de arquivo ou stdin) e produz um resumo conciso no idioma especificado.

Documentação: https://github.com/Elissdev/tl-dr`,
	Args:          cobra.MaximumNArgs(1),
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Carregar configuração
		cfg, err := config.Load()
		if err != nil {
			return WrapExitError(ExitArgumentError, err)
		}

		// 2. Resolver modelo (flag > env > hardcoded)
		resolvedModel := cfg.DefaultModel
		if model != "" {
			resolvedModel = model
		}

		// 3. Resolver idioma (flag > env)
		resolvedLang := lang
		if resolvedLang == "" {
			resolvedLang = cfg.DefaultLang
		}
		if resolvedLang == "" {
			return NewExitError(ExitArgumentError,
				"idioma é obrigatório: use --lang ou defina TLDR_DEFAULT_LANG")
		}

		// 4. Ler entrada
		text, err := input.ReadInput(args)
		if err != nil {
			return WrapExitError(ExitArgumentError, err)
		}
		if text == "" {
			return NewExitError(ExitArgumentError,
				"entrada vazia — forneça um texto para resumir")
		}

		// 5. Construir prompt
		finalPrompt := buildPrompt(resolvedLang, customPromptFlag)

		// 6. Chamar API
		s := summarizer.New(summarizer.Config{
			APIKey:  cfg.APIKey,
			BaseURL: cfg.BaseURL,
			Model:   resolvedModel,
			Timeout: cfg.Timeout,
		})

		// A chave de API já foi copiada para o cliente da API (struct
		// summarizer.Config acima); podemos zerar a nossa cópia local.
		// ATENÇÃO: cfg.Clear() deve permanecer APÓS a cópia da APIKey
		// para summarizer.Config. Se no futuro o Config for passado por
		// referência, esta ordem precisará ser ajustada.
		cfg.Clear()

		summary, err := s.Summarize(cmd.Context(), finalPrompt, text)
		if err != nil {
			return WrapExitError(ExitAPIError,
				fmt.Errorf("erro na API: %w", err))
		}

		// 7. Escrever saída no stdout
		fmt.Print(summary)

		return nil
	},
}

// Execute executa o comando raiz. Retorna o erro, se houver, para que o
// caller (main) possa fazer cleanup adequado antes de os.Exit.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&lang, "lang", "l", "", "Idioma do resumo (ex: pt-br, en, es)")
	rootCmd.Flags().StringVarP(&model, "model", "m", "", "Modelo a usar (default: deepseek/deepseek-v4-flash)")
	rootCmd.Flags().StringVarP(&customPromptFlag, "prompt", "p", "", "Prompt customizado para o resumo")

	// --lang é validado manualmente no RunE (pode vir via TLDR_DEFAULT_LANG)
}

// buildPrompt constrói o prompt final para a API.
// Quando lang é pt-br ou pt, usa um template em português para o prompt padrão.
func buildPrompt(lang, customPrompt string) string {
	if customPrompt != "" {
		return fmt.Sprintf("Responda em %s.\n\n%s", lang, customPrompt)
	}
	switch lang {
	case "pt-br", "pt":
		return fmt.Sprintf("Resuma o texto a seguir em %s. Seja conciso mas capture todos os pontos-chave.", lang)
	default:
		return fmt.Sprintf("Summarize the following text in %s. Be concise but capture all key points.", lang)
	}
}
