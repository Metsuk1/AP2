package usecase

import (
	"sync"

	pb "github.com/Metsuk1/AP2_Generated/order"
)

type OrderUpdatesBroker struct {
	mu          sync.RWMutex
	subscribers map[string][]chan *pb.OrderStatusUpdate
}

func NewOrderUpdatesBroker() *OrderUpdatesBroker {
	return &OrderUpdatesBroker{
		subscribers: make(map[string][]chan *pb.OrderStatusUpdate),
	}
}

func (b *OrderUpdatesBroker) Subscribe(orderID string) chan *pb.OrderStatusUpdate {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan *pb.OrderStatusUpdate, 1)
	b.subscribers[orderID] = append(b.subscribers[orderID], ch)
	return ch
}

func (b *OrderUpdatesBroker) Notify(orderID, status string) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if subs, ok := b.subscribers[orderID]; ok {
		for _, ch := range subs {
			ch <- &pb.OrderStatusUpdate{
				OrderId: orderID,
				Status:  status,
			}
		}
	}
}
