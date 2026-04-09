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

	"order-service/internal/database"
	repo "order-service/internal/repository/postgres"
	handler "order-service/internal/transport/http"
	grpc "order-service/internal/transport/http/grpc"
	"order-service/internal/usecase"
)

func main() {
	db, err := database.Connect("localhost", 5433, "postgres", "0000", "ordersAP2_db")
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()

	//dependency
	orderRepo := repo.NewOrderRepo(db)
	paymentClient, err := grpc.NewPaymentClient("localhost:50051")
	if err != nil {
		log.Fatalf("failed to create payment client: %v", err)
	}
	defer paymentClient.Close()

	orderUC := usecase.NewOrderUseCase(orderRepo, paymentClient)
	orderHandler := handler.NewOrderHandler(orderUC)

	router := gin.Default()
	handler.RegisterRoutes(router, orderHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.Println("Order Service starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start server:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Order Service... ")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown:", err)
	}

	log.Println("Order Service stopped")
}
