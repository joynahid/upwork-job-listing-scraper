package main

import (
	"log"
	"os"

	"upwork-job-api/docs"
	"upwork-job-api/server"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// @title Upwork Job API
// @version 1.0
// @description API for accessing normalized Upwork job listings with advanced filtering capabilities. All endpoints require authentication via X-API-KEY header.
// @contact.name API Support
// @contact.email support@upworkjobapi.com
// @host localhost:8080
// @schemes http https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-KEY
// @BasePath /
func main() {
	// Register custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		server.RegisterCustomValidators(v)
		log.Println("✅ Custom validators registered")
	}

	srv, err := server.NewServer()
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}
	defer srv.Shutdown()

	docs.SwaggerInfo.BasePath = "/"

	router := srv.Router()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("📡 Gin server listening on port %s", port)
	log.Printf("Endpoints:")
	log.Printf("  GET    /jobs                      - Firestore-filtered jobs (requires X-API-KEY)")
	log.Printf("  GET    /health                    - Health check (requires X-API-KEY)")
	log.Printf("  POST   /api-keys/refresh-cache    - Refresh API keys cache (requires X-API-KEY)")
	log.Printf("  DELETE /api-keys/{key}/cache      - Clear specific API key cache (requires X-API-KEY)")
	log.Printf("  GET    /swagger/*                 - API documentation")

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
