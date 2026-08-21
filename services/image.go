package services

import (
	"fmt"
	"net/url"

	"ajarvisual-backend/config"
)

func GenerateImageURL(prompt string, imageModel string) string {
	backendURL := config.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8080"
	}
	encoded := url.QueryEscape(prompt)
	if imageModel != "" {
		return fmt.Sprintf("%s/api/image-proxy?prompt=%s&image_model=%s", backendURL, encoded, url.QueryEscape(imageModel))
	}
	return fmt.Sprintf("%s/api/image-proxy?prompt=%s", backendURL, encoded)
}
