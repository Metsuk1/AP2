package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"

	"notification-service/internal/idempotency"
	"notification-service/internal/messaging"
)

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func main() {
	natsURL := getEnv("NATS_URL", nats.DefaultURL)

	store := idempotency.NewStore()
	consumer, err := messaging.NewConsumer(natsURL, store)
	if err != nil {
		log.Fatal("failed to start nats consumer:", err)
	}
	defer consumer.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- consumer.Run(ctx)
	}()

	select {
	case <-ctx.Done():
		log.Println("Shutting down Notification Service...")
	case err := <-errCh:
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Notification Service stopped")
}
