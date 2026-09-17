package service

import (
	"context"

	"github.com/werastine/CryptoNotifier/pkg/pb"
)

type Sender interface {
	Send(ctx context.Context, symbol string, price string, condition string) (<-chan *pb.PriceTick, error)
}
