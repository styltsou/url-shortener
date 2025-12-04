package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	server "github.com/styltsou/url-shortener/server/pkg"
	"github.com/styltsou/url-shortener/server/pkg/config"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfg, cfgErr := config.Load()
	if cfgErr != nil {
		fmt.Println(cfgErr.Error())
		os.Exit(1)
	}

	log, logErr := logger.New(cfg.AppEnv)
	if logErr != nil {
		fmt.Println(logErr.Error())
		os.Exit(1)
	}

	defer func() {
		_ = log.Sync() // Flush logs on exit
	}()

	srv, err := server.New(cfg, log)
	if err != nil {
		log.Fatal("Failed to initialize server",
			zap.Error(err),
		)
	}

	httpServer := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      srv.Router,
		ReadTimeout:  time.Duration(cfg.ServerReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.ServerWriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.ServerIdleTimeout) * time.Second,
	}

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Info("Shutting down server...")

		// TODO: Need to get more comfortable with what this does
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			log.Error("Error while shutting down server",
				zap.Error(err),
			)
		}

		srv.CloseConnections()
	}()

	log.Info("Server start",
		zap.Int("port", cfg.Port),
		zap.String("env", cfg.AppEnv),
	)

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed",
			zap.Error(err),
		)
	}

	log.Info("Server stopped")
}
