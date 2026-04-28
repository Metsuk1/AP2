package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"notification-service/internal/domain"
	"notification-service/internal/idempotency"
)

const (
	paymentsStream = "PAYMENTS"
	paymentSubject = "payment.completed"
	durableName    = "notification-service"
	reconnectWait  = 2 * time.Second
	maxReconnects  = 10
)

type Consumer struct {
	conn  *nats.Conn
	js    nats.JetStreamContext
	store *idempotency.Store
}

func NewConsumer(url string, store *idempotency.Store) (*Consumer, error) {
	conn, err := nats.Connect(
		url,
		nats.Name("notification-service"),
		nats.ReconnectWait(reconnectWait),
		nats.MaxReconnects(maxReconnects),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to nats: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("create jetstream context: %w", err)
	}

	if _, err := js.StreamInfo(paymentsStream); err != nil {
		if err != nats.ErrStreamNotFound {
			conn.Close()
			return nil, fmt.Errorf("check payments stream: %w", err)
		}
		if _, err := js.AddStream(&nats.StreamConfig{
			Name:     paymentsStream,
			Subjects: []string{paymentSubject},
			Storage:  nats.FileStorage,
		}); err != nil {
			conn.Close()
			return nil, fmt.Errorf("create payments stream: %w", err)
		}
	}

	return &Consumer{conn: conn, js: js, store: store}, nil
}

func (c *Consumer) Run(ctx context.Context) error {
	sub, err := c.js.PullSubscribe(
		paymentSubject,
		durableName,
		nats.BindStream(paymentsStream),
		nats.ManualAck(),
		nats.AckExplicit(),
	)
	if err != nil {
		return fmt.Errorf("subscribe to payment events: %w", err)
	}

	log.Println("Notification Service is listening for payment.completed events")

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		messages, err := sub.Fetch(1, nats.MaxWait(time.Second))
		if err != nil {
			if err == nats.ErrTimeout {
				continue
			}
			return fmt.Errorf("fetch payment event: %w", err)
		}

		for _, msg := range messages {
			c.handleMessage(msg)
		}
	}
}

func (c *Consumer) handleMessage(msg *nats.Msg) {
	var event domain.PaymentCompletedEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("invalid payment event payload: %v", err)
		_ = msg.Nak()
		return
	}

	if event.EventID == "" {
		log.Println("invalid payment event: missing event_id")
		_ = msg.Nak()
		return
	}

	if c.store.IsProcessed(event.EventID) {
		_ = msg.Ack()
		return
	}

	log.Printf("[Notification] Sent email to %s for Order #%s. Amount: $%.2f", event.CustomerEmail, event.OrderID, float64(event.Amount)/100)
	c.store.MarkProcessed(event.EventID)

	if err := msg.Ack(); err != nil {
		log.Printf("failed to ack payment event %s: %v", event.EventID, err)
	}
}

func (c *Consumer) Close() {
	if c == nil || c.conn == nil {
		return
	}
	c.conn.Drain()
	c.conn.Close()
}
