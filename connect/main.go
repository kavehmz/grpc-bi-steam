package main

import (
	"bytes"
	"context"
	"flag"
	"io"
	"log"
	"net/http"

	pb "githib.com/regentmarkets/segmented/hub"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func serveServiceCalls(targetService string, client pb.ServiceHubClient) {
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

		res, err := makeHTTPRequest(targetService, req)
		if err != nil {
			log.Fatalf("ccall to serviec failed: %v", err)
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
	targetService := flag.String("target", "", "The service URL to send the request to")
	flag.Parse()

	conn, err := grpc.NewClient(":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewServiceHubClient(conn)

	ch := make(chan bool, 5)
	for {
		ch <- true
		go func() {
			serveServiceCalls(*targetService, client)
			<-ch
		}()
	}

}

func makeHTTPRequest(service_url string, req *pb.CallHTTPRequest) (*pb.CallHTTPResponse, error) {
	client := &http.Client{}

	// Prepare the request
	httpReq, err := http.NewRequest(req.Method, service_url+req.Target, bytes.NewBuffer([]byte(req.Body)))
	if err != nil {
		return nil, err
	}

	// Add headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Perform the request
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Prepare the response headers
	responseHeaders := make(map[string]string)
	for key, values := range resp.Header {
		responseHeaders[key] = values[0] // Taking the first value for simplicity
	}

	// Prepare the Response struct
	response := &pb.CallHTTPResponse{
		StatusCode: int32(resp.StatusCode),
		Headers:    responseHeaders,
		Body:       string(body),
	}

	return response, nil
}
