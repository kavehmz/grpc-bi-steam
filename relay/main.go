package main

import (
	"context"
	"flag"
	"log"
	"time"

	pb "githib.com/regentmarkets/segmented/hub"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func callHTTP(client pb.ServiceHubClient) {
	stream, err := client.CallHTTP(context.Background())
	if err != nil {
		log.Fatalf("could not call CallHTTP: %v", err)
	}

	req := &pb.CallHTTPRequest{
		Method:  "GET",
		Url:     "http://example.com",
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

func relayHTTP(client pb.ServiceHubClient) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := client.RelayHTTP(ctx)
	if err != nil {
		log.Fatalf("could not call RelayHTTP: %v", err)
	}
	// for {
	req, err := stream.Recv()
	if err != nil {
		log.Fatalf("could not receive response: %v", err)
	}

	res := &pb.CallHTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       req.Body + " :::: " + time.Now().String(),
	}

	if err := stream.Send(res); err != nil {
		log.Fatalf("could not send request: %v", err)
	}
	// }
}

func main() {
	conn, err := grpc.NewClient(":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewServiceHubClient(conn)

	relay := flag.Bool("relay", false, "enable relay")
	flag.Parse()

	if *relay {
		relayHTTP(client)
	} else {
		callHTTP(client)
		callHTTP(client)
		callHTTP(client)
	}
}
