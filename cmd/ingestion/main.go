package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/werastine/CryptoNotifier/internal/ingestion/adapter"
	"github.com/werastine/CryptoNotifier/internal/ingestion/adapter/exchange"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	certFile := os.Getenv("CERT_FILE")
	if certFile == "" {
		slog.Error("TLS cerificate environment variable is not provided")
		return
	}

	wg := sync.WaitGroup{}
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	coreAddr := os.Getenv("CORE_GRPC_ADDR")
	if coreAddr == "" {
		slog.Error("core addres enviroment variable is not provided")
	}

	ctx, cancel := context.WithCancel(signalCtx)
	defer cancel()

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
		slog.Error("failed to create a client", "err", err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Error("failed to close grpc connection", "err", err)
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
	slog.Info("context is done")
	wg.Wait()
	slog.Info("ingestion server stopped gracefully")

}
