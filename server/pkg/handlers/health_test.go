package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func TestHealthHandler_Live(t *testing.T) {
	handler := &HealthHandler{
		logger: createTestLogger(),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	handler.Live(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Live() status = %d, want %d", w.Code, http.StatusOK)
	}

	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response.Status != "ok" {
		t.Errorf("Live() Status = %s, want ok", response.Status)
	}
}

func TestHealthHandler_Ready_NilPool(t *testing.T) {
	handler := &HealthHandler{
		pool:   nil,
		redis:  nil,
		logger: createTestLogger(),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil)
	w := httptest.NewRecorder()
	handler.Ready(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Ready() with nil pool status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}

	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response.Status != healthUnavailable {
		t.Errorf("Ready() Status = %s, want unavailable", response.Status)
	}
	if response.Checks["postgres"] != healthUnavailable {
		t.Errorf("Ready() postgres check = %s, want unavailable", response.Checks["postgres"])
	}
}

func TestHealthHandler_Ready_FastFailPool(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	pool, err := pgxpool.New(ctx, "postgres://localhost:1/nonexistent")
	if err != nil {
		t.Fatalf("pgxpool.New() error = %v", err)
	}
	defer pool.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:        "localhost:1",
		DialTimeout: time.Second,
	})
	defer redisClient.Close()

	handler := &HealthHandler{
		pool:   pool,
		redis:  redisClient,
		logger: createTestLogger(),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil)
	w := httptest.NewRecorder()
	handler.Ready(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Ready() with failing pool status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}

	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response.Checks["postgres"] != healthUnavailable {
		t.Errorf("Ready() postgres check = %s, want unavailable", response.Checks["postgres"])
	}
}

func TestHealthHandler_Ready_NilRedisIsDisabled(t *testing.T) {
	handler := &HealthHandler{
		pool:   nil,
		redis:  nil,
		logger: createTestLogger(),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil)
	w := httptest.NewRecorder()
	handler.Ready(w, req)

	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response.Checks["redis"] != "disabled" {
		t.Errorf("Ready() redis check = %s, want disabled", response.Checks["redis"])
	}
}
