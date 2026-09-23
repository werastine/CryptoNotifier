package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/werastine/CryptoNotifier/internal/notifier/adapter"
	"github.com/werastine/CryptoNotifier/internal/notifier/delivery"
	"github.com/werastine/CryptoNotifier/internal/notifier/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	coreAddr := os.Getenv("CORE_GRPC_ADDR")
	if coreAddr == "" {
		slog.Error("core addres environment variable is not provided")
	}

	conn, err := grpc.NewClient(
		coreAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("[ERROR] did not connect: %v", err)
	}

	defer func() {
		if err = conn.Close(); err != nil {
			log.Panicf("[ERROR] did not closed: %v", err)
		}
	}()

	var sender service.Sender = adapter.NewGRPCNotifySender(conn) // real sender
	// var sender service.Sender = adapter.NewMockSender(conn) // mock sender

	mux := &http.ServeMux{}
	mux.HandleFunc("/subscribe", delivery.Subscribe(sender))

	log.Println("[INFO] Listening port 8081")

	if err = http.ListenAndServe(":8081", mux); err != nil {
		log.Printf("[ERROR] listening interrupted: %v", err)
		return
	}

}
