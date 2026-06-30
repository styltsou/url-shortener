// Package server implements the Server abstraction
package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/styltsou/url-shortener/server/pkg/analytics"
	"github.com/styltsou/url-shortener/server/pkg/config"
	"github.com/styltsou/url-shortener/server/pkg/db"
	"github.com/styltsou/url-shortener/server/pkg/handlers"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	"github.com/styltsou/url-shortener/server/pkg/middleware"
	"github.com/styltsou/url-shortener/server/pkg/router"
	"github.com/styltsou/url-shortener/server/pkg/service"
	"go.uber.org/zap"
)

// sanitizeDSN redacts the password portion of a connection string for safe logging.
// Handles formats like:
//
//	postgres://user:pass@host/db         → postgres://user:****@host/db
//	postgres://user@host/db              → postgres://user@host/db (no password)
//	redis://:password@host:6379          → redis://:****@host:6379
//	redis://user:password@host:6379      → redis://user:****@host:6379
func sanitizeDSN(dsn string) string {
	schemeIdx := strings.Index(dsn, "://")
	if schemeIdx < 0 {
		return dsn
	}
	scheme := dsn[:schemeIdx]
	rest := dsn[schemeIdx+3:]

	atIdx := strings.Index(rest, "@")
	if atIdx < 0 {
		return dsn
	}
	credentials := rest[:atIdx]
	hostPart := rest[atIdx:]

	colonIdx := strings.Index(credentials, ":")
	if colonIdx < 0 {
		return dsn
	}
	// For redis://:password@host, user is empty so the colon is at position 0
	redactedUser := credentials[:colonIdx+1]
	return scheme + "://" + redactedUser + "****" + hostPart
}

// Server encapsulates the HTTP server, router, database pool, and context
type Server struct {
	Context         context.Context
	Pool            *pgxpool.Pool
	RedisClient     *redis.Client
	AnalyticsClient analytics.ClientInterface
	Router          *chi.Mux
	Logger          logger.Logger
}

// New creates and initializes a new Server instance
// It automatically connects to the database and mounts handlers
// Logger and config should be initialized in the caller (main.go)
func New(config *config.Config, log logger.Logger) (*Server, error) {
	clerk.SetKey(config.ClerkSecretKey)

	s := &Server{
		Context: context.Background(),
		Router:  chi.NewRouter(),
		Logger:  log,
	}

	pool, pgErr := pgxpool.New(s.Context, config.PostgresConnectionString)

	if pgErr != nil {
		return nil, fmt.Errorf("failed to create Postgres pool: %w", pgErr)
	}
	s.Pool = pool
	log.Info("Postgres connected successfully",
		zap.String("pg_connection_str", sanitizeDSN(config.PostgresConnectionString)),
	)

	// Try to connect to Redis, but don't fail if it's unavailable (degraded mode)
	rdb := redis.NewClient(&redis.Options{
		Addr:         config.RedisURL,
		Username:     config.RedisUsername,
		Password:     config.RedisPassword,
		DB:           config.RedisDB,
		MaxRetries:   config.RedisMaxRetries,
		DialTimeout:  time.Duration(config.RedisDialTimeout) * time.Second,
		ReadTimeout:  time.Duration(config.RedisReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.RedisWriteTimeout) * time.Second,
	})

	// Ping Redis with a timeout to avoid hanging
	pingCtx, cancel := context.WithTimeout(s.Context, 3*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		s.RedisClient = nil
		log.Warn("Redis connection failed, running without cache",
			zap.Error(err),
			zap.String("redis_url", sanitizeDSN(config.RedisURL)),
		)
	} else {
		s.RedisClient = rdb
		log.Info("Redis connected successfully",
			zap.String("redis_url", sanitizeDSN(config.RedisURL)),
		)
	}

	queries := db.New(s.Pool)

	var analyticsErr error
	s.AnalyticsClient, analyticsErr = analytics.New(s.Context, analytics.Config{
		URL:             config.ClickhouseURL,
		Username:        config.ClickhouseUsername,
		Password:        config.ClickhousePassword,
		DialTimeout:     time.Duration(config.ClickhouseDialTimeout) * time.Second,
		MaxOpenConns:    config.ClickhouseMaxOpenConns,
		MaxIdleConns:    config.ClickhouseMaxIdleConns,
		ConnMaxLifetime: time.Duration(config.ClickhouseConnMaxLifeMin) * time.Minute,
		TableName:       config.ClickhouseTableName,
	}, s.Logger)
	if analyticsErr != nil {
		log.Warn("ClickHouse analytics disabled",
			zap.Error(analyticsErr),
		)
	}

	linkSvc := service.NewLinkService(s.Pool, queries, s.RedisClient, s.AnalyticsClient, s.Logger)
	linkHandler := handlers.NewLinkHandler(linkSvc, config.BaseURL, s.Logger)

	tagSvc := service.NewTagService(queries, s.Logger)
	tagHandler := handlers.NewTagHandler(tagSvc, s.Logger)

	s.Router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   config.CORSAllowedOrigins,
		AllowedMethods:   config.CORSAllowedMethods,
		AllowedHeaders:   config.CORSAllowedHeaders,
		ExposedHeaders:   config.CORSExposedHeaders,
		AllowCredentials: config.CORSAllowCredentials,
		MaxAge:           config.CORSMaxAge,
	}))
	s.Router.Use(chimw.RequestID)
	s.Router.Use(middleware.RequestIDHeader)
	s.Router.Use(middleware.SecurityHeaders)
	s.Router.Use(middleware.RequestLogger(s.Logger))
	s.Router.Use(chimw.Recoverer)
	s.Router.Use(middleware.RateLimit(middleware.RateLimitConfig{
		Requests: config.RateLimitRequests,
		Window:   time.Duration(config.RateLimitWindowSeconds) * time.Second,
	}))

	healthHandler := handlers.NewHealthHandler(s.Pool, s.RedisClient, s.Logger)
	apiRouter := router.New(linkHandler, tagHandler, healthHandler, s.Logger)
	s.Router.Mount("/", apiRouter)

	return s, nil
}

func (s *Server) CloseConnections() {
	if s.Pool != nil {
		s.Pool.Close()
	}

	if s.RedisClient != nil {
		if err := s.RedisClient.Close(); err != nil {
			s.Logger.Error("Error closing Redis pool",
				zap.Error(err),
			)
		}
	}

	if err := s.AnalyticsClient.Close(); err != nil {
		s.Logger.Error("Error closing ClickHouse connection",
			zap.Error(err),
		)
	}
}
