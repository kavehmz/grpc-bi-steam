package main

import (
	"encoding/json"
	"net/http"
)

type StatusResponse struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	response := StatusResponse{
		Status: "pending",
		Count:  100,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/notification/status", statusHandler)
	http.ListenAndServe(":8080", nil)
}
