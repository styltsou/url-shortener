package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func TestHealthHandler_ReadyRequiresSchema(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL for PostgreSQL integration test")
	}
	ctx := context.Background()
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
		_ = admin.Close(context.Background())
	})
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	handler := NewHealthHandler(pool, nil, createTestLogger())
	readyStatus := func() int {
		response := httptest.NewRecorder()
		handler.Ready(response, httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil))
		return response.Code
	}
	if got := readyStatus(); got != http.StatusServiceUnavailable {
		t.Fatalf("missing schema status = %d, want 503", got)
	}
	for _, table := range []string{"links", "tags", "link_tags"} {
		if _, err := pool.Exec(ctx, "CREATE TABLE "+table+" (id integer)"); err != nil {
			t.Fatal(err)
		}
	}
	if got := readyStatus(); got != http.StatusOK {
		t.Fatalf("migrated schema status = %d, want 200", got)
	}
}
