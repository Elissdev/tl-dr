package input

import (
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

// MaxInputSize é o tamanho máximo aceito para entrada (arquivo ou stdin).
const MaxInputSize int64 = 10 * 1024 * 1024 // 10 MB

// ErrInputTooLarge indica que a entrada excedeu o tamanho máximo permitido.
var ErrInputTooLarge = fmt.Errorf("entrada muito grande — tamanho máximo permitido: %d bytes", MaxInputSize)

// ReadFile lê o conteúdo de um arquivo. Apenas UTF-8 é suportado.
func ReadFromFile(path string) (string, error) {
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

// isTerminal verifica se stdin é um terminal (sem pipe/redirecionamento).
func isTerminal() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// ReadFromStdin lê o conteúdo da entrada padrão. Apenas UTF-8 é suportado.
// Se a entrada exceder MaxInputSize, retorna ErrInputTooLarge.
// Se stdin for um terminal (sem pipe/redirecionamento), retorna erro.
func ReadFromStdin() (string, error) {
	// Detecta se stdin é terminal — sem dados disponíveis
	if isTerminal() {
		return "", fmt.Errorf("nenhum texto fornecido — passe um arquivo ou pipe via stdin")
	}

	// Lê MaxInputSize + 1 bytes para detectar se a entrada foi truncada
	limit := MaxInputSize + 1
	limitedReader := io.LimitReader(os.Stdin, limit)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("erro ao ler da entrada padrão: %w", err)
	}

	if int64(len(data)) > MaxInputSize {
		return "", ErrInputTooLarge
	}

	if len(data) == 0 {
		return "", fmt.Errorf("nenhum texto recebido via stdin")
	}

	if !utf8.Valid(data) {
		return "", fmt.Errorf("entrada não está codificada em UTF-8")
	}

	return string(data), nil
}

// IsStdinAvailable verifica se há dados disponíveis no stdin (pipe/redirecionamento).
// Mantida para compatibilidade externa, mas internamente ReadFromStdin e ReadInput
// fazem esta verificação por conta própria (eliminando TOCTOU).
func IsStdinAvailable() bool {
	return !isTerminal()
}

// ReadInput lê o texto de entrada de arquivo ou stdin, seguindo a ordem
// de precedência: argumento posicional (arquivo) > stdin > erro.
// O parâmetro args são os argumentos posicionais do CLI.
func ReadInput(args []string) (string, error) {
	if len(args) > 0 {
		return ReadFromFile(args[0])
	}
	return ReadFromStdin()
}
