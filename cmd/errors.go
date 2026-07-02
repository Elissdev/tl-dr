package cmd

import (
	"errors"
	"fmt"
)

// ExitCode representa um código de saída para o programa.
type ExitCode int

const (
	ExitSuccess       ExitCode = 0
	ExitGenericError  ExitCode = 1
	ExitAPIError      ExitCode = 2
	ExitArgumentError ExitCode = 3
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
func NewExitError(code ExitCode, msg string) *ExitError {
	return &ExitError{Code: code, Err: fmt.Errorf("%s", msg)}
}

// WrapExitError envolve um erro existente com um código de saída.
func WrapExitError(code ExitCode, err error) *ExitError {
	return &ExitError{Code: code, Err: err}
}

// IsAPIError verifica se um erro é do tipo ExitError com código de API.
func IsAPIError(err error) bool {
	var e *ExitError
	if errors.As(err, &e) {
		return e.Code == ExitAPIError
	}
	return false
}

// IsArgumentError verifica se um erro é do tipo ExitError com código de argumento.
func IsArgumentError(err error) bool {
	var e *ExitError
	if errors.As(err, &e) {
		return e.Code == ExitArgumentError
	}
	return false
}
