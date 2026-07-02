package main

import (
	"fmt"
	"os"

	"github.com/Elissdev/tl-dr/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
}
