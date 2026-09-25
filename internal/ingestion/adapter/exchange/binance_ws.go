package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/werastine/CryptoNotifier/internal/ingestion/service"
	"github.com/werastine/CryptoNotifier/pkg/pb"
)

type DataTick struct {
	Stream string `json:"stream"`
	Data   struct {
		Symbol string `json:"s"`
		Price  string `json:"c"`
		Time   int64  `json:"C"`
	} `json:"data"`
}

func Connect(ctx context.Context, rdb *redis.Client, pbl service.Publisher) {
	url := "wss://stream.binance.com:9443/stream?streams=btcusdt@ticker/ethusdt@ticker/solusdt@ticker"

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "stopping Binance WebSocket worker")
			return
		default:

			conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
			if err != nil {
				slog.ErrorContext(ctx, "failed to connect to binance websocket, reconnecting in 5 seconds", "err", err)
				select {
				case <-ctx.Done():
					slog.InfoContext(ctx, "stopping reconnect during backoff")
					return
				case <-time.After(5 * time.Second):
				}

				continue
			}

			slog.InfoContext(ctx, "created websocket connection")

			readLoop(ctx, rdb, pbl, conn)
			slog.WarnContext(ctx, "Websocket connection lost, reconnecting it...")

			select {
			case <-ctx.Done():
				slog.InfoContext(ctx, "stopping reconnect during backoff")
				return
			case <-time.After(2 * time.Second):
			}
		}
	}
}

func readLoop(ctx context.Context, rdb *redis.Client, pbl service.Publisher, conn *websocket.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			slog.ErrorContext(ctx, "failed to close websocket connection", "err", err)
		}
	}()

	stop := context.AfterFunc(ctx, func() {
		_ = conn.Close()
	})
	defer stop()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, p, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					slog.InfoContext(ctx, "websocket connection is closed", "err", err)
					return
				}
				slog.ErrorContext(ctx, "failed to read bytes from connection", "err", err)
				return
			}

			var dt DataTick
			if err := json.Unmarshal(p, &dt); err != nil {
				slog.WarnContext(ctx, "failed too unmarshal json",
					"err", err,
					"raw_payload", string(p),
				)
				continue
			}

			if err := redisSet(ctx, rdb, dt.Data.Symbol); err != nil {
				slog.Error("failed to set symbol in redis DB", "err", err)
			}

			if dt.Data.Symbol == "" {
				continue
			}

			if err := pbl.Publish(ctx, &pb.PriceTick{
				Symbol: dt.Data.Symbol,
				Price:  dt.Data.Price,
			}); err != nil {
				slog.ErrorContext(ctx, "failed to publish price tick",
					"symbol", dt.Data.Symbol,
					"price", dt.Data.Price,
					"err", err,
				)
				return
			}
		}
	}
}

func redisSet(ctx context.Context, rdb *redis.Client, symbol string) error {
	if err := rdb.SAdd(ctx, "Active_Symbol", symbol).Err(); err != nil {
		return fmt.Errorf("failed to add new symbol: %s in redis DB: %w", symbol, err)
	}
	return nil
}
