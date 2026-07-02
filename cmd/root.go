package cmd

import (
	"fmt"
	"os"

	"github.com/Elissdev/tl-dr/internal/config"
	"github.com/Elissdev/tl-dr/internal/input"
	"github.com/Elissdev/tl-dr/internal/summarizer"
	"github.com/spf13/cobra"
)

var (
	lang   string
	model  string
	prompt string
)

var rootCmd = &cobra.Command{
	Use:   "tldr [flags] [<arquivo>]",
	Short: "tl;dr — Resumidor de texto via CLI",
	Long: `tl;dr é uma ferramenta de linha de comando que recebe um texto
(de arquivo ou stdin) e produz um resumo conciso no idioma especificado.

Documentação: https://github.com/Elissdev/tl-dr`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Carregar configuração
		cfg := config.Load()

			// 2. Validar configuração
		if err := cfg.Validate(); err != nil {
			return err
		}

		// 3. Resolver modelo (flag > env > hardcoded)
		resolvedModel := cfg.DefaultModel
		if model != "" {
			resolvedModel = model
		}

		// 4. Resolver idioma (flag > env)
		resolvedLang := lang
		if resolvedLang == "" {
			resolvedLang = cfg.DefaultLang
		}
		if resolvedLang == "" {
			return fmt.Errorf("idioma é obrigatório: use --lang ou defina TLDR_DEFAULT_LANG")
		}

		// 5. Ler entrada
		var text string
		if len(args) > 0 {
			data, err := input.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("erro ao ler arquivo: %w", err)
			}
			text = data
		} else {
			data, err := input.ReadStdin()
			if err != nil {
				return fmt.Errorf("erro ao ler stdin: %w", err)
			}
			text = data
		}

		if text == "" {
			return fmt.Errorf("nenhum texto fornecido — passe um arquivo ou pipe via stdin")
		}

		// 6. Construir prompt
		finalPrompt := buildPrompt(resolvedLang, prompt)

		// 7. Chamar API
		s := summarizer.New(summarizer.Config{
			APIKey:   cfg.APIKey,
			BaseURL:  cfg.BaseURL,
			Model:    resolvedModel,
		})

		summary, err := s.Summarize(cmd.Context(), finalPrompt, text)
		if err != nil {
			return fmt.Errorf("erro na API: %w", err)
		}

		// 8. Escrever saída no stdout
		fmt.Print(summary)

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&lang, "lang", "l", "", "Idioma do resumo (ex: pt-br, en, es)")
	rootCmd.Flags().StringVarP(&model, "model", "m", "", "Modelo a usar (default: deepseek/deepseek-v4-flash)")
	rootCmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Prompt customizado para o resumo")

	// --lang é validado manualmente no RunE (pode vir via TLDR_DEFAULT_LANG)
}

// buildPrompt constrói o prompt final para a API.
func buildPrompt(lang, customPrompt string) string {
	if customPrompt != "" {
		return fmt.Sprintf("Responda em %s.\n\n%s", lang, customPrompt)
	}
	return fmt.Sprintf("Summarize the following text in %s. Be concise but capture all key points.", lang)
}
