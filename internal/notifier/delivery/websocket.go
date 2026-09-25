package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/werastine/CryptoNotifier/internal/notifier/service"
	"github.com/werastine/CryptoNotifier/pkg/pb"
)

type TargetData struct {
	Symbol    string `json:"symbol"`
	Price     string `json:"price"`
	Condition string `json:"condition"` // above / below
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func Subscribe(ctx context.Context, client service.Sender, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mtx := sync.Mutex{}
		reqCtx := r.Context()

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("[ERROR] upgrading connection: %v", err)
			return
		}

		defer func() {
			if err = conn.Close(); err != nil {
				log.Printf("[ERROR] closing connection: %v", err)
			}
		}()

		log.Println("Successfuly upgraded to a websocket!")

		writeSafe := func(messageType int, data []byte) {
			mtx.Lock()
			defer mtx.Unlock()

			_ = conn.WriteMessage(messageType, data)
		}

		for {
			_, p, err := conn.ReadMessage()
			if err != nil {
				log.Printf("[ERROR] reading bytes from connection: %v", err)
				break
			}

			var tg TargetData

			if err = json.Unmarshal(p, &tg); err != nil {
				log.Printf("[ERROR] unmarshaling json: %v", err)
				writeSafe(websocket.TextMessage, []byte("incorrect data type"))
				continue
			}

			isMember, err := rdb.SIsMember(ctx, "Active_Symbol", tg.Symbol).Result()
			if err != nil {
				slog.Error("failed to check symbol in redis", "err", err)
			}
			if !isMember {
				writeSafe(websocket.TextMessage, []byte("Incorrect symbol\ne.g. BTCUSDT, ETHUSDT"))
			}

			notifyTransfer, err := client.Send(reqCtx, tg.Symbol, tg.Price, tg.Condition)
			if err != nil {
				log.Printf("[ERROR] sending requsest to core layer: %v", err)
				writeSafe(websocket.TextMessage, []byte(fmt.Sprintf("recieved an error: %v", err)))
				continue
			}

			go func(ch <-chan *pb.PriceTick) {
				select {
				case nf, ok := <-ch:
					if !ok {
						return
					}
					msg := fmt.Sprintf("Got a notification! \nSymbol: %s\nPrice: %s", nf.Symbol, nf.Price)
					writeSafe(websocket.TextMessage, []byte(msg))

				case <-reqCtx.Done():
					log.Println("[INFO] client disconnected, exiting alert goroutine")
					return
				}

			}(notifyTransfer)
		}
	}
}
