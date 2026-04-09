package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	payment_grpc "payment-service/internal/transport/http/grpc"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"payment-service/internal/database"
	repo "payment-service/internal/repository/postgres"
	"payment-service/internal/usecase"

	"github.com/Metsuk1/AP2_Generated/payment"
)

func main() {
	db, err := database.Connect("localhost", 5434, "postgres", "0000", "paymentsAP2_db")
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()

	paymentRepo := repo.NewPaymentRepo(db)
	paymentUC := usecase.NewPaymentUseCase(paymentRepo)

	paymentGRPCHandler := payment_grpc.NewPaymentGRPCHandler(paymentUC)

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	payment.RegisterPaymentServiceServer(srv, paymentGRPCHandler)

	reflection.Register(srv)

	go func() {
		log.Printf("Payment gRPC Service starting on :%s", port)
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
