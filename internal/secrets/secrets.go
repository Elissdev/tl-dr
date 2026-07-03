package secrets

import (
	"fmt"
	"os"
	"unsafe"
)

// ProtectedAPIKey wraps an API key and provides a method to zero it out from
// memory when it is no longer needed, reducing the window of exposure.
type ProtectedAPIKey struct {
	data []byte // mantido como []byte para permitir limpeza na memória
}

// stringHeader reflete a representação interna de uma string em Go.
// Usado internamente para acessar o buffer subjacente e zeroá-lo.
type stringHeader struct {
	data unsafe.Pointer
	len  int
}

// LoadAPIKey reads the API key from the TLDR_API_KEY environment variable,
// then immediately zeroes the original string buffer returned by os.Getenv
// to minimize exposure in memory.
// Returns the protected key and any error encountered.
func LoadAPIKey() (*ProtectedAPIKey, error) {
	k := os.Getenv("TLDR_API_KEY")
	if k == "" {
		return nil, fmt.Errorf("TLDR_API_KEY não definida")
	}

	// Cria uma cópia em []byte (que poderá ser zerada via Clear())
	key := &ProtectedAPIKey{data: []byte(k)}

	// Zera o buffer da string original retornada por os.Getenv.
	// Isto é seguro porque:
	// 1. os.Getenv() sempre retorna uma string nova (não uma substring)
	// 2. A string não é compartilhada com outras variáveis neste escopo
	// 3. O buffer é conhecido pelo Header abaixo
	//
	// ATENÇÃO: Esta técnica usa unsafe e depende de detalhes de implementação
	// do Go runtime. É amplamente usada em bibliotecas de segurança (ex: memguard).
	// Se o Go runtime mudar a representação interna de strings, isto quebrará.
	hdr := (*stringHeader)(unsafe.Pointer(&k))
	buf := unsafe.Slice((*byte)(hdr.data), hdr.len)
	for i := range buf {
		buf[i] = 0
	}

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
