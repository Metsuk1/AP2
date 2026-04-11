package grpc

import (
	"context"
	"fmt"
	"time"

	pb "github.com/Metsuk1/AP2_Generated/payment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PaymentClient struct {
	client pb.PaymentServiceClient
	conn   *grpc.ClientConn
}

func NewPaymentClient(addr string) (*PaymentClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("could not connect to payment service: %v", err)
	}

	return &PaymentClient{
		client: pb.NewPaymentServiceClient(conn),
		conn:   conn,
	}, nil
}

func (pc *PaymentClient) Close() {
	pc.conn.Close()
}

func (pc *PaymentClient) RequestPayment(orderID string, amount int64, idempotencyKey string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := pc.client.ProcessPayment(ctx, &pb.PaymentRequest{
		OrderId:        orderID,
		Amount:         amount,
		IdempotencyKey: idempotencyKey,
	})

	if err != nil {
		return "", fmt.Errorf("gRPC call failed: %w", err)
	}

	return resp.Status, nil
}
