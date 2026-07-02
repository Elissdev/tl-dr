package cmd

import (
	"errors"
	"fmt"
	"testing"
)

func TestExitError(t *testing.T) {
	t.Run("NewExitError", func(t *testing.T) {
		err := NewExitError(ExitAPIError, "erro de API")
		if err.Code != ExitAPIError {
			t.Errorf("Code = %d, want %d", err.Code, ExitAPIError)
		}
		if err.Error() != "erro de API" {
			t.Errorf("Error() = %q, want %q", err.Error(), "erro de API")
		}
	})

	t.Run("WrapExitError", func(t *testing.T) {
		original := fmt.Errorf("erro original: %s", "detalhe")
		err := WrapExitError(ExitArgumentError, original)
		if err.Code != ExitArgumentError {
			t.Errorf("Code = %d, want %d", err.Code, ExitArgumentError)
		}
		if !errors.Is(err, original) {
			t.Error("WrapExitError deveria preservar a cadeia de erros")
		}
	})

	t.Run("IsAPIError", func(t *testing.T) {
		err := NewExitError(ExitAPIError, "erro de API")
		if !IsAPIError(err) {
			t.Error("IsAPIError(ExitAPIError) = false, want true")
		}

		genericErr := NewExitError(ExitGenericError, "erro genérico")
		if IsAPIError(genericErr) {
			t.Error("IsAPIError(ExitGenericError) = true, want false")
		}

		if IsAPIError(fmt.Errorf("erro comum")) {
			t.Error("IsAPIError(erro comum) = true, want false")
		}
	})

	t.Run("IsArgumentError", func(t *testing.T) {
		err := NewExitError(ExitArgumentError, "erro de argumento")
		if !IsArgumentError(err) {
			t.Error("IsArgumentError(ExitArgumentError) = false, want true")
		}

		apiErr := NewExitError(ExitAPIError, "erro de API")
		if IsArgumentError(apiErr) {
			t.Error("IsArgumentError(ExitAPIError) = true, want false")
		}
	})

	t.Run("Unwrap", func(t *testing.T) {
		original := fmt.Errorf("original error")
		err := WrapExitError(ExitGenericError, original)
		unwrapped := errors.Unwrap(err)
		if unwrapped != original {
			t.Errorf("Unwrap() = %v, want %v", unwrapped, original)
		}
	})
}

func TestExitCodes(t *testing.T) {
	if ExitSuccess != 0 {
		t.Errorf("ExitSuccess = %d, want 0", ExitSuccess)
	}
	if ExitGenericError != 1 {
		t.Errorf("ExitGenericError = %d, want 1", ExitGenericError)
	}
	if ExitAPIError != 2 {
		t.Errorf("ExitAPIError = %d, want 2", ExitAPIError)
	}
	if ExitArgumentError != 3 {
		t.Errorf("ExitArgumentError = %d, want 3", ExitArgumentError)
	}
}
