package server

import (
	"context"
	"fmt"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/styltsou/url-shortener/server/pkg/api"
	"github.com/styltsou/url-shortener/server/pkg/config"
	"github.com/styltsou/url-shortener/server/pkg/db"
	"github.com/styltsou/url-shortener/server/pkg/handlers"
	"github.com/styltsou/url-shortener/server/pkg/service"
)

// Server encapsulates the HTTP server, router, database pool, and context
type Server struct {
	Context context.Context
	Router  *chi.Mux
	Pool    *pgxpool.Pool
}

// NewServer creates and initializes a new Server instance
func NewServer() *Server {
	s := &Server{
		Context: context.Background(),
		Router:  chi.NewRouter(),
	}

	config.Load()
	clerk.SetKey(config.ClerkSecretKey)

	return s
}

// ConnectDB establishes a connection to the database
func (s *Server) ConnectDB(connString string) error {
	pool, err := pgxpool.New(s.Context, connString)
	if err != nil {
		return fmt.Errorf("failed to create database pool: %w", err)
	}

	s.Pool = pool
	return nil
}

// MountHandlers sets up routes and middleware
func (s *Server) MountHandlers() error {
	if s.Pool == nil {
		return fmt.Errorf("database pool not initialized, call ConnectDB first")
	}

	queries := db.New(s.Pool)

	linkSvc := service.NewLinkService(queries)
	linkHandler := handlers.NewLinkHandler(linkSvc)

	s.Router.Use(cors.Handler(cors.Options{}))
	s.Router.Use(middleware.RequestID)
	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Recoverer)

	apiRouter := api.NewRouter(linkHandler)
	s.Router.Mount("", apiRouter)

	return nil
}

// Shutdown gracefully shuts down the server and closes resources
func (s *Server) Shutdown() error {
	if s.Pool != nil {
		s.Pool.Close()
	}

	return nil
}
