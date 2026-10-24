package analytics

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	clickQueueName           = "link4it.clicks"
	clickDeadLetterQueueName = "link4it.clicks.dead"
)

type eventPublisher interface {
	Publish(context.Context, ClickEvent) error
	Close() error
}

type rabbitPublisher struct {
	url     string
	mu      sync.Mutex
	conn    *amqp.Connection
	channel *amqp.Channel
	returns chan amqp.Return
}

func newRabbitPublisher(url string) *rabbitPublisher {
	return &rabbitPublisher{url: url}
}

type writeTimeoutConn struct{ net.Conn }

func (c writeTimeoutConn) Write(data []byte) (int, error) {
	if err := c.SetWriteDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return 0, err
	}
	return c.Conn.Write(data)
}

func openRabbit(url string) (*amqp.Connection, error) {
	dial := amqp.DefaultDial(3 * time.Second)
	return amqp.DialConfig(url, amqp.Config{
		Dial: func(network, address string) (net.Conn, error) {
			conn, err := dial(network, address)
			if err != nil {
				return nil, err
			}
			return writeTimeoutConn{conn}, nil
		},
		Heartbeat: 10 * time.Second,
	})
}

func declareClickQueues(channel *amqp.Channel) error {
	if _, err := channel.QueueDeclare(clickDeadLetterQueueName, true, false, false, false,
		amqp.Table{"x-queue-type": "quorum"}); err != nil {
		return fmt.Errorf("declare click dead-letter queue: %w", err)
	}
	if _, err := channel.QueueDeclare(clickQueueName, true, false, false, false,
		amqp.Table{
			"x-queue-type":              "quorum",
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": clickDeadLetterQueueName,
			"x-dead-letter-strategy":    "at-least-once",
			"x-overflow":                "reject-publish",
			"x-delivery-limit":          int32(1000),
		}); err != nil {
		return fmt.Errorf("declare click queue: %w", err)
	}
	return nil
}

func (p *rabbitPublisher) connect() error {
	if p.channel != nil && !p.channel.IsClosed() && p.conn != nil && !p.conn.IsClosed() {
		return nil
	}
	_ = p.closeLocked()
	conn, err := openRabbit(p.url)
	if err != nil {
		return fmt.Errorf("connect to RabbitMQ: %w", err)
	}
	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("open RabbitMQ channel: %w", err)
	}
	if err := declareClickQueues(channel); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return err
	}
	if err := channel.Confirm(false); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return fmt.Errorf("enable publisher confirms: %w", err)
	}
	p.conn = conn
	p.channel = channel
	p.returns = channel.NotifyReturn(make(chan amqp.Return, 1))
	return nil
}

func (p *rabbitPublisher) Publish(ctx context.Context, event ClickEvent) error {
	if event.EventID == uuid.Nil {
		return errors.New("click event has no event ID")
	}
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode click event: %w", err)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.connect(); err != nil {
		return err
	}
	confirmation, err := p.channel.PublishWithDeferredConfirmWithContext(ctx, "", clickQueueName, true, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    event.EventID.String(),
		Timestamp:    event.Timestamp,
		Type:         "click.recorded.v1",
		Body:         body,
	})
	if err != nil {
		_ = p.closeLocked()
		return fmt.Errorf("publish click event: %w", err)
	}
	acknowledged, err := confirmation.WaitContext(ctx)
	if err != nil {
		_ = p.closeLocked()
		return fmt.Errorf("RabbitMQ did not confirm click event: %w", err)
	}
	if !acknowledged {
		_ = p.closeLocked()
		return errors.New("RabbitMQ rejected click event")
	}
	select {
	case returned := <-p.returns:
		return fmt.Errorf("RabbitMQ returned click event: %s", returned.ReplyText)
	default:
		return nil
	}
}

func (p *rabbitPublisher) closeLocked() error {
	var err error
	if p.channel != nil {
		err = p.channel.Close()
		p.channel = nil
	}
	if p.conn != nil {
		if closeErr := p.conn.Close(); err == nil {
			err = closeErr
		}
		p.conn = nil
	}
	return err
}

func (p *rabbitPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.closeLocked()
}
