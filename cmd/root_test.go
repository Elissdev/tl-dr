package cmd

import (
	"os"
	"strings"
	"testing"
)

// As constantes SafetyPrefixPT e SafetyPrefixEN estão definidas em root.go
// e são reexportadas aqui para uso nos testes.
// NOTA: usamos as próprias constantes do pacote, não duplicatas.

func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		name         string
		lang         string
		customPrompt string
		wantPrefix   string // verifica que o resultado começa com este prefixo
		wantSuffix   string // verifica que o resultado termina com este sufixo
	}{
		{
			name:         "default prompt en",
			lang:         "en",
			customPrompt: "",
			wantPrefix:   SafetyPrefixEN,
			wantSuffix:   "Summarize the following text in en. Be concise but capture all key points.",
		},
		{
			name:         "default prompt pt-br",
			lang:         "pt-br",
			customPrompt: "",
			wantPrefix:   SafetyPrefixPT,
			wantSuffix:   "Resuma o texto a seguir em pt-br. Seja conciso mas capture todos os pontos-chave.",
		},
		{
			name:         "default prompt pt",
			lang:         "pt",
			customPrompt: "",
			wantPrefix:   SafetyPrefixPT,
			wantSuffix:   "Resuma o texto a seguir em pt. Seja conciso mas capture todos os pontos-chave.",
		},
		{
			name:         "custom prompt with lang en",
			lang:         "en",
			customPrompt: "Explain like I'm 5",
			wantPrefix:   SafetyPrefixEN,
			wantSuffix:   "Explain like I'm 5",
		},
		{
			name:         "custom prompt with lang pt-br",
			lang:         "pt-br",
			customPrompt: "Resuma para um leigo no assunto",
			wantPrefix:   SafetyPrefixPT,
			wantSuffix:   "Resuma para um leigo no assunto",
		},
		{
			name:         "custom prompt with lang es",
			lang:         "es",
			customPrompt: "Haz un resumen",
			wantPrefix:   SafetyPrefixEN,
			wantSuffix:   "Haz un resumen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPrompt(tt.lang, tt.customPrompt)

			if !strings.HasPrefix(result, tt.wantPrefix) {
				t.Errorf("buildPrompt(%q, %q) não começa com o prefixo esperado\nresult = %q\nprefix = %q", tt.lang, tt.customPrompt, result, tt.wantPrefix)
			}
			if !strings.HasSuffix(result, tt.wantSuffix) {
				t.Errorf("buildPrompt(%q, %q) não termina com o sufixo esperado\nresult = %q\nsuffix = %q", tt.lang, tt.customPrompt, result, tt.wantSuffix)
			}
		})
	}
}

func TestBuildPromptSafetyPrefixAlwaysPresent(t *testing.T) {
	// Verifica que TODAS as combinações de idioma incluem o prefixo de segurança
	langs := []struct {
		lang   string
		prefix string
	}{
		{"en", SafetyPrefixEN},
		{"pt", SafetyPrefixPT},
		{"pt-br", SafetyPrefixPT},
		{"es", SafetyPrefixEN},
		{"fr", SafetyPrefixEN},
	}

	t.Run("without custom prompt", func(t *testing.T) {
		for _, l := range langs {
			result := buildPrompt(l.lang, "")
			if !strings.HasPrefix(result, l.prefix) {
				t.Errorf("buildPrompt(%q, \"\") = %q, esperava prefixo %q", l.lang, result, l.prefix)
			}
		}
	})

	t.Run("with custom prompt", func(t *testing.T) {
		for _, l := range langs {
			result := buildPrompt(l.lang, "custom text")
			if !strings.HasPrefix(result, l.prefix) {
				t.Errorf("buildPrompt(%q, \"custom text\") = %q, esperava prefixo %q", l.lang, result, l.prefix)
			}
		}
	})
}

func TestSanitizeOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "texto normal sem escapes",
			input: "Este é um resumo normal.",
			want:  "Este é um resumo normal.",
		},
		{
			name:  "remove CSI code (ESC[31m)",
			input: "Texto \x1b[31mvermelho\x1b[0m normal",
			want:  "Texto vermelho normal",
		},
		{
			name:  "remove OSC sequence (ESC]...BEL)",
			input: "Teste \x1b]8;;https://evil.com\x07link\x1b]8;;\x07",
			want:  "Teste link",
		},
		{
			name:  "remove ESC sequence longa (cores 256)",
			input: "\x1b[38;5;196mcolorido\x1b[0m",
			want:  "colorido",
		},
		{
			name:  "preserva newlines e tabs",
			input: "linha1\n\tlinha2\nlinha3",
			want:  "linha1\n\tlinha2\nlinha3",
		},
		{
			name:  "preserva carriage return",
			input: "texto\rcom CR",
			want:  "texto\rcom CR",
		},
		{
			name:  "remove DCS sequence (ESC P...ST)",
			input: "\x1bP?Amazing\x1b\\",
			want:  "",
		},
		{
			name:  "remove SOS sequence (ESC X...ST)",
			input: "a\x1bXevil\x1b\\b",
			want:  "ab",
		},
		{
			name:  "remove PM sequence (ESC ^...ST)",
			input: "\x1b^manipulate\x1b\\text",
			want:  "text",
		},
		{
			name:  "remove APC sequence (ESC _...ST)",
			input: "\x1b_injection\x1b\\ok",
			want:  "ok",
		},
		{
			name:  "string vazia",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeOutput(tt.input)
			if result != tt.want {
				t.Errorf("sanitizeOutput(%q) = %q, want %q", tt.input, result, tt.want)
			}
		})
	}
}

func TestSanitizeOutputEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "ESC isolado no final",
			input: "texto\x1b",
			want:  "texto",
		},
		{
			name:  "CSI incompleto no final (ESC[ sem final)",
			input: "texto\x1b[",
			want:  "texto",
		},
		{
			name:  "OSC incompleto no final (ESC] sem ST)",
			input: "texto\x1b]8;;https://evil.com",
			want:  "texto",
		},
		{
			name:  "ESC isolado no meio",
			input: "a\x1bb",
			want:  "ab",
		},
		{
			name:  "ESC com caractere de controle após",
			input: "a\x1b\x00b",
			want:  "ab",
		},
		{
			name:  "ESC + caractere não ASCII após",
			input: "a\x1b\xC3\xA1b", // ESC seguido de á em UTF-8
			want:  "ab",
		},
		{
			name:  "múltiplos ESC consecutivos",
			input: "a\x1b\x1b[31mb",
			want:  "ab",
		},
		{
			name:  "caractere de controle \\x00 isolado",
			input: "a\x00b\x01c",
			want:  "abc",
		},
		{
			name:  "unicode longo preservado",
			input: "Olá mundo 🌍! ✅",
			want:  "Olá mundo 🌍! ✅",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeOutput(tt.input)
			if result != tt.want {
				t.Errorf("sanitizeOutput(%q) = %q, want %q", tt.input, result, tt.want)
			}
		})
	}
}

