package payment_grpc

import (
	"context"
	"payment-service/internal/usecase"

	"github.com/Metsuk1/AP2_Generated/payment"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentGRPCHandler struct {
	payment.UnimplementedPaymentServiceServer
	useCase *usecase.PaymentUseCase
}

func NewPaymentGRPCHandler(uc *usecase.PaymentUseCase) *PaymentGRPCHandler {
	return &PaymentGRPCHandler{useCase: uc}
}

func (h *PaymentGRPCHandler) ProcessPayment(ctx context.Context, req *payment.PaymentRequest) (*payment.PaymentResponse, error) {
	customerEmail := req.GetCustomerEmail()
	if customerEmail == "" {
		return nil, status.Error(codes.InvalidArgument, "customer email is required")
	}

	p, err := h.useCase.CreatePayment(req.GetOrderId(), req.GetAmount(), req.GetIdempotencyKey(), customerEmail)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to process payment: %v", err)
	}

	return &payment.PaymentResponse{
		Id:            p.ID,
		TransactionId: p.TransactionID,
		Status:        p.Status,
	}, nil
}
