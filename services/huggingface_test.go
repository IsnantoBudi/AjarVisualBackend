package services

import (
	"os"
	"testing"
	"github.com/joho/godotenv"
)

func TestQueryHuggingFace(t *testing.T) {
	// Try to load .env from parent directory if needed
	_ = godotenv.Load("../.env")

	hfToken := os.Getenv("HF_TOKEN")
	if hfToken == "" {
		t.Skip("HF_TOKEN not set, skipping integration test")
	}

	prompt := "A small tabby cat reading a book, cartoon style"
	imgData, contentType, err := QueryHuggingFace(prompt)
	
	if err != nil {
		t.Fatalf("QueryHuggingFace failed: %v", err)
	}

	if len(imgData) == 0 {
		t.Error("Returned image data is empty")
	}

	if contentType == "" {
		t.Error("Content-Type is empty")
	}
	
	t.Logf("Success! Received %d bytes of %s", len(imgData), contentType)
}
