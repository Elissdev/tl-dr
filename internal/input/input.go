package input

import (
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

// ReadFile lê o conteúdo de um arquivo. Apenas UTF-8 é suportado.
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("não foi possível ler o arquivo %s: %w", path, err)
	}
	if !utf8.Valid(data) {
		return "", fmt.Errorf("arquivo %s não está codificado em UTF-8", path)
	}
	return string(data), nil
}

// ReadStdin lê o conteúdo da entrada padrão. Apenas UTF-8 é suportado.
func ReadStdin() (string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("erro ao ler da entrada padrão: %w", err)
	}
	if len(data) == 0 {
		return "", nil
	}
	if !utf8.Valid(data) {
		return "", fmt.Errorf("entrada não está codificada em UTF-8")
	}
	return string(data), nil
}

// IsStdinAvailable verifica se há dados disponíveis no stdin (pipe/redirecionamento).
func IsStdinAvailable() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}
