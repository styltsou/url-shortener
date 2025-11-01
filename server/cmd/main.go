package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	server "github.com/styltsou/url-shortener/server/pkg"
	"github.com/styltsou/url-shortener/server/pkg/config"
)

func main() {
	// Initialize the server struct which encapsulates router, context, and database pool
	srv := server.NewServer()

	// Connect to the database and create a connection pool
	// This establishes the pgxpool.Pool that will be reused for all database operations
	if err := srv.ConnectDB(config.DatabaseURL); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// defer ensures cleanup happens when main exits (normal or panic)
	// This deferred function will close the database pool connections gracefully
	// Note: This runs in reverse order (LIFO), so if there were multiple defers,
	// they'd execute from bottom to top
	defer func() {
		if err := srv.Shutdown(); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}
	}()

	// Mount handlers, services, and routes
	// This sets up the dependency chain: queries -> services -> handlers -> routes
	if err := srv.MountHandlers(); err != nil {
		log.Fatalf("Failed to mount handlers: %v", err)
	}

	// Create the HTTP server instance
	// http.Server provides more control than http.ListenAndServe directly:
	// - Allows graceful shutdown via Shutdown() method
	// - Can set timeouts (ReadTimeout, WriteTimeout, etc.)
	// - Can use custom listeners
	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(config.Port),
		Handler: srv.Router,
	}

	// Graceful shutdown pattern in Go
	// This goroutine runs concurrently with the HTTP server
	go func() {
		// Create a buffered channel for OS signals
		// Buffer size of 1 prevents missed signals if the channel isn't ready
		sigint := make(chan os.Signal, 1)

		// signal.Notify registers the channel to receive specified signals
		// os.Interrupt = SIGINT (Ctrl+C)
		// syscall.SIGTERM = termination signal (used by process managers like systemd)
		// When these signals are received, they're sent to the sigint channel
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)

		// Block until a signal is received
		// The <- operator reads from the channel, blocking until data arrives
		<-sigint

		log.Println("Shutting down server...")

		// Shutdown gracefully stops the HTTP server
		// - Stops accepting new connections
		// - Waits for existing requests to complete (with context timeout)
		// - Then closes all connections
		// context.Background() gives no timeout (you could use context.WithTimeout here)
		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down server: %v", err)
		}
	}()

	// Start the HTTP server (blocking call)
	// ListenAndServe starts the server and blocks until it stops
	// It returns an error if something goes wrong, OR http.ErrServerClosed
	// if it was gracefully shut down (which is expected, not an error)
	log.Printf("Server starting on port %d", config.Port)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}

	// This log happens after ListenAndServe returns (after graceful shutdown)
	log.Println("Server stopped")
}
