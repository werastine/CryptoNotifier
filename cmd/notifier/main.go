package main

import (
	"log"
	"net/http"

	"github.com/werastine/CryptoNotifier/internal/notifier/adapter"
	"github.com/werastine/CryptoNotifier/internal/notifier/delivery"
	"github.com/werastine/CryptoNotifier/internal/notifier/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient(
		"localhost:50051",
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

	// var sender service.Sender = adapter.NewGRPCNotifySender(conn)
	var sender service.Sender = adapter.NewMockSender(conn) // mock sender

	mux := &http.ServeMux{}
	mux.HandleFunc("/subscribe", delivery.Subscribe(sender))

	log.Println("[INFO] Listening port 8081")

	if err = http.ListenAndServe(":8081", mux); err != nil {
		log.Printf("[ERROR] listening interrupted: %v", err)
		return
	}

}
