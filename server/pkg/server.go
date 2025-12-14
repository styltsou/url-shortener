// Package server implements the Server abstraction
package server

import (
	"context"
	"fmt"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/styltsou/url-shortener/server/pkg/config"
	"github.com/styltsou/url-shortener/server/pkg/db"
	"github.com/styltsou/url-shortener/server/pkg/handlers"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	"github.com/styltsou/url-shortener/server/pkg/middleware"
	"github.com/styltsou/url-shortener/server/pkg/router"
	"github.com/styltsou/url-shortener/server/pkg/service"
	"go.uber.org/zap"
)

// Server encapsulates the HTTP server, router, database pool, and context
type Server struct {
	Context     context.Context
	Pool        *pgxpool.Pool
	RedisClient *redis.Client
	Router      *chi.Mux
	Logger      logger.Logger
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
		zap.String("pg_connection_str", config.PostgresConnectionString),
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

	// TODO: Need to understand this better
	// Ping Redis with a timeout to avoid hanging
	pingCtx, cancel := context.WithTimeout(s.Context, 3*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		s.RedisClient = nil
		log.Warn("Redis connection failed, running without cache",
			zap.Error(err),
			zap.String("redis_url", config.RedisURL),
		)
	} else {
		s.RedisClient = rdb
		log.Info("Redis connected successfully",
			zap.String("redis_url", config.RedisURL),
		)
	}

	queries := db.New(s.Pool)
	linkSvc := service.NewLinkService(queries, s.RedisClient, s.Logger)
	linkHandler := handlers.NewLinkHandler(linkSvc, s.Logger)

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
	s.Router.Use(middleware.RequestLogger(s.Logger))
	s.Router.Use(chimw.Recoverer)

	apiRouter := router.New(linkHandler, tagHandler, s.Logger)
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
}
