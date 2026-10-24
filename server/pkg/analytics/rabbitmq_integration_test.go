package analytics

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRabbitConsumerWritesClickBatch(t *testing.T) {
	rabbitURL := os.Getenv("TEST_RABBITMQ_URL")
	clickhouseURL := os.Getenv("TEST_CLICKHOUSE_URL")
	if rabbitURL == "" || clickhouseURL == "" {
		t.Skip("set TEST_RABBITMQ_URL and TEST_CLICKHOUSE_URL for queue integration test")
	}

	log := newTestLogger(t)
	clientInterface, err := New(context.Background(), Config{URL: clickhouseURL}, log)
	if err != nil {
		t.Fatalf("open ClickHouse client: %v", err)
	}
	client := clientInterface.(*Client)
	t.Cleanup(func() { _ = client.Close() })

	consumerCtx, cancelConsumer := context.WithCancel(context.Background())
	consumerDone := make(chan error, 1)
	go func() {
		consumerDone <- RunRabbitConsumer(consumerCtx, rabbitURL, client, log)
	}()
	t.Cleanup(func() {
		cancelConsumer()
		select {
		case err := <-consumerDone:
			if err != nil {
				t.Errorf("stop RabbitMQ consumer: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("RabbitMQ consumer did not stop")
		}
	})

	publisher := newRabbitPublisher(rabbitURL)
	t.Cleanup(func() { _ = publisher.Close() })
	event := ClickEvent{
		EventID:   uuid.New(),
		LinkID:    uuid.New(),
		Timestamp: time.Now().UTC(),
		IP:        "203.0.113.8",
		UserAgent: "integration-test",
		Referrer:  "https://example.com",
	}
	publishCtx, cancelPublish := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPublish()
	if err := publisher.Publish(publishCtx, event); err != nil {
		t.Fatalf("publish click event: %v", err)
	}

	deadline := time.Now().Add(10 * time.Second)
	var lastErr error
	lastTotal := -1
	for time.Now().Before(deadline) {
		analytics, err := client.GetLinkAnalytics(context.Background(), event.LinkID, event.Timestamp.Add(-time.Minute), event.Timestamp.Add(time.Minute))
		lastErr = err
		if analytics != nil {
			lastTotal = analytics.TotalClicks
		}
		if err == nil && analytics.TotalClicks == 1 {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("click event %s was not observed in analytics: total=%d error=%v", event.EventID, lastTotal, lastErr)
}
