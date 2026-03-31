package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HTTP request to the Payment Service
type PaymentClient struct {
	baseURL string
	client  *http.Client
}

// Client with  timeout 2sec
func NewPaymentClient(baseURL string) *PaymentClient {
	return &PaymentClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

type paymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

type paymentResponse struct {
	ID            string `json:"id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
}

func (pc *PaymentClient) RequestPayment(orderID string, amount int64) (string, error) {
	reqBody, err := json.Marshal(paymentRequest{
		OrderID: orderID,
		Amount:  amount,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := pc.client.Post(
		pc.baseURL+"/payments",
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		// Timeout сработает здесь — Payment Service недоступен
		return "", fmt.Errorf("payment service unavailable: %w", err)
	}
	defer resp.Body.Close()

	var paymentResp paymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return paymentResp.Status, nil
}
