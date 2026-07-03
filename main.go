package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/Elissdev/tl-dr/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)

		var exitErr *cmd.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(int(exitErr.Code))
		}
		// Se o erro não for um ExitError (ex: panic recuperado inesperado),
		// usa código genérico 1 como fallback.
		os.Exit(1)
	}
}
