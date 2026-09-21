package service

import (
	"context"

	"github.com/werastine/CryptoNotifier/pkg/pb"
)

type Publisher interface {
	Publish(ctx context.Context, tick *pb.PriceTick) error
}
