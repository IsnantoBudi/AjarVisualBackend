package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

// HF_MODEL is the model ID.
const HF_MODEL = "black-forest-labs/FLUX.1-schnell"

// QueryHuggingFace calls the HuggingFace Inference API
func QueryHuggingFace(prompt string) ([]byte, string, error) {
	hfToken := os.Getenv("HF_TOKEN")
	if hfToken == "" {
		return nil, "", fmt.Errorf("HF_TOKEN not found in environment")
	}

	apiURL := fmt.Sprintf("https://router.huggingface.co/hf-inference/models/%s", HF_MODEL)

	refinedPrompt := prompt + ", cartoon style, vibrant colors, white background, educational illustration for kids, cute digital art, high resolution"

	payload := map[string]interface{}{
		"inputs": refinedPrompt,
		"options": map[string]interface{}{
			"wait_for_model": true,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, "", err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("Authorization", "Bearer "+hfToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("huggingface request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("[HuggingFace] Error %d: %s", resp.StatusCode, string(bodyBytes))
		return nil, "", fmt.Errorf("huggingface api error: status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	imgData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return imgData, contentType, nil
}

// QueryPollinationsImage fetches an image from Pollinations.ai
func QueryPollinationsImage(prompt string) ([]byte, string, error) {
	encoded := url.QueryEscape(prompt + ", cartoon style, educational, kids illustration, vibrant, cute, white background")
	apiURL := fmt.Sprintf("https://image.pollinations.ai/prompt/%s?width=512&height=512&nologo=true&seed=%d", encoded, time.Now().UnixMilli()%10000)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, "", fmt.Errorf("pollinations request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("pollinations error: status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	imgData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return imgData, contentType, nil
}

// GenerateImage tries Pollinations first (reliable, free), falls back to HuggingFace
func GenerateImage(prompt string) ([]byte, string, error) {
	// Try Pollinations first (free, no quota issues)
	imgData, ct, err := QueryPollinationsImage(prompt)
	if err == nil {
		log.Printf("[image] Pollinations OK for: %s", prompt[:min(len(prompt), 60)])
		return imgData, ct, nil
	}

	log.Printf("[image] Pollinations failed (%v), trying HuggingFace...", err)

	// Fallback: HuggingFace
	imgData, ct, err = QueryHuggingFace(prompt)
	if err != nil {
		return nil, "", fmt.Errorf("all image providers failed: %w", err)
	}
	return imgData, ct, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GenerateImageURLFromOs returns the backend proxy URL for the given prompt
func GenerateImageURLFromOs(prompt string) string {
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8080"
	}
	encoded := url.QueryEscape(prompt)
	return fmt.Sprintf("%s/api/image-proxy?prompt=%s", backendURL, encoded)
}
