package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"payment-service/internal/usecase"
)

type PaymentHandler struct {
	uc *usecase.PaymentUseCase
}

func NewPaymentHandler(uc *usecase.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

type createPaymentRequest struct {
	OrderID       string `json:"order_id" binding:"required"`
	Amount        int64  `json:"amount" binding:"required"`
	CustomerEmail string `json:"customer_email" binding:"required,email"`
}

// POST /payments
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req createPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	idempotencyKey := c.GetHeader("Idempotency-Key")
	payment, err := h.uc.CreatePayment(req.OrderID, req.Amount, idempotencyKey, req.CustomerEmail)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, payment)
}

// GET /payments/:order_id
func (h *PaymentHandler) GetPaymentByOrderID(c *gin.Context) {
	orderID := c.Param("order_id")

	payment, err := h.uc.GetPaymentByOrderID(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})

		return
	}

	c.JSON(http.StatusOK, payment)
}

func RegisterRoutes(router *gin.Engine, handler *PaymentHandler) {
	router.POST("/payments", handler.CreatePayment)
	router.GET("/payments/:order_id", handler.GetPaymentByOrderID)
}
