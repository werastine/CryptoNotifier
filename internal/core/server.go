package core

import (
	"fmt"
	"io"
	"log"

	"github.com/google/uuid"
	"github.com/werastine/CryptoNotifier/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedCryptoServer
	service *CryptoService
}

func NewServer(service *CryptoService) *Server {
	return &Server{
		service: service,
	}
}

func (s *Server) PushRates(stream grpc.ClientStreamingServer[pb.PriceTick, pb.PushSummary]) error {
	for {
		pt, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.PushSummary{})
		}
		if err != nil {
			log.Printf("[ERROR] reciving price tick from exchange")
			return err
		}

		targets := s.service.FindTarget(pt.GetSymbol(), pt.GetPrice())
		if len(targets) == 0 {
			continue
		}

		for _, k := range targets {
			select {
			case k <- pt:

			default:
				log.Println("[WARN] client channel full, dropping tick")
			}
		}
	}
}

func (s *Server) WaitNotification(nq *pb.NotifyRequest, stream grpc.ServerStreamingServer[pb.PriceTick]) error {
	log.Printf("Recieved notify target. Symbol: %s, Price: %s", nq.GetSymbol(), nq.GetTargetPrice())

	clientChan := make(chan *pb.PriceTick, 1)
	uid := uuid.NewString()

	if nq.GetCondition() != pb.Condition_CONDITION_UNSPECIFIED {
		return status.Error(codes.InvalidArgument, "condition is requied")
	}

	if err := s.service.AddTarget(
		nq.GetTargetPrice(),
		nq.GetSymbol(),
		uid,
		nq.GetCondition(),
		clientChan,
	); err != nil {
		return fmt.Errorf("adding target: %v", err)
	}

	defer func() {
		s.service.RemoveTarget(uid)
		close(clientChan)
	}()

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()

		case tick := <-clientChan:
			if err := stream.Send(tick); err != nil {
				return fmt.Errorf("sending tick by gRPC stream: %v", err)
			}
			log.Printf("[INFO] %s reached the target price %s", nq.GetSymbol(), nq.GetTargetPrice())
			return nil
		}

	}
}
