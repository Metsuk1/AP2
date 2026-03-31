package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"payment-service/internal/database"
	repo "payment-service/internal/repository/postgres"
	handler "payment-service/internal/transport/http"
	"payment-service/internal/usecase"
)

func main() {
	db, err := database.Connect("localhost", 5434, "postgres", "0000", "paymentsAP2_db")
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()

	// dependency
	paymentRepo := repo.NewPaymentRepo(db)
	paymentUC := usecase.NewPaymentUseCase(paymentRepo)
	paymentHandler := handler.NewPaymentHandler(paymentUC)

	router := gin.Default()
	handler.RegisterRoutes(router, paymentHandler)

	srv := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}

	go func() {
		log.Println("Payment Service starting on :8081")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start server:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Payment Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown:", err)
	}

	log.Println("Payment Service stopped")
}
