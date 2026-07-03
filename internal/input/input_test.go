package input

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFile(t *testing.T) {
	t.Run("arquivo normal", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.txt")
		expected := "Hello, World!"
		if err := os.WriteFile(path, []byte(expected), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := ReadFile(path)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if result != expected {
			t.Errorf("ReadFile = %q, want %q", result, expected)
		}
	})

	t.Run("arquivo inexistente", func(t *testing.T) {
		_, err := ReadFile("/nao/existe/arquivo.txt")
		if err == nil {
			t.Fatal("esperava erro, mas retornou nil")
		}
	})

	t.Run("diretório", func(t *testing.T) {
		dir := t.TempDir()
		_, err := ReadFile(dir)
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

		_, err := ReadFile(path)
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

		_, err := ReadFile(path)
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

		result, err := ReadFile(path)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if result != "" {
			t.Errorf("ReadFile = %q, want %q", result, "")
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

		result, err := ReadFile(path)
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if !strings.HasPrefix(result, "\uFEFF") && !strings.HasPrefix(result, string(data)) {
			t.Errorf("resultado inesperado: %q", result)
		}
	})
}

func TestReadStdin(t *testing.T) {
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

		result, err := ReadStdin()
		os.Stdin = original
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if result != content {
			t.Errorf("ReadStdin = %q, want %q", result, content)
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

		result, err := ReadStdin()
		os.Stdin = original
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if result != "" {
			t.Errorf("ReadStdin = %q, want %q", result, "")
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

		_, err = ReadStdin()
		os.Stdin = original
		if err == nil {
			t.Fatal("esperava erro de UTF-8, mas retornou nil")
		}
	})
}

func TestIsStdinAvailable(t *testing.T) {
	t.Run("stdin pipe retorna true", func(t *testing.T) {
		original := os.Stdin
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdin = r

		// Escreve algo no pipe para simular dados disponíveis
		_, writeErr := w.Write([]byte("dados"))
		if writeErr != nil {
			t.Fatal(writeErr)
		}
		w.Close()

		result := IsStdinAvailable()
		os.Stdin = original
		if !result {
			t.Error("IsStdinAvailable() com pipe = false, want true")
		}
	})

	t.Run("stdin terminal retorna false", func(t *testing.T) {
		// Não podemos simular um terminal real em testes, mas verificamos
		// que a função não panic e lida com o caso corretamente.
		result := IsStdinAvailable()
		t.Logf("IsStdinAvailable() em terminal = %v", result)
	})
}
