package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"order-service/internal/domain"
)

type OrderUseCase struct {
	repo    domain.OrderRepository
	payment domain.PaymentClient
	Broker  *OrderUpdatesBroker
}

func NewOrderUseCase(repo domain.OrderRepository, payment domain.PaymentClient, broker *OrderUpdatesBroker) *OrderUseCase {
	return &OrderUseCase{
		repo:    repo,
		payment: payment,
		Broker:  broker,
	}
}

func (uc *OrderUseCase) CreateOrder(customerID, itemName string, amount int64, idempotencyKey string) (*domain.Order, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	if idempotencyKey != "" {
		existing, err := uc.repo.GetByIdempotencyKey(idempotencyKey)
		if err == nil {
			return existing, nil
		}

	}

	order := &domain.Order{
		ID:             uuid.New().String(),
		CustomerID:     customerID,
		ItemName:       itemName,
		Amount:         amount,
		Status:         "Pending",
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now(),
	}

	if err := uc.repo.Create(order); err != nil {
		return nil, err
	}

	paymentStatus, err := uc.payment.RequestPayment(order.ID, order.Amount, idempotencyKey)
	if err != nil {
		order.Status = "Failed"
		uc.UpdateStatus(order.ID, order.Status)
		return order, nil
	}

	if paymentStatus == "Authorized" {
		order.Status = "Paid"
	} else {
		order.Status = "Failed"
	}
	uc.UpdateStatus(order.ID, order.Status)

	return order, nil
}

func (uc *OrderUseCase) GetOrder(id string) (*domain.Order, error) {
	return uc.repo.GetByID(id)
}

func (uc *OrderUseCase) CancelOrder(id string) (*domain.Order, error) {
	order, err := uc.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("order not found")
	}

	if order.Status == "Paid" {
		return nil, errors.New("cannot cancel a paid order")
	}

	if order.Status != "Pending" {
		return nil, errors.New("only pending orders can be cancelled")
	}

	order.Status = "Cancelled"
	if err := uc.UpdateStatus(order.ID, order.Status); err != nil {
		return nil, err
	}

	return order, nil
}

func (uc *OrderUseCase) UpdateStatus(orderID, status string) error {
	if err := uc.repo.UpdateStatus(orderID, status); err != nil {
		return err
	}
	if uc.Broker != nil {
		uc.Broker.Notify(orderID, status)
	}

	return nil
}
