package secrets

import (
	"fmt"
	"os"
	"path/filepath"
)

// ProtectedAPIKey encapsula uma chave de API e fornece um método para
// zerá-la da memória quando não for mais necessária, reduzindo a janela
// de exposição da credencial.
type ProtectedAPIKey struct {
	data []byte // mantido como []byte para permitir limpeza na memória
}

// LoadAPIKey lê a chave de API da variável de ambiente TLDR_API_KEY
// ou do arquivo apontado por TLDR_API_KEY_FILE.
// A precedence é: TLDR_API_KEY (diretamente) > TLDR_API_KEY_FILE.
// Retorna a chave protegida (envolvida em um []byte que pode ser zerado
// posteriormente via Clear()) e qualquer erro encontrado.
func LoadAPIKey() (*ProtectedAPIKey, error) {
	// 1. Tenta ler diretamente da variável de ambiente
	if k := os.Getenv("TLDR_API_KEY"); k != "" {
		// Cria uma cópia em []byte (que poderá ser zerada via Clear())
		// A string k é cópia do buffer interno do runtime; não a zeramos
		// pois afetaria o environment do processo. A cópia em []byte é
		// nossa responsabilidade.
		return &ProtectedAPIKey{data: []byte(k)}, nil
	}

	// 2. Tenta ler de TLDR_API_KEY_FILE
	if path := os.Getenv("TLDR_API_KEY_FILE"); path != "" {
		cleanPath := filepath.Clean(path)
		data, err := os.ReadFile(cleanPath)
		if err != nil {
			return nil, fmt.Errorf("TLDR_API_KEY_FILE %q: %w", path, err)
		}
		// Remove trailing newline/carriage return
		if len(data) > 0 && data[len(data)-1] == '\n' {
			data = data[:len(data)-1]
		}
		if len(data) > 0 && data[len(data)-1] == '\r' {
			data = data[:len(data)-1]
		}
		if len(data) == 0 {
			return nil, fmt.Errorf("TLDR_API_KEY_FILE %q está vazio", path)
		}
		return &ProtectedAPIKey{data: data}, nil
	}

	return nil, fmt.Errorf("TLDR_API_KEY não definida (defina a variável ou TLDR_API_KEY_FILE)")
}

// Get retorna a chave de API como string. A string retornada é uma cópia;
// o buffer interno do ProtectedAPIKey permanece intacto até Clear() ser chamado.
func (p *ProtectedAPIKey) Get() string {
	return string(p.data)
}

// Bytes retorna uma cópia do slice interno da chave.
// O caller pode modificar ou zerar a cópia sem afetar o estado interno
// do ProtectedAPIKey. Retorna nil se o wrapper já foi limpo via Clear().
func (p *ProtectedAPIKey) Bytes() []byte {
	if p == nil || p.data == nil {
		return nil
	}
	b := make([]byte, len(p.data))
	copy(b, p.data)
	return b
}

// Clear zera os bytes da chave na memória e invalida o wrapper.
// Após chamar Clear o wrapper não deve mais ser usado.
// Esta função é nil-safe: chamar Clear() em um ponteiro nil é seguro.
func (p *ProtectedAPIKey) Clear() {
	if p == nil {
		return
	}
	for i := range p.data {
		p.data[i] = 0
	}
	p.data = nil
}
