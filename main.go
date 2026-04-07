package main

import (
"log"
"os"

"ajarvisual-backend/config"
"ajarvisual-backend/handlers"

"github.com/gin-contrib/cors"
"github.com/gin-gonic/gin"
"github.com/joho/godotenv"
)

func main() {
if err := godotenv.Load(); err != nil {
log.Println("Warning: .env file not found, using system env")
}

config.ConnectDB()

r := gin.Default()

// CORS - allow frontend
frontendURL := os.Getenv("FRONTEND_URL")
if frontendURL == "" {
frontendURL = "http://localhost:3000"
}

r.Use(cors.New(cors.Config{
AllowOrigins:     []string{frontendURL, "http://localhost:3000"},
AllowMethods:     []string{"GET", "POST", "DELETE", "OPTIONS"},
AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
AllowCredentials: true,
}))

api := r.Group("/api")
{
api.POST("/generate", handlers.GenerateWorksheet)
api.GET("/history", handlers.GetAllHistory)
api.GET("/history/:id", handlers.GetWorksheetByID)
api.DELETE("/history/:id", handlers.DeleteWorksheet)
api.POST("/history/:id/add-soal", handlers.AddSoalToWorksheet)
api.POST("/regenerate-image", handlers.RegenerateImage)
api.GET("/image-proxy", handlers.ProxyImage)


// Health check
api.GET("/health", func(c *gin.Context) {
c.JSON(200, gin.H{"status": "ok", "message": "AjarVisual API is running"})
})
}

port := os.Getenv("PORT")
if port == "" {
port = "8080"
}

log.Printf(" AjarVisual API running on port %s", port)
r.Run(":" + port)
}
