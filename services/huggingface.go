package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"ajarvisual-backend/config"
	"ajarvisual-backend/models"
)

// HF_MODEL is the model ID for FLUX
const HF_MODEL = "black-forest-labs/FLUX.1-schnell"

// computeDeterministicSeed converts prompt into a stable positive integer seed
func computeDeterministicSeed(prompt string) int {
	h := sha256.Sum256([]byte(prompt))
	seed := int(h[0]) | (int(h[1]) << 8) | (int(h[2]) << 16) | ((int(h[3]) & 0x7f) << 24)
	if seed < 0 {
		return -seed
	}
	return seed
}

// QueryHuggingFace calls the HuggingFace Inference API with optimized education clipart prompt
func QueryHuggingFace(prompt string) ([]byte, string, error) {
	hfToken := config.Getenv("HF_TOKEN")
	if hfToken == "" {
		return nil, "", fmt.Errorf("HF_TOKEN not found in environment")
	}

	apiURL := fmt.Sprintf("https://api-inference.huggingface.co/models/%s", HF_MODEL)

	refinedPrompt := prompt + ", isolated on pure solid white background, flat 2d vector clipart for kids, clean sharp lines, vibrant colors, minimal digital art"

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

	log.Printf("[HuggingFace] Sending POST request to: %s", apiURL)
	client := &http.Client{Timeout: 25 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[HuggingFace] Request error: %v", err)
		return nil, "", fmt.Errorf("huggingface request error: %w", err)
	}
	defer resp.Body.Close()
	log.Printf("[HuggingFace] Response received: status %d", resp.StatusCode)

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

// QueryPollinationsImage fetches an image from Pollinations.ai with deterministic seed and prompt optimization
func QueryPollinationsImage(prompt string, model string) ([]byte, string, error) {
	refinedPrompt := prompt + ", isolated on pure solid white background, flat 2d vector clipart for kids, clean sharp lines, vibrant colors, minimal digital art"
	encoded := url.QueryEscape(refinedPrompt)
	seed := computeDeterministicSeed(prompt)

	modelParam := ""
	if model == "flux" {
		modelParam = "&model=flux"
	}

	apiURL := fmt.Sprintf("https://image.pollinations.ai/prompt/%s?width=512&height=512&nologo=true&seed=%d%s", encoded, seed, modelParam)

	log.Printf("[Pollinations] Sending GET request with seed %d: %s", seed, apiURL[:min(len(apiURL), 90)]+"...")
	client := &http.Client{Timeout: 25 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		log.Printf("[Pollinations] Request error: %v", err)
		return nil, "", fmt.Errorf("pollinations request error: %w", err)
	}
	defer resp.Body.Close()
	log.Printf("[Pollinations] Response received: status %d", resp.StatusCode)

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

var (
	// imageMutex serializes generation requests to avoid 429 Too Many Requests
	imageMutex sync.Mutex
)

// GenerateImage checks TiDB/D1 cache first, then generates and saves if not found.
func GenerateImage(prompt string, model string) ([]byte, string, error) {
	// Normalize model name
	normalizedModel := "pollinations"
	if model == "flux" || model == "flux-schnell" || model == "huggingface" || model == HF_MODEL {
		normalizedModel = "flux"
	}

	// 1. Compute deterministic SHA-256 hash
	hashBytes := sha256.Sum256([]byte(normalizedModel + ":" + prompt))
	promptHash := hex.EncodeToString(hashBytes[:])

	// 2. Check Database Cache (CONCURRENT - lock-free query)
	var cached models.CachedImage
	err := config.DB.QueryRow("SELECT prompt_hash, prompt, image_data, content_type, created_at FROM cached_images WHERE prompt_hash = ?", promptHash).
		Scan(&cached.PromptHash, &cached.Prompt, &cached.ImageData, &cached.ContentType, &cached.CreatedAt)
	if err == nil && len(cached.ImageData) > 0 {
		log.Printf("[image cache] HIT for [%s]: %s (hash: %s)", normalizedModel, prompt[:min(len(prompt), 40)], promptHash)
		return cached.ImageData, cached.ContentType, nil
	}

	log.Printf("[image cache] MISS for [%s]: %s. Waiting for generator...", normalizedModel, prompt[:min(len(prompt), 40)])

	// 3. Serial execution queue
	imageMutex.Lock()
	defer imageMutex.Unlock()

	// Double-Check Locking Pattern
	err = config.DB.QueryRow("SELECT prompt_hash, prompt, image_data, content_type, created_at FROM cached_images WHERE prompt_hash = ?", promptHash).
		Scan(&cached.PromptHash, &cached.Prompt, &cached.ImageData, &cached.ContentType, &cached.CreatedAt)
	if err == nil && len(cached.ImageData) > 0 {
		log.Printf("[image cache] HIT (double-check) for [%s]: %s (hash: %s)", normalizedModel, prompt[:min(len(prompt), 40)], promptHash)
		return cached.ImageData, cached.ContentType, nil
	}

	// Small pause between serial external requests
	time.Sleep(300 * time.Millisecond)

	var imgData []byte
	var ct string

	if normalizedModel == "flux" {
		// Try HuggingFace FLUX.1 first, fallback to Pollinations FLUX engine
		imgData, ct, err = QueryHuggingFace(prompt)
		if err == nil {
			log.Printf("[image] HuggingFace FLUX OK for: %s", prompt[:min(len(prompt), 60)])
		} else {
			log.Printf("[image] HuggingFace FLUX failed (%v), fallback to Pollinations FLUX Engine...", err)
			imgData, ct, err = QueryPollinationsImage(prompt, "flux")
		}
	} else {
		// Default: Pollinations Turbo
		imgData, ct, err = QueryPollinationsImage(prompt, "pollinations")
		if err == nil {
			log.Printf("[image] Pollinations Turbo OK for: %s", prompt[:min(len(prompt), 60)])
		} else {
			log.Printf("[image] Pollinations failed (%v), fallback to HuggingFace FLUX...", err)
			imgData, ct, err = QueryHuggingFace(prompt)
		}
	}

	if err != nil {
		return nil, "", fmt.Errorf("all image providers failed: %w", err)
	}

	if ct == "" {
		ct = "image/jpeg"
	}

	// 4. Save to database cache for permanent fast retrieval
	_, dbErr := config.DB.Exec(
		"INSERT INTO cached_images (prompt_hash, prompt, image_data, content_type) VALUES (?, ?, ?, ?)",
		promptHash, prompt, imgData, ct,
	)
	if dbErr != nil {
		log.Printf("[image cache] Notice: DB cache insert result: %v", dbErr)
	} else {
		log.Printf("[image cache] Successfully cached image for hash: %s (%d bytes)", promptHash, len(imgData))
	}

	return imgData, ct, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GenerateImageURLFromOs returns the backend proxy URL for the given prompt and model
func GenerateImageURLFromOs(prompt string, model string) string {
	backendURL := config.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8080"
	}
	encoded := url.QueryEscape(prompt)
	if model != "" {
		return fmt.Sprintf("%s/api/image-proxy?prompt=%s&image_model=%s", backendURL, encoded, url.QueryEscape(model))
	}
	return fmt.Sprintf("%s/api/image-proxy?prompt=%s", backendURL, encoded)
}
