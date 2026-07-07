package cmd

import (
	"errors"
	"testing"
)

func TestExitError(t *testing.T) {
	t.Run("NewExitError", func(t *testing.T) {
		err := NewExitError(ExitAPI, "erro de API")
		if err.Code != ExitAPI {
			t.Errorf("Code = %d, want %d", err.Code, ExitAPI)
		}
		if err.Error() != "erro de API" {
			t.Errorf("Error() = %q, want %q", err.Error(), "erro de API")
		}
	})

	t.Run("WrapExitError", func(t *testing.T) {
		original := errors.New("erro original: detalhe")
		err := WrapExitError(ExitArgs, original)
		if err.Code != ExitArgs {
			t.Errorf("Code = %d, want %d", err.Code, ExitArgs)
		}
		if !errors.Is(err, original) {
			t.Error("WrapExitError deveria preservar a cadeia de erros")
		}
	})

	t.Run("IsAPIError", func(t *testing.T) {
		err := NewExitError(ExitAPI, "erro de API")
		if !IsAPIError(err) {
			t.Error("IsAPIError(ExitAPI) = false, want true")
		}

		genericErr := NewExitError(ExitInternal, "erro interno")
		if IsAPIError(genericErr) {
			t.Error("IsAPIError(ExitInternal) = true, want false")
		}

		if IsAPIError(errors.New("erro comum")) {
			t.Error("IsAPIError(erro comum) = true, want false")
		}
	})

	t.Run("IsArgumentError", func(t *testing.T) {
		err := NewExitError(ExitArgs, "erro de argumento")
		if !IsArgumentError(err) {
			t.Error("IsArgumentError(ExitArgs) = false, want true")
		}

		apiErr := NewExitError(ExitAPI, "erro de API")
		if IsArgumentError(apiErr) {
			t.Error("IsArgumentError(ExitAPI) = true, want false")
		}
	})

	t.Run("Unwrap", func(t *testing.T) {
		original := errors.New("original error")
		err := WrapExitError(ExitInternal, original)
		unwrapped := errors.Unwrap(err)
		if unwrapped != original {
			t.Errorf("Unwrap() = %v, want %v", unwrapped, original)
		}
	})
}

func TestExitCodes(t *testing.T) {
	if ExitOK != 0 {
		t.Errorf("ExitOK = %d, want 0", ExitOK)
	}
	if ExitInternal != 1 {
		t.Errorf("ExitInternal = %d, want 1", ExitInternal)
	}
	if ExitAPI != 2 {
		t.Errorf("ExitAPI = %d, want 2", ExitAPI)
	}
	if ExitArgs != 3 {
		t.Errorf("ExitArgs = %d, want 3", ExitArgs)
	}
	if ExitTimeout != 4 {
		t.Errorf("ExitTimeout = %d, want 4", ExitTimeout)
	}
}
