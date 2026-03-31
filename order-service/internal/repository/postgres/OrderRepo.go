package postgres

import (
	"database/sql"
	"fmt"
	"order-service/internal/domain"
)

type OrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (repo *OrderRepo) Create(o *domain.Order) error {
	var idempotencyKey interface{}
	if o.IdempotencyKey != "" {
		idempotencyKey = o.IdempotencyKey
	}

	_, err := repo.db.Exec("INSERT INTO orders (id, customer_id, item_name, amount, status, idempotency_key, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		o.ID, o.CustomerID, o.ItemName, o.Amount, o.Status, idempotencyKey, o.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create Order: %w", err)
	}

	return nil
}

func (repo *OrderRepo) GetByID(id string) (*domain.Order, error) {
	row := repo.db.QueryRow("SELECT id, customer_id, item_name, amount, status, idempotency_key, created_at FROM orders WHERE id=$1", id)

	var o domain.Order
	var idemKey sql.NullString
	err := row.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &idemKey, &o.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get Order: %w", err)
	}
	o.IdempotencyKey = idemKey.String

	return &o, nil
}

func (repo *OrderRepo) UpdateStatus(id string, status string) error {
	_, err := repo.db.Exec("UPDATE orders SET status=$1 WHERE id=$2", status, id)
	if err != nil {
		return fmt.Errorf("failed to update Order status: %w", err)
	}

	return nil
}

func (repo *OrderRepo) GetByIdempotencyKey(key string) (*domain.Order, error) {
	row := repo.db.QueryRow("SELECT id, customer_id, item_name, amount, status, idempotency_key, created_at FROM orders WHERE idempotency_key=$1", key)

	var o domain.Order
	var idemKey sql.NullString
	err := row.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &idemKey, &o.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get Order by idempotency key: %w", err)
	}
	o.IdempotencyKey = idemKey.String

	return &o, nil
}
