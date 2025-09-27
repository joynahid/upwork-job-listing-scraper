package main

import (
	"log"
	"os"

	"upwork-job-api/docs"
	"upwork-job-api/server"
)

// @title Upwork Job API
// @version 1.0
// @description Provides filtered access to normalized Upwork job data stored in Firestore.
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-KEY
// @BasePath /
func main() {
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
	log.Printf("  GET /jobs   - Firestore-filtered jobs (requires X-API-KEY)")
	log.Printf("  GET /health - Health check (requires X-API-KEY)")

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
