package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	"go.uber.org/zap"
)

type HealthHandler struct {
	pool   *pgxpool.Pool
	redis  *redis.Client
	logger logger.Logger
}

type HealthResponse struct {
	Status   string            `json:"status"`
	Service  string            `json:"service"`
	Checks   map[string]string `json:"checks,omitempty"`
	Degraded bool              `json:"degraded,omitempty"`
}

func NewHealthHandler(pool *pgxpool.Pool, redis *redis.Client, logger logger.Logger) *HealthHandler {
	return &HealthHandler{pool: pool, redis: redis, logger: logger}
}

func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, HealthResponse{
		Status:  "ok",
		Service: "URL Shortener API",
	})
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	checks := map[string]string{
		"postgres": "ok",
		"redis":    "disabled",
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	status := http.StatusOK
	overall := "ok"
	degraded := false

	if h.pool == nil {
		checks["postgres"] = "unavailable"
		status = http.StatusServiceUnavailable
		overall = "unavailable"
	} else if err := h.pool.Ping(ctx); err != nil {
		checks["postgres"] = "unavailable"
		status = http.StatusServiceUnavailable
		overall = "unavailable"
		h.logger.Error("Postgres readiness check failed", zap.Error(err))
	}

	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			checks["redis"] = "degraded"
			degraded = true
			h.logger.Warn("Redis readiness check failed", zap.Error(err))
		} else {
			checks["redis"] = "ok"
		}
	}

	render.Status(r, status)
	render.JSON(w, r, HealthResponse{
		Status:   overall,
		Service:  "URL Shortener API",
		Checks:   checks,
		Degraded: degraded,
	})
}
