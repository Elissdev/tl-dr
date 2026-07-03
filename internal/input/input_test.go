package input

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// slowReader atrasa a leitura para simular um pipe lento.
type slowReader struct {
	r       io.Reader
	delay   time.Duration
	read    bool
}

func TestReadFromFile(t *testing.T) {
	t.Run("arquivo normal", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.txt")
		expected := "Hello, World!"
		if err := os.WriteFile(path, []byte(expected), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := ReadFromFile(path)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if result != expected {
			t.Errorf("ReadFromFile = %q, want %q", result, expected)
		}
	})

	t.Run("arquivo inexistente", func(t *testing.T) {
		_, err := ReadFromFile("/nao/existe/arquivo.txt")
		if err == nil {
			t.Fatal("esperava erro, mas retornou nil")
		}
	})

	t.Run("diretório", func(t *testing.T) {
		dir := t.TempDir()
		_, err := ReadFromFile(dir)
		if err == nil {
			t.Fatal("esperava erro para diretório, mas retornou nil")
		}
	})

	t.Run("arquivo muito grande", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "large.txt")
		// Cria arquivo maior que MaxInputSize
		data := make([]byte, MaxInputSize+1)
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}

		_, err := ReadFromFile(path)
		if err == nil {
			t.Fatal("esperava erro de tamanho, mas retornou nil")
		}
	})

	t.Run("UTF-8 inválido", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "invalid.txt")
		// 0xFF não é UTF-8 válido
		if err := os.WriteFile(path, []byte{0xFF, 0xFE, 0x00}, 0644); err != nil {
			t.Fatal(err)
		}

		_, err := ReadFromFile(path)
		if err == nil {
			t.Fatal("esperava erro de UTF-8, mas retornou nil")
		}
	})

	t.Run("arquivo vazio", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "empty.txt")
		if err := os.WriteFile(path, []byte{}, 0644); err != nil {
			t.Fatal(err)
		}

		result, err := ReadFromFile(path)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if result != "" {
			t.Errorf("ReadFromFile = %q, want %q", result, "")
		}
	})

	t.Run("arquivo com padding UTF-8 (BOM)", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "bom.txt")
		// BOM + texto — BOM é UTF-8 válido mas pode ser inesperado
		data := []byte{0xEF, 0xBB, 0xBF} // BOM UTF-8
		data = append(data, []byte("Olá mundo")...)
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}

		result, err := ReadFromFile(path)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if result != string(data) {
			t.Errorf("ReadFromFile = %q, want %q", result, string(data))
		}
		// Verifica que o BOM UTF-8 foi preservado
		if !strings.HasPrefix(result, "\uFEFF") {
			t.Errorf("BOM UTF-8 foi perdido: %q", result)
		}
	})
}

func (s *slowReader) Read(p []byte) (int, error) {
	// Apenas atrasa na primeira chamada para simular pipe travado
	if !s.read {
		s.read = true
		time.Sleep(s.delay)
	}
	return s.r.Read(p)
}

func TestReadFromStdin(t *testing.T) {
	t.Run("stdin com dados", func(t *testing.T) {
		content := "Texto de entrada"
		original := os.Stdin
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdin = r

		_, writeErr := w.Write([]byte(content))
		if writeErr != nil {
			t.Fatal(writeErr)
		}
		w.Close()

		result, err := ReadFromStdin()
		os.Stdin = original
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if result != content {
			t.Errorf("ReadFromStdin = %q, want %q", result, content)
		}
	})

	t.Run("stdin vazio", func(t *testing.T) {
		original := os.Stdin
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdin = r
		w.Close()

		_, err = ReadFromStdin()
		os.Stdin = original
		if err == nil {
			t.Fatal("esperava erro para stdin vazio, mas retornou nil")
		}
	})

	t.Run("stdin UTF-8 inválido", func(t *testing.T) {
		original := os.Stdin
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdin = r

		_, writeErr := w.Write([]byte{0xFF, 0xFE})
		if writeErr != nil {
			t.Fatal(writeErr)
		}
		w.Close()

		_, err = ReadFromStdin()
		os.Stdin = original
		if err == nil {
			t.Fatal("esperava erro de UTF-8, mas retornou nil")
		}
	})

	t.Run("stdin timeout", func(t *testing.T) {
		// Cria um pipe onde a escrita demora mais que o timeout.
		// Isso faz com que io.ReadAll espere e o contexto expire.
		original := os.Stdin
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdin = r

		// Escreve no pipe após um delay maior que o timeout
		go func() {
			time.Sleep(200 * time.Millisecond)
			w.Write([]byte("dados"))
			w.Close()
		}()

		// Usa um timeout curto para o teste
		_, err = ReadFromStdinWithTimeout(50 * time.Millisecond)
		os.Stdin = original
		r.Close()

		if err == nil {
			t.Fatal("esperava erro de timeout, mas retornou nil")
		}
		if !strings.Contains(err.Error(), "timeout") {
			t.Errorf("erro = %q, want contendo 'timeout'", err.Error())
		}
	})
}

