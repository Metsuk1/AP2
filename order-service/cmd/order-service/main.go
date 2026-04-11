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
	"github.com/joho/godotenv"
	googleGrpc "google.golang.org/grpc"

	"order-service/internal/database"
	repo "order-service/internal/repository/postgres"
	handler "order-service/internal/transport/http"
	myGrpc "order-service/internal/transport/http/grpc"
	"order-service/internal/usecase"
)

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return fallback
}

func main() {

	_ = godotenv.Load()

	dbHost := getEnv("DB_HOST", "localhost")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "0000")
	dbName := getEnv("DB_NAME", "ordersAP2_db")
	paymentGrpcUrl := getEnv("PAYMENT_GRPC_URL", "localhost:50051")
	grpcPort := getEnv("GRPC_PORT", "50052")
	restPort := getEnv("REST_PORT", "8080")

	db, err := database.Connect(dbHost, 5432, dbUser, dbPass, dbName)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()

	//create broker
	broker := usecase.NewOrderUpdatesBroker()

	//dependency
	orderRepo := repo.NewOrderRepo(db)

	paymentClient, err := myGrpc.NewPaymentClient(paymentGrpcUrl)
	if err != nil {
		log.Fatalf("failed to create payment client: %v", err)
	}
	defer paymentClient.Close()

	orderUC := usecase.NewOrderUseCase(orderRepo, paymentClient, broker)
	orderHandler := handler.NewOrderHandler(orderUC)

	router := gin.Default()
	handler.RegisterRoutes(router, orderHandler)

	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		s := googleGrpc.NewServer()
		pb.RegisterOrderServiceServer(s, myGrpc.NewOrderGRPCServer(orderUC))

		log.Printf("Order gRPC Server listening on :%s\n", grpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	srv := &http.Server{
		Addr:    ":" + restPort,
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
