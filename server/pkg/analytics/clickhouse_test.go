package analytics

import (
	"context"
	"testing"

	"github.com/styltsou/url-shortener/server/pkg/logger"
)

func TestNewReturnsNopClientWhenURLMissing(t *testing.T) {
	log := newTestLogger(t)

	client, err := New(context.Background(), Config{}, log)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	if _, ok := client.(*nopClient); !ok {
		t.Fatalf("New() client = %T, want *nopClient", client)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v, want nil", err)
	}
}

func newTestLogger(t *testing.T) logger.Logger {
	t.Helper()

	log, err := logger.New("test")
	if err != nil {
		t.Fatalf("logger.New() error = %v", err)
	}
	return log
}
