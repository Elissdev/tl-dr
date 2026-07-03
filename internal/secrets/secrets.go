package secrets

import (
	"fmt"
	"os"
)

// ProtectedAPIKey encapsula uma chave de API e fornece um método para
// zerá-la da memória quando não for mais necessária, reduzindo a janela
// de exposição da credencial.
type ProtectedAPIKey struct {
	data []byte // mantido como []byte para permitir limpeza na memória
}

// LoadAPIKey lê a chave de API da variável de ambiente TLDR_API_KEY.
// Retorna a chave protegida (envolvida em um []byte que pode ser zerado
// posteriormente via Clear()) e qualquer erro encontrado.
//
// NOTA: Em Go, não é possível zerar de forma confiável o buffer interno
// da string retornada por os.Getenv(), pois o runtime pode compartilhar
// o buffer com outras variáveis de ambiente. Em vez disso, copiamos o
// valor para um []byte controlado que pode ser zerado via Clear().
func LoadAPIKey() (*ProtectedAPIKey, error) {
	k := os.Getenv("TLDR_API_KEY")
	if k == "" {
		return nil, fmt.Errorf("TLDR_API_KEY não definida")
	}

	// Cria uma cópia em []byte (que poderá ser zerada via Clear())
	key := &ProtectedAPIKey{data: []byte(k)}

	return key, nil
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
