package domain

type Payment struct {
	ID             string
	OrderID        string
	TransactionID string
	Amount        int64
	Status        string // "Authorized", "Declined"
	IdempotencyKey string
}

type PaymentCompletedEvent struct {
	EventID       string `json:"event_id"`
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
}

type PaymentRepository interface {
	Create(*Payment) error
	GetPayment(orderId string) (*Payment, error)
	GetByIdempotencyKey(key string) (*Payment, error)
}

type PaymentEventPublisher interface {
	PublishPaymentCompleted(event PaymentCompletedEvent) error
}
