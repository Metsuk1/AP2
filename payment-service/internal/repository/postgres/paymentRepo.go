package postgres

import (
	"database/sql"
	"fmt"
	"payment-service/internal/domain"
)

type PaymentRepo struct {
	db *sql.DB
}

func NewPaymentRepo(db *sql.DB) *PaymentRepo {
	return &PaymentRepo{db: db}
}

func (repo *PaymentRepo) Create(p *domain.Payment) error {
	var idempotencyKey interface{}
	if p.IdempotencyKey != "" {
		idempotencyKey = p.IdempotencyKey
	}

	_, err := repo.db.Exec("INSERT INTO payments (id, order_id, transaction_id, amount, status, idempotency_key) VALUES ($1, $2, $3, $4, $5, $6)",
		p.ID, p.OrderID, p.TransactionID, p.Amount, p.Status, idempotencyKey)
	if err != nil {
		return fmt.Errorf("failed to create Payment: %w", err)
	}

	return nil
}

func (repo *PaymentRepo) GetPayment(orderID string) (*domain.Payment, error) {
	row := repo.db.QueryRow("SELECT id, order_id, transaction_id, amount, status, idempotency_key FROM payments WHERE order_id=$1", orderID)

	var p domain.Payment
	var idemKey sql.NullString
	err := row.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status, &idemKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get Payment: %w", err)
	}
	p.IdempotencyKey = idemKey.String

	return &p, nil
}

func (repo *PaymentRepo) GetByIdempotencyKey(key string) (*domain.Payment, error) {
	row := repo.db.QueryRow("SELECT id, order_id, transaction_id, amount, status, idempotency_key FROM payments WHERE idempotency_key=$1", key)

	var p domain.Payment
	var idemKey sql.NullString
	err := row.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status, &idemKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get Payment by idempotency key: %w", err)
	}
	p.IdempotencyKey = idemKey.String

	return &p, nil
}
