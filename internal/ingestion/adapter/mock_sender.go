package adapter

import (
	"context"
	"fmt"

	"github.com/werastine/CryptoNotifier/pkg/pb"
	"google.golang.org/grpc"
)

type MockTickSender struct{}

func NewMockTickSender(conn *grpc.ClientConn) *MockTickSender {
	return &MockTickSender{}
}

func (ts *MockTickSender) Publish(ctx context.Context, tick *pb.PriceTick) error {

	fmt.Printf("Recieved new Ticket: Symbol: %s, Price: %s\n", tick.Symbol, tick.Price)
	return nil
}
