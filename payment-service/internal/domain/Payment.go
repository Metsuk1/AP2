package domain

type Payment struct {
	ID              string
	OrderID         string
	TransactionID   string
	Amount          int64
	Status          string // "Authorized", "Declined"
	IdempotencyKey  string
}

type PaymentRepository interface {
	Create(*Payment) error
	GetPayment(orderId string) (*Payment, error)
	GetByIdempotencyKey(key string) (*Payment, error)
}
