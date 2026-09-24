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
	"google.golang.org/grpc/credentials"
)

func main() {
	certFile := os.Getenv("CERT_FILE")
	keyFile := os.Getenv("KEY_FILE")
	if certFile == "" || keyFile == "" {
		slog.Error("TLS cerificates environment variables are not provided")
		return
	}

	coreAddr := os.Getenv("CORE_GRPC_ADDR")
	if coreAddr == "" {
		slog.Error("core addres environment variable is not provided")
		return
	}

	creds, err := credentials.NewClientTLSFromFile(certFile, "")
	if err != nil {
		slog.Error("failed to set tls certification")
		return
	}

	conn, err := grpc.NewClient(
		coreAddr,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		log.Fatalf("[ERROR] did not connect: %v", err)
	}

	defer func() {
		if err = conn.Close(); err != nil {
			log.Printf("[ERROR] did not closed: %v", err)
		}
	}()

	var sender service.Sender = adapter.NewGRPCNotifySender(conn) // real sender
	// var sender service.Sender = adapter.NewMockSender(conn) // mock sender

	mux := &http.ServeMux{}
	mux.HandleFunc("/subscribe", delivery.Subscribe(sender))

	log.Println("[INFO] Listening port 8081")

	if err = http.ListenAndServeTLS(":8081", certFile, keyFile, mux); err != nil {
		log.Printf("[ERROR] listening interrupted: %v", err)
		return
	}

}
