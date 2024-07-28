package main

import (
	"context"
	"flag"
	"log"

	pb "githib.com/regentmarkets/segmented/hub"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func relayCall(client pb.ServiceHubClient) {
	stream, err := client.RelayCall(context.Background())
	if err != nil {
		log.Fatalf("could not call CallHTTP: %v", err)
	}

	req := &pb.CallHTTPRequest{
		Method:  "GET",
		Target:  "/notification/latest",
		Headers: map[string]string{"Accept": "application/json"},
		Body:    "",
	}

	log.Println("sending the requests", req)
	if err := stream.Send(req); err != nil {
		log.Fatalf("could not send request: %v", err)
	}

	res, err := stream.Recv()
	if err != nil {
		log.Fatalf("could not receive response: %v", err)
	}

	log.Printf("Response: %v", res)
}

func serveServiceCalls(client pb.ServiceHubClient) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := client.ServeServiceCalls(ctx)
	if err != nil {
		log.Fatalf("could not call RelayHTTP: %v", err)
	}

	s := &pb.ServiceCall{
		Details: &pb.ServiceCallDetails{
			ServiceName: "notification",
		},
		Response: nil,
	}
	if err := stream.Send(s); err != nil {
		log.Fatalf("could not send request: %v", err)
	}

	for i := 0; i < 1; i++ {
		req, err := stream.Recv()
		if err != nil {
			log.Fatalf("could not receive response: %v", err)
		}

		res := &pb.CallHTTPResponse{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       "served:" + req.Target,
		}

		s := &pb.ServiceCall{
			Details:  nil,
			Response: res,
		}
		if err := stream.Send(s); err != nil {
			log.Fatalf("could not send request: %v", err)
		}

	}
}

func main() {
	conn, err := grpc.NewClient(":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewServiceHubClient(conn)

	relay := flag.Bool("serve", false, "accept a service call for this service")
	flag.Parse()

	if *relay {
		serveServiceCalls(client)
	} else {
		relayCall(client)
		relayCall(client)
		relayCall(client)
	}
}
