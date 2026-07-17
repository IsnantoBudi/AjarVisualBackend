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
	"os"
	"sync"
	"time"

	"ajarvisual-backend/config"
	"ajarvisual-backend/models"
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
	apiURL := fmt.Sprintf("https://image.pollinations.ai/prompt/%s?width=100&height=100&nologo=true&seed=%d", encoded, time.Now().UnixMilli()%10000)

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

var (
	// imageMutex memastikan generator gambar berjalan secara serial (antrean satu per satu)
	// untuk menghindari status 429 (Too Many Requests) dari API eksternal.
	imageMutex sync.Mutex
)

// GenerateImage checks TiDB cache first, then generates and saves if not found.
func GenerateImage(prompt string) ([]byte, string, error) {
	// 1. Hitung SHA-256 hash dari prompt
	hashBytes := sha256.Sum256([]byte(prompt))
	promptHash := hex.EncodeToString(hashBytes[:])

	// 2. Cari di database cache menggunakan config.DB (CONCURRENT - tanpa Lock)
	var cached models.CachedImage
	if err := config.DB.Where("prompt_hash = ?", promptHash).First(&cached).Error; err == nil {
		log.Printf("[image cache] HIT for: %s (hash: %s)", prompt[:min(len(prompt), 40)], promptHash)
		return cached.ImageData, cached.ContentType, nil
	}

	log.Printf("[image cache] MISS for: %s. Menunggu giliran generator serial...", prompt[:min(len(prompt), 40)])

	// 3. Masuk antrean serial (hanya satu generator yang berjalan pada satu waktu)
	imageMutex.Lock()
	defer imageMutex.Unlock()

	// 3a. Periksa kembali cache setelah mendapatkan lock (Double-Check Locking Pattern)
	// Hal ini untuk menghindari duplikasi request jika request identik sedang mengantre.
	if err := config.DB.Where("prompt_hash = ?", promptHash).First(&cached).Error; err == nil {
		log.Printf("[image cache] HIT (double-check) for: %s (hash: %s)", prompt[:min(len(prompt), 40)], promptHash)
		return cached.ImageData, cached.ContentType, nil
	}

	// Jeda singkat (misal 500ms) antar pemanggilan API serial agar API luar tidak merasa dibombardir
	time.Sleep(500 * time.Millisecond)

	var imgData []byte
	var ct string
	var err error

	// Coba Pollinations.ai dulu
	imgData, ct, err = QueryPollinationsImage(prompt)
	if err == nil {
		log.Printf("[image] Pollinations OK for: %s", prompt[:min(len(prompt), 60)])
	} else {
		log.Printf("[image] Pollinations failed (%v), trying HuggingFace...", err)
		// Fallback ke HuggingFace
		imgData, ct, err = QueryHuggingFace(prompt)
	}

	if err != nil {
		return nil, "", fmt.Errorf("all image providers failed: %w", err)
	}

	if ct == "" {
		ct = "image/jpeg"
	}

	// 4. Simpan ke database cache untuk request berikutnya
	newCached := models.CachedImage{
		PromptHash:  promptHash,
		Prompt:      prompt,
		ImageData:   imgData,
		ContentType: ct,
	}

	if dbErr := config.DB.Create(&newCached).Error; dbErr != nil {
		log.Printf("[image cache] Warning: failed to save cache to database: %v", dbErr)
	} else {
		log.Printf("[image cache] Saved to database for hash: %s", promptHash)
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
