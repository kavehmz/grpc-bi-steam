package main

import (
	"log"
	"net"

	pb "githib.com/regentmarkets/segmented/hub"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedServiceHubServer
}

type chRelay struct {
	req   *pb.CallHTTPRequest
	resCh chan *pb.CallHTTPResponse
}

var ch = make(chan *chRelay)

func (s *server) CallHTTP(stream pb.ServiceHub_CallHTTPServer) error {
	for {
		req, err := stream.Recv()
		_ = req
		if err != nil {
			return err
		}

		p := &chRelay{req: req, resCh: make(chan *pb.CallHTTPResponse)}
		ch <- p
		res := <-p.resCh
		if err := stream.Send(res); err != nil {
			return err
		}
	}
}

func (s *server) RelayHTTP(stream pb.ServiceHub_RelayHTTPServer) error {
	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			log.Println("Client disconnected.")
			return ctx.Err()
		case p := <-ch:
			log.Println("received", p.req)
			if err := stream.Send(p.req); err != nil {
				return err
			}

			res, err := stream.Recv()
			p.resCh <- res
			if err != nil {
				return err
			}
		}

	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterServiceHubServer(s, &server{})
	log.Printf("Server listening on %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
