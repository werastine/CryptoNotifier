package main

import (
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/werastine/CryptoNotifier/internal/core"
	"github.com/werastine/CryptoNotifier/pkg/pb"
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

	cs := core.NewCryptoService()
	srv := core.NewServer(cs)
	port := os.Getenv("PORT")

	if port == "" {
		slog.Error("port enviroment variable is not provided")
		return
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Printf("[ERROR] listening tcp: %v", err)
		return
	}

	defer func() {
		if err = lis.Close(); err != nil {
			log.Printf("[ERROR] closing connection: %v", err)
		}
	}()

	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		slog.Error("failed to set tls certification")
		return
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
	)

	pb.RegisterCryptoServer(grpcServer, srv)

	if err = grpcServer.Serve(lis); err != nil {
		log.Printf("[ERROR] serving grpc server: %v", err)
		return
	}
}
