package input

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

// MaxInputSize é o tamanho máximo aceito para entrada (arquivo ou stdin).
const MaxInputSize int64 = 10 * 1024 * 1024 // 10 MB

// StdinTimeout é o tempo máximo de espera para leitura do stdin.
// Pode ser alterado para testes, mas não é seguro para concorrência.
var StdinTimeout = 30 * time.Second

// ErrInputTooLarge indica que a entrada excedeu o tamanho máximo permitido.
var ErrInputTooLarge = fmt.Errorf("entrada muito grande — tamanho máximo permitido: %d bytes", MaxInputSize)

// ReadFromFile lê o conteúdo de um arquivo. Apenas UTF-8 é suportado.
// Expande "~" para o diretório home do usuário.
// NOTA: Usa uma única chamada de leitura (evita TOCTOU entre Stat e ReadFile).
func ReadFromFile(path string) (string, error) {
	// Expande ~ para o diretório home
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("não foi possível obter o diretório home: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("não foi possível abrir %s: %w", path, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("não foi possível acessar %s: %w", path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s é um diretório, não um arquivo", path)
	}

	// Lê no máximo MaxInputSize+1 bytes para detectar se o arquivo é maior
	limit := MaxInputSize + 1
	limitedReader := io.LimitReader(f, limit)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("não foi possível ler o arquivo %s: %w", path, err)
	}

	if int64(len(data)) > MaxInputSize {
		return "", ErrInputTooLarge
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
// O timeout máximo de espera é definido por StdinTimeout.
//
// Para testes, use ReadFromStdinWithTimeout para evitar modificar a variável global.
func ReadFromStdin() (string, error) {
	return readFromStdin(StdinTimeout)
}

// ReadFromStdinWithTimeout é como ReadFromStdin mas com timeout customizável.
// Útil para evitar modificar a variável global StdinTimeout em testes.
func ReadFromStdinWithTimeout(timeout time.Duration) (string, error) {
	return readFromStdin(timeout)
}

// readFromStdin é a implementação compartilhada.
func readFromStdin(timeout time.Duration) (string, error) {
	// Detecta se stdin é terminal — sem dados disponíveis
	if isTerminal() {
		return "", fmt.Errorf("nenhum texto fornecido — forneça um arquivo ou pipe via stdin\n\n"+
			"Exemplos:\n"+
			"  tldr documento.txt --lang pt-br\n"+
			"  echo \"texto longo\" | tldr --lang en")
	}

	// Valida o timeout: se for zero ou negativo, usa o padrão de 30s
	// para evitar cancelamento imediato do contexto.
	readTimeout := timeout
	if readTimeout <= 0 {
		readTimeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	defer cancel()

	type readResult struct {
		data []byte
		err  error
	}
	ch := make(chan readResult, 1)

	// Lê MaxInputSize + 1 bytes para detectar se a entrada foi truncada
	limit := MaxInputSize + 1

	// Captura o stdin em variável local antes de lançar a goroutine,
	// evitando data race entre a leitura da goroutine e a restauração
	// de os.Stdin pelo caller (ex: testes que substituem temporariamente).
	stdin := os.Stdin

	go func() {
		limitedReader := io.LimitReader(stdin, limit)
		data, err := io.ReadAll(limitedReader)
		ch <- readResult{data, err}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			return "", fmt.Errorf("erro ao ler da entrada padrão: %w", res.err)
		}

		if int64(len(res.data)) > MaxInputSize {
			return "", ErrInputTooLarge
		}

		if len(res.data) == 0 {
			return "", fmt.Errorf("nenhum texto recebido via stdin")
		}

		if !utf8.Valid(res.data) {
			return "", fmt.Errorf("entrada não está codificada em UTF-8")
		}

		return string(res.data), nil

	case <-ctx.Done():
		// NOTA: A goroutine de leitura (io.ReadAll) será abandonada.
		// Isso é aceitável pois:
		// 1. O processo CLI será encerrado em seguida via main().
		// 2. Em testes, o pipe de teste eventualmente será fechado
		//    pelo lado de escrita, desbloqueando a goroutine.
		// 3. Fechar os.Stdin como efeito colateral global não é
		//    seguro — evita-se essa prática.
		return "", fmt.Errorf("timeout ao ler da entrada padrão (%v) — o pipe pode estar travado", timeout)
	}
}

// IsStdinRedirected verifica se stdin está conectado a um pipe ou
// redirecionamento (ou seja, não é um terminal).
func IsStdinRedirected() bool {
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