func TestGetLocale(t *testing.T) {
	t.Run("pt-br retorna config portuguesa", func(t *testing.T) {
		lc := getLocale("pt-br")
		if lc.SafetyPrefix != SafetyPrefixPT {
			t.Errorf("SafetyPrefix = %q, want %q", lc.SafetyPrefix, SafetyPrefixPT)
		}
		if !strings.Contains(lc.DefaultPrompt, "Resuma") {
			t.Errorf("DefaultPrompt = %q, want contendo 'Resuma'", lc.DefaultPrompt)
		}
	})

	t.Run("pt retorna config portuguesa", func(t *testing.T) {
		lc := getLocale("pt")
		if lc.SafetyPrefix != SafetyPrefixPT {
			t.Errorf("SafetyPrefix = %q, want %q", lc.SafetyPrefix, SafetyPrefixPT)
		}
	})

	t.Run("en retorna default (inglês)", func(t *testing.T) {
		lc := getLocale("en")
		if lc.SafetyPrefix != SafetyPrefixEN {
			t.Errorf("SafetyPrefix = %q, want %q", lc.SafetyPrefix, SafetyPrefixEN)
		}
	})

	t.Run("idioma desconhecido retorna default", func(t *testing.T) {
		lc := getLocale("es")
		if lc.SafetyPrefix != SafetyPrefixEN {
			t.Errorf("SafetyPrefix = %q, want %q", lc.SafetyPrefix, SafetyPrefixEN)
		}
	})

	t.Run("adicionar novo idioma ao mapa não quebra existentes", func(t *testing.T) {
		// Simula: se alguém adicionar "es" ao mapa, os outros devem continuar
		lcPT := getLocale("pt")
		lcEN := getLocale("fr")
		if lcPT.SafetyPrefix != SafetyPrefixPT {
			t.Error("pt deve continuar com prefixo PT")
		}
		if lcEN.SafetyPrefix != SafetyPrefixEN {
			t.Error("fr deve continuar com prefixo EN (fallback)")
		}
	})
}

func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		name string
		vals []string
		want string
	}{
		{name: "primeiro valor não vazio", vals: []string{"a", "b", "c"}, want: "a"},
		{name: "primeiro após vazios", vals: []string{"", "", "x", "y"}, want: "x"},
		{name: "todos vazios", vals: []string{"", "", ""}, want: ""},
		{name: "único valor", vals: []string{"único"}, want: "único"},
		{name: "slice vazio", vals: []string{}, want: ""},
		{name: "string vazia primeiro", vals: []string{"", "segundo"}, want: "segundo"},
		{name: "todos preenchidos", vals: []string{"primeiro", "segundo", "terceiro"}, want: "primeiro"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := firstNonEmpty(tt.vals...)
			if got != tt.want {
				t.Errorf("firstNonEmpty(%v) = %q, want %q", tt.vals, got, tt.want)
			}
		})
	}
}

