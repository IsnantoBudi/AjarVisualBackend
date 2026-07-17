package main

import (
	"log"
	"net/http"
	"strings"

	"ajarvisual-backend/config"
	"ajarvisual-backend/handlers"

	"github.com/joho/godotenv"
)

func setupApp() http.Handler {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system env")
	}

	config.ConnectDB()

	mux := http.NewServeMux()

	// Daftarkan route menggunakan standard http.ServeMux (Go 1.22+ pattern)
	mux.HandleFunc("POST /api/generate", handlers.GenerateWorksheet)
	mux.HandleFunc("GET /api/history", handlers.GetAllHistory)
	mux.HandleFunc("GET /api/history/{id}", handlers.GetWorksheetByID)
	mux.HandleFunc("DELETE /api/history/{id}", handlers.DeleteWorksheet)
	mux.HandleFunc("POST /api/history/{id}/add-soal", handlers.AddSoalToWorksheet)
	mux.HandleFunc("POST /api/regenerate-image", handlers.RegenerateImage)
	mux.HandleFunc("GET /api/image-proxy", handlers.ProxyImage)

	// Health check
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","message":"AjarVisual API is running"}`))
	})

	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigins := map[string]bool{
			"http://localhost:3002": true,
		}

		frontendURL := config.Getenv("FRONTEND_URL")
		if frontendURL != "" {
			for _, url := range strings.Split(frontendURL, ",") {
				url = strings.TrimSpace(url)
				url = strings.TrimRight(url, "/")
				if url != "" {
					allowedOrigins[url] = true
				}
			}
		}

		origin := r.Header.Get("Origin")
		isAllowed := false

		if origin != "" {
			// Dynamic checks: vercel.app domains or localhost
			if strings.HasSuffix(origin, ".vercel.app") || 
			   strings.Contains(origin, "localhost:") || 
			   strings.Contains(origin, "127.0.0.1:") ||
			   origin == "https://ajar-visual.vercel.app" {
				isAllowed = true
			} else if allowedOrigins[origin] {
				isAllowed = true
			}
		}

		if isAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
