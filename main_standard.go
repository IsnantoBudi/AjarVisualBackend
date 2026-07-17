//go:build !wasm
// +build !wasm

package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	handler := setupApp()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("AjarVisual API running locally on port %s", port)
	if err := http.ListenAndServe(":" + port, handler); err != nil {
		log.Fatalf("Failed to run local server: %v", err)
	}
}
