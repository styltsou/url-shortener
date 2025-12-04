package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	// Test that sentinel errors are defined and can be checked with errors.Is
	if LinkNotFound == nil {
		t.Error("ErrLinkNotFound should not be nil")
	}

	if InvalidURL == nil {
		t.Error("ErrInvalidURL should not be nil")
	}

	// Test that errors.Is works with wrapped errors
	wrappedErr := fmt.Errorf("context: %w", LinkNotFound)
	if !errors.Is(wrappedErr, LinkNotFound) {
		t.Error("errors.Is() should return true for wrapped ErrLinkNotFound")
	}

	wrappedInvalidURL := fmt.Errorf("context: %w", InvalidURL)
	if !errors.Is(wrappedInvalidURL, InvalidURL) {
		t.Error("errors.Is() should return true for wrapped ErrInvalidURL")
	}
}
