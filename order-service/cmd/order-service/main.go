package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/Metsuk1/AP2_Generated/order"
	"github.com/gin-gonic/gin"

	googleGrpc "google.golang.org/grpc"

	"order-service/internal/database"
	repo "order-service/internal/repository/postgres"
	handler "order-service/internal/transport/http"
	myGrpc "order-service/internal/transport/http/grpc"
	"order-service/internal/usecase"
)

func main() {
	db, err := database.Connect("localhost", 5432, "postgres", "0000", "ordersAP2_db")
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()

	//create broker
	broker := usecase.NewOrderUpdatesBroker()

	//dependency
	orderRepo := repo.NewOrderRepo(db)

	paymentClient, err := myGrpc.NewPaymentClient("localhost:50051")
	if err != nil {
		log.Fatalf("failed to create payment client: %v", err)
	}
	defer paymentClient.Close()

	orderUC := usecase.NewOrderUseCase(orderRepo, paymentClient, broker)
	orderHandler := handler.NewOrderHandler(orderUC)

	router := gin.Default()
	handler.RegisterRoutes(router, orderHandler)

	go func() {
		lis, err := net.Listen("tcp", ":50052")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		s := googleGrpc.NewServer()
		pb.RegisterOrderServiceServer(s, myGrpc.NewOrderGRPCServer(orderUC))

		log.Println("Order gRPC Server listening on :50052")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

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
