package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/werastine/CryptoNotifier/internal/ingestion/adapter"
	"github.com/werastine/CryptoNotifier/internal/ingestion/adapter/exchange"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	wg := sync.WaitGroup{}
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ctx, cancel := context.WithCancel(signalCtx)
	defer cancel()

	conn, err := grpc.NewClient(
		"127.0.0.1:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("[ERROR] creating new client")
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("[ERROR] closing grpc connection: %v", err)
		}
	}()

	ts := adapter.NewGRPCTickSender(conn) // real sender
	// ts := adapter.NewMockTickSender(conn) // mock sender

	wg.Add(1)

	go func() {
		defer cancel()
		defer wg.Done()

		exchange.Connect(ctx, ts)
	}()

	<-ctx.Done()
	log.Println("[INFO] Context is done")
	wg.Wait()
	log.Println("[INFO] Ingestion server stopped gracefully")

}