func TestExecute(t *testing.T) {
	t.Run("langPattern aceita formatos válidos", func(t *testing.T) {
		validLangs := []string{
			"en", "pt", "es", "fr", "de", "ja", "zh",
			"pt-br", "en-US", "zh-CN", "zh-Hans", "sgn-BR",
			"pt_BR", "en_US",
			"ast", "fil", "mwl",
		}
		for _, l := range validLangs {
			if !langPattern.MatchString(l) {
				t.Errorf("langPattern.MatchString(%q) = false, want true", l)
			}
		}
	})

	t.Run("langPattern rejeita formatos inválidos", func(t *testing.T) {
		invalidLangs := []string{
			"", "1234", "a", "abcde-fghij", "pt_br_extra",
		}
		for _, l := range invalidLangs {
			if langPattern.MatchString(l) {
				t.Errorf("langPattern.MatchString(%q) = true, want false", l)
			}
		}
	})

	t.Run("--no-sanitize flag não quebra execução normal", func(t *testing.T) {
		prevKey, keyExisted := os.LookupEnv("TLDR_API_KEY")
		os.Setenv("TLDR_API_KEY", "sk-test-key-for-execute")
		defer func() {
			if keyExisted {
				os.Setenv("TLDR_API_KEY", prevKey)
			} else {
				os.Unsetenv("TLDR_API_KEY")
			}
		}()

		cmd := newRootCommand("test")
		cmd.SetArgs([]string{"--lang", "en", "--no-sanitize"})
		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() com --no-sanitize sem entrada = nil, want erro de entrada")
		}
		if strings.Contains(err.Error(), "unknown flag") {
			t.Errorf("flag --no-sanitize não foi reconhecida: %v", err)
		}
	})

	t.Run("--no-sanitize flag tem valor padrão false", func(t *testing.T) {
		cmd := newRootCommand("test")
		flag := cmd.Flags().Lookup("no-sanitize")
		if flag == nil {
			t.Fatal("flag --no-sanitize não registrada")
		}
		if flag.DefValue != "false" {
			t.Errorf("default de --no-sanitize = %q, want 'false'", flag.DefValue)
		}
		// Verifica que o valor inicial é false obtendo-o via GetBool
		val, _ := cmd.Flags().GetBool("no-sanitize")
		if val {
			t.Error("no-sanitize padrão = true, want false (default)")
		}
	})

	t.Run("sem chave de API", func(t *testing.T) {
		// Garante que TLDR_API_KEY não está definida
		prevKey, keyExisted := os.LookupEnv("TLDR_API_KEY")
		os.Unsetenv("TLDR_API_KEY")

		cmd := newRootCommand("test")
		cmd.SetArgs([]string{"--lang", "pt-br"})
		err := cmd.Execute()

		// Restaura para não afetar outros testes
		if keyExisted {
			os.Setenv("TLDR_API_KEY", prevKey)
		} else {
			os.Unsetenv("TLDR_API_KEY")
		}

		if err == nil {
			t.Fatal("Execute() sem API key = nil, want erro")
		}
		if !strings.Contains(err.Error(), "chave") {
			t.Errorf("erro = %q, want contendo 'chave'", err.Error())
		}
	})

	t.Run("idioma resolvido de env vars, mas sem entrada", func(t *testing.T) {
		prevKey, keyExisted := os.LookupEnv("TLDR_API_KEY")
		os.Setenv("TLDR_API_KEY", "sk-test-key-for-execute")
		// Define o idioma via variável de ambiente (simula .env carregado)
		prevLang, langExisted := os.LookupEnv("TLDR_DEFAULT_LANG")
		os.Setenv("TLDR_DEFAULT_LANG", "pt-br")
		defer func() {
			if keyExisted {
				os.Setenv("TLDR_API_KEY", prevKey)
			} else {
				os.Unsetenv("TLDR_API_KEY")
			}
			if langExisted {
				os.Setenv("TLDR_DEFAULT_LANG", prevLang)
			} else {
				os.Unsetenv("TLDR_DEFAULT_LANG")
			}
		}()

		// Não passa --lang — deve usar TLDR_DEFAULT_LANG
		cmd := newRootCommand("test")
		cmd.SetArgs([]string{})
		err := cmd.Execute()

		// O idioma vem da env var, então o erro deve ser de leitura (sem stdin)
		if err == nil {
			t.Fatal("Execute() sem entrada = nil, want erro")
		}
		if !strings.Contains(err.Error(), "nenhum texto fornecido") {
			t.Errorf("erro = %q, want 'nenhum texto fornecido'", err.Error())
		}
	})

	t.Run("idioma inválido", func(t *testing.T) {
		prevKey, keyExisted := os.LookupEnv("TLDR_API_KEY")
		os.Setenv("TLDR_API_KEY", "sk-test-key-for-execute")
		defer func() {
			if keyExisted {
				os.Setenv("TLDR_API_KEY", prevKey)
			} else {
				os.Unsetenv("TLDR_API_KEY")
			}
		}()

		cmd := newRootCommand("test")
		cmd.SetArgs([]string{"--lang", "1234"})
		err := cmd.Execute()

		if err == nil {
			t.Fatal("Execute() com idioma inválido = nil, want erro")
		}
		if !strings.Contains(err.Error(), "idioma inválido") {
			t.Errorf("erro = %q, want 'idioma inválido'", err.Error())
		}
	})

	t.Run("arquivo inexistente", func(t *testing.T) {
		prevKey, keyExisted := os.LookupEnv("TLDR_API_KEY")
		os.Setenv("TLDR_API_KEY", "sk-test-key-for-execute")
		defer func() {
			if keyExisted {
				os.Setenv("TLDR_API_KEY", prevKey)
			} else {
				os.Unsetenv("TLDR_API_KEY")
			}
		}()

		cmd := newRootCommand("test")
		cmd.SetArgs([]string{"--lang", "en", "/caminho/inexistente/arquivo.txt"})
		err := cmd.Execute()

		if err == nil {
			t.Fatal("Execute() com arquivo inexistente = nil, want erro")
		}
	})
}
