package adapter

import (
	"context"
	"fmt"
	"sync"

	"github.com/werastine/CryptoNotifier/pkg/pb"
	"google.golang.org/grpc"
)

type GRPCTickSender struct {
	client pb.CryptoClient
	stream pb.Crypto_PushRatesClient
	mu     sync.Mutex
}

func NewGRPCTickSender(conn *grpc.ClientConn) *GRPCTickSender {
	return &GRPCTickSender{
		client: pb.NewCryptoClient(conn),
		mu:     sync.Mutex{},
	}
}

func (ts *GRPCTickSender) Publish(ctx context.Context, tick *pb.PriceTick) error {

	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.stream == nil {
		stream, err := ts.client.PushRates(ctx)
		if err != nil {
			return fmt.Errorf("failed to return PushRates stream: %w", err)
		}
		ts.stream = stream
	}

	if err := ts.stream.Send(tick); err != nil {
		ts.stream = nil
		return fmt.Errorf("failed to send tick: %w", err)
	}

	return nil
}
