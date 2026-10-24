package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	"go.uber.org/zap"
)

const (
	clickBatchSize = 500
	clickBatchWait = time.Second
)

type ClickBatchWriter interface {
	InsertBatch(context.Context, []ClickEvent) error
}

// RunRabbitConsumer processes one queue consumer. Several worker replicas may run
// concurrently; RabbitMQ distributes deliveries between them.
func RunRabbitConsumer(ctx context.Context, url string, writer ClickBatchWriter, log logger.Logger) error {
	for ctx.Err() == nil {
		if err := consumeClicks(ctx, url, writer, log); err != nil && ctx.Err() == nil {
			log.Warn("Click consumer disconnected; reconnecting", zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(2 * time.Second):
		}
	}
	return nil
}

func consumeClicks(ctx context.Context, url string, writer ClickBatchWriter, log logger.Logger) error {
	conn, err := openRabbit(url)
	if err != nil {
		return err
	}
	// Closing the connection also closes its channel and requeues any
	// unacknowledged deliveries. Use a deadline because workers must be able to
	// stop promptly when a deployment replaces them or RabbitMQ is unavailable.
	defer func() { _ = conn.CloseDeadline(time.Now().Add(time.Second)) }()
	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	if err := declareClickQueues(channel); err != nil {
		return err
	}
	if err := channel.Qos(clickBatchSize, 0, false); err != nil {
		return fmt.Errorf("set click consumer prefetch: %w", err)
	}
	deliveries, err := channel.ConsumeWithContext(ctx, clickQueueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume clicks: %w", err)
	}
	ticker := time.NewTicker(clickBatchWait)
	defer ticker.Stop()
	batch := make([]amqp.Delivery, 0, clickBatchSize)
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		events := make([]ClickEvent, 0, len(batch))
		valid := make([]amqp.Delivery, 0, len(batch))
		for _, delivery := range batch {
			var event ClickEvent
			if err := json.Unmarshal(delivery.Body, &event); err != nil || event.EventID == uuid.Nil || event.LinkID == uuid.Nil || event.Timestamp.IsZero() {
				log.Warn("Invalid click message sent to dead-letter queue", zap.Error(err))
				_ = delivery.Nack(false, false)
				continue
			}
			events = append(events, event)
			valid = append(valid, delivery)
		}
		batch = batch[:0]
		if len(events) == 0 {
			return nil
		}
		insertCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		err := writer.InsertBatch(insertCtx, events)
		cancel()
		if err != nil {
			log.Warn("ClickHouse batch failed; RabbitMQ will retry", zap.Int("count", len(events)), zap.Error(err))
			for _, delivery := range valid {
				_ = delivery.Nack(false, true)
			}
			return err
		}
		for _, delivery := range valid {
			if err := delivery.Ack(false); err != nil {
				log.Warn("Click batch stored but RabbitMQ acknowledgement failed", zap.Error(err))
			}
		}
		log.Info("Click batch stored", zap.Int("count", len(events)))
		return nil
	}
	for {
		select {
		case <-ctx.Done():
			for _, delivery := range batch {
				_ = delivery.Nack(false, true)
			}
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				for _, pending := range batch {
					_ = pending.Nack(false, true)
				}
				return fmt.Errorf("click delivery channel closed")
			}
			batch = append(batch, delivery)
			if len(batch) >= clickBatchSize {
				if err := flush(); err != nil {
					return err
				}
			}
		case <-ticker.C:
			if err := flush(); err != nil {
				return err
			}
		}
	}
}
