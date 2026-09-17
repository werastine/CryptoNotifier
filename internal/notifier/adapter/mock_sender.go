package adapter

import (
	"context"

	"github.com/werastine/CryptoNotifier/pkg/pb"
	"google.golang.org/grpc"
)

type MockSender struct{}

func NewMockSender(conn *grpc.ClientConn) *MockSender {
	return &MockSender{}
}

func (ns *MockSender) Send(ctx context.Context, symbol string, price string, condition string) (<-chan *pb.PriceTick, error) {
	notifyTransfer := make(chan *pb.PriceTick, 1)
	notifyTransfer <- &pb.PriceTick{
		Symbol: "BTC",
		Price:  "192021",
	}
	return notifyTransfer, nil
}
