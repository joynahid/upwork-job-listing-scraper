package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type JobData struct {
	JobID         string      `json:"job_id"`
	Data          interface{} `json:"data"`
	LastVisitedAt string      `json:"last_visited_at"`
}

type APIResponse struct {
	Success     bool      `json:"success"`
	Data        []JobData `json:"data,omitempty"`
	Count       int       `json:"count"`
	LastUpdated string    `json:"last_updated"`
	Message     string    `json:"message,omitempty"`
}

type JobAPI struct {
	db          *sql.DB
	cache       []JobData
	cacheMutex  sync.RWMutex
	lastUpdated time.Time
	apiKey      string
}

func NewJobAPI() (*JobAPI, error) {
	// Get configuration from environment
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API_KEY environment variable is required")
	}

	// Build PostgreSQL connection string
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}
	database := os.Getenv("POSTGRES_DB")
	if database == "" {
		database = "upwork_jobs"
	}
	username := os.Getenv("POSTGRES_USER")
	if username == "" {
		username = "postgres"
	}
	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		password = "postgres"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, username, password, database)

	// Open PostgreSQL database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Configure connection pool for better concurrent access
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test database connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	log.Printf("Connected to PostgreSQL database: %s:%s/%s", host, port, database)

	api := &JobAPI{
		db:          db,
		cache:       make([]JobData, 0),
		lastUpdated: time.Now(),
		apiKey:      apiKey,
	}

	// Initial cache load
	if err := api.refreshCache(); err != nil {
		log.Printf("Warning: Failed to load initial cache: %v", err)
	}

	return api, nil
}

func (api *JobAPI) refreshCache() error {
	// Force a new connection to ensure we see latest data
	if err := api.db.Ping(); err != nil {
		log.Printf("❌ Database ping failed: %v", err)
	}

	log.Printf("🔍 Refreshing cache from PostgreSQL database...")

	// Query latest jobs from PostgreSQL
	query := `
		SELECT job_id, data, last_visited_at 
		FROM job_entries 
		WHERE entry_type = 'latest'
		ORDER BY last_visited_at DESC
	`

	rows, err := api.db.Query(query)
	if err != nil {
		log.Printf("❌ Database query failed: %v", err)
		return fmt.Errorf("failed to query jobs: %v", err)
	}
	defer rows.Close()

	var newCache []JobData

	for rows.Next() {
		var jobID, dataStr string
		var lastVisitedAt time.Time
		var parsedData interface{}

		if err := rows.Scan(&jobID, &dataStr, &lastVisitedAt); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Parse JSON data
		if err := json.Unmarshal([]byte(dataStr), &parsedData); err != nil {
			log.Printf("Error parsing JSON for job %s: %v", jobID, err)
			continue
		}

		job := JobData{
			JobID:         jobID,
			Data:          parsedData,
			LastVisitedAt: lastVisitedAt.Format(time.RFC3339),
		}

		newCache = append(newCache, job)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %v", err)
	}

	// Update cache atomically
	api.cacheMutex.Lock()
	oldCount := len(api.cache)
	api.cache = newCache
	api.lastUpdated = time.Now()
	api.cacheMutex.Unlock()

	// Always log cache updates for debugging
	if len(newCache) != oldCount {
		log.Printf("✅ PostgreSQL cache updated: %d jobs (was %d)", len(newCache), oldCount)
	} else {
		log.Printf("🔄 PostgreSQL cache refreshed: %d jobs (no change)", len(newCache))
	}
	return nil
}

func (api *JobAPI) startBackgroundWorker() {
	ticker := time.NewTicker(500 * time.Millisecond) // More aggressive refresh for real-time updates
	logTicker := time.NewTicker(30 * time.Second)    // More frequent status logs for debugging

	go func() {
		defer ticker.Stop()
		defer logTicker.Stop()

		lastLogTime := time.Now()
		log.Printf("🔄 Background worker goroutine started")

		for {
			select {
			case <-ticker.C:
				if err := api.refreshCache(); err != nil {
					log.Printf("❌ Error refreshing cache: %v", err)
				}
			case <-logTicker.C:
				api.cacheMutex.RLock()
				jobCount := len(api.cache)
				api.cacheMutex.RUnlock()
				log.Printf("📊 Cache status: %d jobs, last updated: %s", jobCount, time.Since(lastLogTime).Round(time.Second))
				lastLogTime = time.Now()
			}
		}
	}()

	log.Printf("🔄 Background worker setup completed")
}

func (api *JobAPI) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-KEY")
		if apiKey != api.apiKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(APIResponse{
				Success: false,
				Message: "Invalid or missing X-API-KEY header",
			})
			return
		}
		next(w, r)
	}
}

func (api *JobAPI) handleJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	// Get cached data
	api.cacheMutex.RLock()
	jobs := make([]JobData, len(api.cache))
	copy(jobs, api.cache)
	lastUpdated := api.lastUpdated
	api.cacheMutex.RUnlock()

	response := APIResponse{
		Success:     true,
		Data:        jobs,
		Count:       len(jobs),
		LastUpdated: lastUpdated.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (api *JobAPI) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(APIResponse{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	// Simple health check
	response := APIResponse{
		Success: true,
		Message: "API is healthy",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (api *JobAPI) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Apply auth middleware to all routes
	mux.HandleFunc("/jobs", api.authMiddleware(api.handleJobs))
	mux.HandleFunc("/health", api.authMiddleware(api.handleHealth))

	return mux
}

func (api *JobAPI) close() {
	if api.db != nil {
		api.db.Close()
	}
}

func main() {
	log.Println("Starting Job API server...")

	// Initialize API
	api, err := NewJobAPI()
	if err != nil {
		log.Fatalf("Failed to initialize API: %v", err)
	}
	defer api.close()

	// Start background worker
	api.startBackgroundWorker()
	log.Println("🔄 Background worker started (500ms refresh interval, 30 second status logs)")

	// Log initial cache state
	api.cacheMutex.RLock()
	initialCount := len(api.cache)
	api.cacheMutex.RUnlock()
	log.Printf("📊 Initial cache loaded with %d jobs", initialCount)

	// Setup routes
	mux := api.setupRoutes()

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Endpoints:")
	log.Printf("  GET /jobs   - Returns all latest jobs (requires X-API-KEY header)")
	log.Printf("  GET /health - Health check (requires X-API-KEY header)")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
