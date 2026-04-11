package grpc

import (
	"order-service/internal/usecase"

	pb "github.com/Metsuk1/AP2_Generated/order"
)

type OrderGRPCServer struct {
	pb.UnimplementedOrderServiceServer
	useCase *usecase.OrderUseCase
}

func NewOrderGRPCServer(uc *usecase.OrderUseCase) *OrderGRPCServer {
	return &OrderGRPCServer{useCase: uc}
}

func (s *OrderGRPCServer) SubscribeToOrderUpdates(req *pb.OrderRequest, stream pb.OrderService_SubscribeToOrderUpdatesServer) error {
	updateChan := s.useCase.Broker.Subscribe(req.OrderId)

	for {
		select {
		case update := <-updateChan:
			if err := stream.Send(update); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}
