package secrets

import (
	"fmt"
	"os"
)

// ProtectedAPIKey wraps an API key and provides a method to zero it out from
// memory when it is no longer needed, reducing the window of exposure.
type ProtectedAPIKey struct {
	data []byte // mantido como []byte para permitir limpeza na memória
}

// LoadAPIKey reads the API key from the TLDR_API_KEY environment variable.
// Returns the protected key and any error encountered.
func LoadAPIKey() (*ProtectedAPIKey, error) {
	k := os.Getenv("TLDR_API_KEY")
	if k == "" {
		return nil, fmt.Errorf("TLDR_API_KEY não definida")
	}
	return &ProtectedAPIKey{data: []byte(k)}, nil
}

// Get retorna a chave de API como string. A string retornada é uma cópia;
// o buffer interno do ProtectedAPIKey permanece intacto até Clear() ser chamado.
func (p *ProtectedAPIKey) Get() string {
	return string(p.data)
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