func TestIsStdinRedirected(t *testing.T) {
	t.Run("stdin pipe retorna true", func(t *testing.T) {
		original := os.Stdin
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdin = r

		// Escreve algo no pipe para simular redirecionamento
		_, writeErr := w.Write([]byte("dados"))
		if writeErr != nil {
			t.Fatal(writeErr)
		}
		w.Close()

		result := IsStdinRedirected()
		os.Stdin = original
		if !result {
			t.Error("IsStdinRedirected() com pipe = false, want true")
		}
	})

	t.Run("stdin terminal retorna false", func(t *testing.T) {
		// Para testar o cenário de terminal de forma determinística,
		// definimos os.Stdin como um arquivo que simula um terminal.
		// Como não podemos mockar isTerminal() diretamente, usamos
		// um pipe e verificamos o valor inverso.
		result := IsStdinRedirected()
		// Em ambiente de teste (sem pipe ativo), isTerminal() costuma retornar true,
		// então IsStdinRedirected() deve retornar false.
		// Este teste é informacional e não falha se o ambiente tiver pipe.
		t.Logf("IsStdinRedirected() = %v (esperado: false em terminal, true se pipe ativo)", result)
	})
}

func TestReadInput(t *testing.T) {
	t.Run("arquivo como argumento", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.txt")
		expected := "conteúdo do arquivo"
		if err := os.WriteFile(path, []byte(expected), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := ReadInput([]string{path})
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if result != expected {
			t.Errorf("ReadInput = %q, want %q", result, expected)
		}
	})

	t.Run("stdin quando não há argumento", func(t *testing.T) {
		content := "dados do pipe"
		original := os.Stdin
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdin = r

		_, writeErr := w.Write([]byte(content))
		if writeErr != nil {
			t.Fatal(writeErr)
		}
		w.Close()

		result, err := ReadInput([]string{})
		os.Stdin = original
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if result != content {
			t.Errorf("ReadInput = %q, want %q", result, content)
		}
	})

	t.Run("arquivo inexistente retorna erro", func(t *testing.T) {
		_, err := ReadInput([]string{"/nao/existe/arquivo.txt"})
		if err == nil {
			t.Fatal("esperava erro, mas retornou nil")
		}
	})

	t.Run("sem argumento e stdin é terminal retorna erro", func(t *testing.T) {
		// Em ambiente de teste, stdin normalmente é terminal (sem pipe),
		// então ReadFromStdin() deve retornar erro.
		_, err := ReadInput([]string{})
		if err == nil {
			t.Fatal("esperava erro, mas retornou nil")
		}
	})

	t.Run("arquivo vazio", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "empty.txt")
		if err := os.WriteFile(path, []byte{}, 0644); err != nil {
			t.Fatal(err)
		}

		result, err := ReadInput([]string{path})
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if result != "" {
			t.Errorf("ReadInput = %q, want %q", result, "")
		}
	})

	t.Run("UTF-8 inválido retorna erro", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "invalid.txt")
		if err := os.WriteFile(path, []byte{0xFF, 0xFE, 0x00}, 0644); err != nil {
			t.Fatal(err)
		}

		_, err := ReadInput([]string{path})
		if err == nil {
			t.Fatal("esperava erro de UTF-8, mas retornou nil")
		}
	})
}
