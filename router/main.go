package main

import (
	"log"
	"net"
	"strings"

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

var chPool = make(map[string]chan *chRelay)

func init() {
	chPool["notification"] = make(chan *chRelay)
}

func (s *server) RelayCall(stream pb.ServiceHub_RelayCallServer) error {
	for {
		req, err := stream.Recv()
		_ = req
		if err != nil {
			return err
		}

		service, _ := getFirstPartOfPath(req.Target)
		log.Println(service)
		if ch, found := chPool[service]; found {
			p := &chRelay{req: req, resCh: make(chan *pb.CallHTTPResponse)}
			ch <- p
			res := <-p.resCh
			if err := stream.Send(res); err != nil {
				return err
			}
		} else {
			res := &pb.CallHTTPResponse{
				StatusCode: 500,
			}
			if err := stream.Send(res); err != nil {
				return err
			}
		}
	}
}

func (s *server) ServeServiceCalls(stream pb.ServiceHub_ServeServiceCallsServer) error {
	ctx := stream.Context()
	rec, err := stream.Recv()
	if err != nil {
		return err
	}

	ch := chPool[rec.Details.ServiceName]

	// here we route the calls to different channels
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

			rec, err := stream.Recv()
			if err != nil {
				return err
			}
			p.resCh <- rec.Response
		}

	}
}

// getFirstPartOfPath extracts the first part of the path from a URL string.
func getFirstPartOfPath(target string) (string, error) {
	parts := strings.Split(strings.TrimPrefix(target, "/"), "/")
	if len(parts) > 0 {
		return parts[0], nil
	}
	return "", nil
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
