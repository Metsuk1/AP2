package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	payment_grpc "payment-service/internal/transport/http/grpc"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"payment-service/internal/database"
	"payment-service/internal/messaging"
	repo "payment-service/internal/repository/postgres"
	"payment-service/internal/usecase"

	"github.com/Metsuk1/AP2_Generated/payment"
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
	dbName := getEnv("DB_NAME", "paymentsAP2_db")
	grpcPort := getEnv("GRPC_PORT", "50051")
	natsURL := getEnv("NATS_URL", nats.DefaultURL)

	db, err := database.Connect(dbHost, 5432, dbUser, dbPass, dbName)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()

	paymentRepo := repo.NewPaymentRepo(db)
	publisher, err := messaging.NewNATSPublisher(natsURL)
	if err != nil {
		log.Fatal("failed to connect to nats:", err)
	}
	defer publisher.Close()

	paymentUC := usecase.NewPaymentUseCase(paymentRepo, publisher)

	paymentGRPCHandler := payment_grpc.NewPaymentGRPCHandler(paymentUC)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	payment.RegisterPaymentServiceServer(srv, paymentGRPCHandler)

	reflection.Register(srv)

	go func() {
		log.Printf("Payment gRPC Service starting on :%s", grpcPort)
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutdown payment service..")
	srv.GracefulStop()
	log.Println("Payment-service is stopped")
}
