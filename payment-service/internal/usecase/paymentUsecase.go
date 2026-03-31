package usecase

import (
	"errors"

	"github.com/google/uuid"

	"payment-service/internal/domain"
)

type PaymentUseCase struct {
	repo domain.PaymentRepository
}

func NewPaymentUseCase(repo domain.PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: repo}
}

func (uc *PaymentUseCase) CreatePayment(orderID string, amount int64, idempotencyKey string) (*domain.Payment, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	if idempotencyKey != "" {
		existing, err := uc.repo.GetByIdempotencyKey(idempotencyKey)
		if err == nil {
			return existing, nil
		}
	}

	status := "Authorized"
	if amount > 100000 {
		status = "Declined"
	}

	payment := &domain.Payment{
		ID:             uuid.New().String(),
		OrderID:        orderID,
		TransactionID:  uuid.New().String(),
		Amount:         amount,
		Status:         status,
		IdempotencyKey: idempotencyKey,
	}

	if err := uc.repo.Create(payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (uc *PaymentUseCase) GetPaymentByOrderID(orderID string) (*domain.Payment, error) {
	return uc.repo.GetPayment(orderID)
}
