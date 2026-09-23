package main

import (
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/werastine/CryptoNotifier/internal/core"
	"github.com/werastine/CryptoNotifier/pkg/pb"
	"google.golang.org/grpc"
)

func main() {
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

	grpcServer := grpc.NewServer()

	pb.RegisterCryptoServer(grpcServer, srv)

	if err = grpcServer.Serve(lis); err != nil {
		log.Printf("[ERROR] serving grpc server: %v", err)
		return
	}
}
