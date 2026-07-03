package cmd

import (
	"errors"
)

// ExitCode representa um código de saída para o programa.
type ExitCode int

const (
	ExitOK       ExitCode = 0
	ExitInternal ExitCode = 1
	ExitAPI      ExitCode = 2
	ExitArgs     ExitCode = 3
	ExitTimeout  ExitCode = 4
)

// ExitError é um erro que carrega um código de saída específico.
// Isso permite que o main.go mapeie o erro para o exit code correto.
type ExitError struct {
	Code ExitCode
	Err  error
}

func (e *ExitError) Error() string {
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error {
	return e.Err
}

// NewExitError cria um novo ExitError com o código e mensagem fornecidos.
// Use esta função quando você tem uma mensagem estática (sem erro encapsulado).
// Exemplo: return NewExitError(ExitArgs, "idioma é obrigatório")
func NewExitError(code ExitCode, msg string) *ExitError {
	return &ExitError{Code: code, Err: errors.New(msg)}
}

// WrapExitError envolve um erro existente com um código de saída.
// Use esta função quando você já tem um erro de outra camada (ex: I/O, API).
// Exemplo: return WrapExitError(ExitInternal, err)
func WrapExitError(code ExitCode, err error) *ExitError {
	return &ExitError{Code: code, Err: err}
}

// IsAPIError verifica se um erro é do tipo ExitError com código de API.
func IsAPIError(err error) bool {
	var e *ExitError
	if errors.As(err, &e) {
		return e.Code == ExitAPI
	}
	return false
}

// IsArgumentError verifica se um erro é do tipo ExitError com código de argumento.
func IsArgumentError(err error) bool {
	var e *ExitError
	if errors.As(err, &e) {
		return e.Code == ExitArgs
	}
	return false
}
