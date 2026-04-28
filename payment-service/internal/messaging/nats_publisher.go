package messaging

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"

	"payment-service/internal/domain"
)

const (
	paymentsStream  = "PAYMENTS"
	paymentSubject  = "payment.completed"
	publishTimeout  = 5 * time.Second
	reconnectWait   = 2 * time.Second
	maxReconnects   = 10
)

type NATSPublisher struct {
	conn *nats.Conn
	js   nats.JetStreamContext
}

func NewNATSPublisher(url string) (*NATSPublisher, error) {
	conn, err := nats.Connect(
		url,
		nats.Name("payment-service"),
		nats.ReconnectWait(reconnectWait),
		nats.MaxReconnects(maxReconnects),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to nats: %w", err)
	}

	js, err := conn.JetStream(nats.PublishAsyncMaxPending(256))
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

	return &NATSPublisher{conn: conn, js: js}, nil
}

func (p *NATSPublisher) PublishPaymentCompleted(event domain.PaymentCompletedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal payment event: %w", err)
	}

	ack, err := p.js.Publish(paymentSubject, body, nats.MsgId(event.EventID), nats.AckWait(publishTimeout))
	if err != nil {
		return fmt.Errorf("publish payment event: %w", err)
	}
	if ack == nil || ack.Stream != paymentsStream {
		return fmt.Errorf("payment event was not acknowledged by stream")
	}

	return nil
}

func (p *NATSPublisher) Close() {
	if p == nil || p.conn == nil {
		return
	}
	p.conn.Drain()
	p.conn.Close()
}
