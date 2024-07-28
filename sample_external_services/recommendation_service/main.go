package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type StatusResponse struct {
	Status string
	Time   string
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	response := StatusResponse{
		Status: "recommendation pending",
		Time:   time.Now().String(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/recommendation/list", statusHandler)
	http.ListenAndServe(":8081", nil)
}
