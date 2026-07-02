package input

import (
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

// MaxInputSize é o tamanho máximo aceito para entrada (arquivo ou stdin).
const MaxInputSize int64 = 10 * 1024 * 1024 // 10 MB

// ReadFile lê o conteúdo de um arquivo. Apenas UTF-8 é suportado.
func ReadFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("não foi possível acessar %s: %w", path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s é um diretório, não um arquivo", path)
	}
	if info.Size() > MaxInputSize {
		return "", fmt.Errorf(
			"arquivo muito grande (%d bytes) — tamanho máximo permitido: %d bytes",
			info.Size(), MaxInputSize,
		)
	}
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
	limitedReader := io.LimitReader(os.Stdin, MaxInputSize) // limite de 10 MB
	data, err := io.ReadAll(limitedReader)
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
