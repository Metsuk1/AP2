package domain

import "time"

type Order struct {
	ID             string
	CustomerID     string
	ItemName       string
	Amount         int64  // Amount in cents
	Status         string // "Pending", "Paid", "Failed", "Cancelled"
	IdempotencyKey string
	CreatedAt      time.Time
}

type OrderRepository interface {
	Create(order *Order) error
	GetByID(id string) (*Order, error)
	UpdateStatus(id string, status string) error
	GetByIdempotencyKey(key string) (*Order, error)
}

type PaymentClient interface {
	RequestPayment(orderID string, amount int64, idempotencyKey string) (string, error)
}
