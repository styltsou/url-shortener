package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/styltsou/url-shortener/server/pkg/analytics"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	clickhouseURL := os.Getenv("CLICKHOUSE_URL")
	if rabbitURL == "" || clickhouseURL == "" {
		fmt.Fprintln(os.Stderr, "RABBITMQ_URL and CLICKHOUSE_URL are required")
		os.Exit(1)
	}
	log, err := logger.New(os.Getenv("APP_ENV"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer log.Sync()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	for ctx.Err() == nil {
		client, err := analytics.New(ctx, analytics.Config{
			URL:       clickhouseURL,
			Username:  os.Getenv("CLICKHOUSE_USERNAME"),
			Password:  os.Getenv("CLICKHOUSE_PASSWORD"),
			TableName: os.Getenv("CLICKHOUSE_TABLE_NAME"),
		}, log)
		if err == nil {
			writer, ok := client.(analytics.ClickBatchWriter)
			if !ok {
				log.Error("ClickHouse client does not support batch inserts")
				_ = client.Close()
				return
			}
			log.Info("Click consumer ready")
			err = analytics.RunRabbitConsumer(ctx, rabbitURL, writer, log)
			_ = client.Close()
			if err != nil && ctx.Err() == nil {
				log.Warn("Click consumer restarting", zap.Error(err))
			}
		} else {
			log.Warn("ClickHouse unavailable; worker retrying", zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}
