package exchange

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
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

func Connect(ctx context.Context, pbl service.Publisher) {
	url := "wss://stream.binance.com:9443/stream?streams=btcusdt@ticker/ethusdt@ticker/solusdt@ticker"

	for {
		select {
		case <-ctx.Done():
			log.Println("[INFO] Stopping Binance WebSocket worker")
			return
		default:

			conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
			if err != nil {
				log.Printf("[ERROR] connecting to binance wss: %v, reconnecting in 5 seconds", err)
				select {
				case <-ctx.Done():
					log.Println("[INFO] Stopping reconnect during backoff")
					return
				case <-time.After(5 * time.Second):
				}

				continue
			}

			log.Println("[INFO] Created websocket connection")

			readLoop(ctx, pbl, conn)
			log.Println("[WARN] Websocket connection lost, reconnecting it...")

			select {
			case <-ctx.Done():
				log.Println("[INFO] Stopping reconnect during backoff")
				return
			case <-time.After(2 * time.Second):
			}
		}
	}
}

func readLoop(ctx context.Context, pbl service.Publisher, conn *websocket.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("[ERROR] closing websocket connection: %v", err)
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
				log.Printf("[ERROR] reading bytes from connection: %v", err)
				return
			}
			var dt DataTick
			if err := json.Unmarshal(p, &dt); err != nil {
				log.Printf("[ERROR] unmarshaling json: %v", err)
				continue
			}

			if dt.Data.Symbol == "" {
				continue
			}

			if err := pbl.Publish(ctx, &pb.PriceTick{
				Symbol: dt.Data.Symbol,
				Price:  dt.Data.Price,
			}); err != nil {
				log.Printf("[ERROR] failed to publish DataTick: %v", err)
				return
			}
		}
	}
}
