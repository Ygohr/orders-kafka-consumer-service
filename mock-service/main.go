package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type OrderPayload struct {
	OrdemDeVenda string `json:"ordemDeVenda"`
	EtapaAtual   string `json:"etapaAtual"`
}

type Response struct {
	Message   string       `json:"message"`
	Data      OrderPayload `json:"data"`
	Timestamp time.Time    `json:"timestamp"`
}

func main() {
	http.HandleFunc("/api/v1/orders", handleOrderUpdate)

	fmt.Println("Started mock service on port 8081")

	log.Fatal(http.ListenAndServe(":8081", nil))
}

func handleOrderUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PATCH" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload OrderPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Error decoding payload", http.StatusBadRequest)
		return
	}

	response := Response{
		Message:   "Order updated successfully",
		Data:      payload,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)

	fmt.Printf("Order updated: %s - Step: %s\n", payload.OrdemDeVenda, payload.EtapaAtual)
}
