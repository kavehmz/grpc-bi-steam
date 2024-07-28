package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	pb "githib.com/regentmarkets/segmented/hub"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func relayCall(client pb.ServiceHubClient, req *pb.CallHTTPRequest) *pb.CallHTTPResponse {
	stream, err := client.RelayCall(context.Background())
	if err != nil {
		log.Fatalf("could not call CallHTTP: %v", err)
	}

	log.Println("sending the requests", req)
	if err := stream.Send(req); err != nil {
		log.Fatalf("could not send request: %v", err)
	}

	res, err := stream.Recv()
	if err != nil {
		log.Fatalf("could not receive response: %v", err)
	}

	return res
}

func relay(w http.ResponseWriter, r *http.Request) {
	conn, err := grpc.NewClient(":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewServiceHubClient(conn)

	headersMap := make(map[string]string)
	for key, values := range r.Header {
		// Join multiple values for the same key with a comma
		headersMap[key] = values[0] // Use values[0] if you want just the first value
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	bodyString := string(body)

	req := &pb.CallHTTPRequest{
		Method:  r.Method,
		Target:  r.URL.String(),
		Headers: headersMap,
		Body:    bodyString,
	}

	response := relayCall(client, req)
	w.WriteHeader(int(response.StatusCode))
	for key, value := range response.Headers {
		w.Header().Set(key, value)
	}
	_, err = w.Write([]byte(response.Body))
	if err != nil {
		fmt.Println("Error writing response body:", err)
	}
}

func main() {
	http.HandleFunc("/", relay)
	http.ListenAndServe(":8081", nil)
}
