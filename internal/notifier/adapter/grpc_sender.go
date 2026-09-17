package adapter

import (
	"context"
	"fmt"
	"log"

	"github.com/werastine/CryptoNotifier/pkg/pb"
	"google.golang.org/grpc"
)

type GRPCNotifySender struct {
	client pb.CryptoClient
}

func NewGRPCNotifySender(conn *grpc.ClientConn) *GRPCNotifySender {
	return &GRPCNotifySender{
		client: pb.NewCryptoClient(conn),
	}
}

func (ns *GRPCNotifySender) Send(ctx context.Context, symbol string, price string, condition string) (<-chan *pb.PriceTick, error) {
	notifyTransfer := make(chan *pb.PriceTick, 1)

	var protoCond pb.Condition
	switch condition {
	case "ABOVE", "above", ">", ">=":
		protoCond = pb.Condition_CONDITION_ABOVE

	case "BELOW", "below", "<", "<=":
		protoCond = pb.Condition_CONDITION_BELOW

	default:
		return nil, fmt.Errorf("invalid condition: %s", condition)
	}

	req := &pb.NotifyRequest{
		Symbol:      symbol,
		TargetPrice: price,
		Condition:   protoCond,
	}

	stream, err := ns.client.WaitNotification(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("creating stream: %v", err)
	}

	go func() {
		defer close(notifyTransfer)

		tick, err := stream.Recv()
		if err != nil {
			log.Printf("[INFO] gRPC stream closed: %v", err)
			return
		}

		notifyTransfer <- tick
	}()

	return notifyTransfer, nil
}
