package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultLimit   = 20
	maxLimit       = 50
	maxDocsCap     = 500
	docsMultiplier = 8
	requestTimeout = 20 * time.Second
)

type Server struct {
	rootCtx           context.Context
	cancelRoot        context.CancelFunc
	client            *firestore.Client
	redisClient       *RedisClient
	apiKeyService     *APIKeyService
	collectionName    string
	jobListCollection string
	apiKey            string // Legacy API key for backward compatibility
}

// NewServer creates a server with Firestore client and configuration.
func NewServer() (*Server, error) {
	apiKey := mustEnv("API_KEY")
	log.Printf("🔐 Legacy API key loaded (%d chars): %s", len(apiKey), maskAPIKey(apiKey))

	serviceAccountPath := mustEnv("FIREBASE_SERVICE_ACCOUNT_PATH")

	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		var err error
		projectID, err = loadProjectID(serviceAccountPath)
		if err != nil {
			return nil, fmt.Errorf("failed to determine Firestore project ID: %w", err)
		}
	}

	collectionName := os.Getenv("FIRESTORE_COLLECTION")
	if collectionName == "" {
		collectionName = "individual_jobs"
	}

	jobListCollection := os.Getenv("FIRESTORE_JOB_LIST_COLLECTION")
	if jobListCollection == "" {
		jobListCollection = "job_list"
	}

	ctx, cancel := context.WithCancel(context.Background())
	client, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create Firestore client: %w", err)
	}

	log.Printf("🔥 Firestore client initialized: project=%s, collections=%s,%s", projectID, collectionName, jobListCollection)

	// Initialize Redis client
	redisClient, err := NewRedisClient()
	if err != nil {
		cancel()
		client.Close()
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	// Initialize API key service
	apiKeyService := NewAPIKeyService(client, redisClient)

	return &Server{
		rootCtx:           ctx,
		cancelRoot:        cancel,
		client:            client,
		redisClient:       redisClient,
		apiKeyService:     apiKeyService,
		collectionName:    collectionName,
		jobListCollection: jobListCollection,
		apiKey:            apiKey,
	}, nil
}

// Shutdown releases Firestore and Redis resources.
func (s *Server) Shutdown() {
	if s.cancelRoot != nil {
		s.cancelRoot()
	}
	if s.redisClient != nil {
		if err := s.redisClient.Close(); err != nil {
			log.Printf("error closing Redis client: %v", err)
		}
	}
	if s.client != nil {
		if err := s.client.Close(); err != nil {
			log.Printf("error closing Firestore client: %v", err)
		}
	}
}

// Router constructs the Gin router with middleware and routes.
func (s *Server) Router() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(s.loggingMiddleware())

	group := router.Group("/")
	group.Use(s.authMiddleware())
	group.GET("/health", s.handleHealth)
	group.GET("/jobs", s.handleJobs)
	group.GET("/job-list", s.handleJobList)

	// API key management endpoints
	group.POST("/api-keys/refresh-cache", s.handleRefreshAPIKeysCache)
	group.DELETE("/api-keys/:key/cache", s.handleClearAPIKeyCache)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}

// loggingMiddleware emits structured request logs.
func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Round(time.Millisecond)
		log.Printf("➡️ %s %s (status=%d, duration=%s, ip=%s)", c.Request.Method, c.Request.RequestURI, c.Writer.Status(), duration, clientIP(c))
	}
}

// authMiddleware ensures requests include a valid API key.
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-KEY")
		if apiKey == "" {
			respondError(c, http.StatusUnauthorized, "Missing X-API-KEY header")
			c.Abort()
			return
		}

		// First try the new API key service
		validAPIKey, err := s.apiKeyService.ValidateAPIKey(c.Request.Context(), apiKey)
		if err == nil && validAPIKey != nil {
			// Store API key info in context for potential use in handlers
			c.Set("api_key_info", validAPIKey)
			c.Next()
			return
		}

		// Fallback to legacy API key for backward compatibility
		if apiKey == s.apiKey {
			log.Printf("🔑 Using legacy API key: %s", maskAPIKey(apiKey))
			c.Next()
			return
		}

		// Log the validation error for debugging
		log.Printf("🚫 API key validation failed: %v", err)
		respondError(c, http.StatusUnauthorized, "Invalid or expired X-API-KEY")
		c.Abort()
	}
}

// handleHealth is a simple readiness endpoint.
// @Summary Health check
// @Description Returns a 200 response when the API is up.
// @Tags health
// @Produce json
// @Success 200 {object} JobsResponse
// @Failure 401 {object} JobsResponse
// @Router /health [get]
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, JobsResponse{
		Success:     true,
		Message:     "API is healthy",
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	})
}

// handleJobs queries Firestore with filters and returns normalized job data.
// @Summary List jobs
// @Description Retrieve normalized job documents with optional filters.
// @Tags jobs
// @Produce json
// @Param limit query int false "Number of records (1-50)"
// @Param payment_verified query bool false "Filter by payment verification"
// @Param category query string false "Filter by category slug"
// @Param category_group query string false "Filter by category group slug"
// @Param status query string false "Filter by job status (open|closed or numeric code)"
// @Param job_type query string false "Filter by job type (hourly|fixed-price or numeric code)"
// @Param contractor_tier query string false "Filter by contractor tier (entry|intermediate|expert or numeric code)"
// @Param country query string false "Filter by buyer country"
// @Param tags query string false "Comma-separated required tags"
// @Param posted_after query string false "ISO timestamp lower bound"
// @Param posted_before query string false "ISO timestamp upper bound"
// @Param budget_min query number false "Minimum fixed budget"
// @Param budget_max query number false "Maximum fixed budget"
// @Param sort query string false "Sort mode (posted_on_asc, posted_on_desc, last_visited_asc, last_visited_desc)"
// @Success 200 {object} JobsResponse
// @Failure 400 {object} JobsResponse
// @Failure 401 {object} JobsResponse
// @Failure 500 {object} JobsResponse
// @Security ApiKeyAuth
// @Router /jobs [get]
func (s *Server) handleJobs(c *gin.Context) {
	opts, err := parseFilterOptions(c.Request.URL.Query())
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("🎯 Firestore filter options: %s", formatFilterOptions(opts))

	jobs, err := s.queryJobs(c.Request.Context(), opts)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	dtos := make([]JobDTO, 0, len(jobs))
	for _, job := range jobs {
		dtos = append(dtos, job.ToDTO())
	}

	c.JSON(http.StatusOK, JobsResponse{
		Success:     true,
		Data:        dtos,
		Count:       len(dtos),
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	})
}

// handleJobList returns job_list entries with optional filtering.
// @Summary List job summaries
// @Description Retrieve Upwork job summaries from the job_list collection with optional filters.
// @Tags job-list
// @Produce json
// @Param limit query int false "Number of records (1-50)"
// @Param payment_verified query bool false "Filter by client payment verification"
// @Param country query string false "Filter by client country"
// @Param skills query string false "Comma-separated list of required skill labels"
// @Param job_type query string false "Filter by job type (hourly|fixed-price or numeric code)"
// @Param duration query string false "Filter by duration label"
// @Param hourly_min query number false "Minimum hourly budget"
// @Param hourly_max query number false "Maximum hourly budget"
// @Param budget_min query number false "Minimum fixed budget"
// @Param budget_max query number false "Maximum fixed budget"
// @Param search query string false "Case-insensitive search term for title or description"
// @Param sort query string false "Sort mode (published_on_asc, published_on_desc, last_visited_asc, last_visited_desc)"
// @Success 200 {object} JobListResponse
// @Failure 400 {object} JobListResponse
// @Failure 401 {object} JobListResponse
// @Failure 500 {object} JobListResponse
// @Security ApiKeyAuth
// @Router /job-list [get]
func (s *Server) handleJobList(c *gin.Context) {
	opts, err := parseJobListFilterOptions(c.Request.URL.Query())
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("🎯 Job list filter options: %s", formatJobListFilterOptions(opts))

	jobs, err := s.queryJobList(c.Request.Context(), opts)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	dtos := make([]JobSummaryDTO, 0, len(jobs))
	for _, job := range jobs {
		dtos = append(dtos, job.ToDTO())
	}

	c.JSON(http.StatusOK, JobListResponse{
		Success:     true,
		Data:        dtos,
		Count:       len(dtos),
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) queryJobs(requestCtx context.Context, opts FilterOptions) ([]JobRecord, error) {
	ctx := s.rootCtx
	if requestCtx != nil {
		if deadline, ok := requestCtx.Deadline(); ok {
			remaining := time.Until(deadline)
			if remaining > 0 && remaining < requestTimeout {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(s.rootCtx, remaining)
				defer cancel()
			}
		}
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	iter := s.client.Collection(s.collectionName).Documents(ctx)
	defer iter.Stop()

	results := make([]JobRecord, 0, opts.Limit)
	lookedAt := 0
	targetMatches := opts.Limit + opts.Offset
	if targetMatches < opts.Limit {
		targetMatches = opts.Limit
	}
	maxDocs := int(math.Min(
		maxDocsCap,
		math.Max(float64(targetMatches)*docsMultiplier, float64(targetMatches*2)),
	))
	matched := 0

outer:
	for {
		if lookedAt >= maxDocs {
			break
		}

		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			if isContextCanceled(err) {
				return nil, fmt.Errorf("firestore query cancelled: %w", err)
			}
			return nil, fmt.Errorf("firestore query failed: %w", err)
		}
		lookedAt++

		records, err := transformDocument(doc)
		if err != nil {
			log.Printf("Skipping document %s: %v", doc.Ref.ID, err)
			continue
		}

		for _, rec := range records {
			job := rec
			if !applyFilters(&job, opts) {
				continue
			}
			matched++
			if matched <= opts.Offset {
				continue
			}

			results = append(results, job)
			if len(results) >= opts.Limit {
				break outer
			}
		}
	}

	sortJobs(results, opts)
	if len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

func (s *Server) queryJobList(requestCtx context.Context, opts JobListFilterOptions) ([]JobSummaryRecord, error) {
	ctx := s.rootCtx
	if requestCtx != nil {
		if deadline, ok := requestCtx.Deadline(); ok {
			remaining := time.Until(deadline)
			if remaining > 0 && remaining < requestTimeout {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(s.rootCtx, remaining)
				defer cancel()
			}
		}
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	iter := s.client.Collection(s.jobListCollection).Documents(ctx)
	defer iter.Stop()

	results := make([]JobSummaryRecord, 0, opts.Limit)
	lookedAt := 0
	maxDocs := int(math.Min(maxDocsCap, math.Max(float64(opts.Limit)*docsMultiplier, float64(opts.Limit*2))))

	for {
		if lookedAt >= maxDocs {
			break
		}

		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			if isContextCanceled(err) {
				return nil, fmt.Errorf("firestore query cancelled: %w", err)
			}
			return nil, fmt.Errorf("firestore query failed: %w", err)
		}
		lookedAt++

		record, err := transformJobListDocument(doc)
		if err != nil {
			log.Printf("Skipping job_list document %s: %v", doc.Ref.ID, err)
			continue
		}

		if !applyJobListFilters(record, opts) {
			continue
		}

		results = append(results, *record)
		if len(results) >= opts.Limit {
			break
		}
	}

	sortJobSummaries(results, opts)
	if len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

// handleRefreshAPIKeysCache forces a refresh of the API keys cache
// @Summary Refresh API keys cache
// @Description Forces a refresh of the API keys cache from Firestore
// @Tags api-keys
// @Produce json
// @Success 200 {object} JobsResponse
// @Failure 401 {object} JobsResponse
// @Failure 500 {object} JobsResponse
// @Security ApiKeyAuth
// @Router /api-keys/refresh-cache [post]
func (s *Server) handleRefreshAPIKeysCache(c *gin.Context) {
	if err := s.apiKeyService.RefreshCache(c.Request.Context()); err != nil {
		respondError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to refresh cache: %v", err))
		return
	}

	c.JSON(http.StatusOK, JobsResponse{
		Success:     true,
		Message:     "API keys cache refreshed successfully",
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	})
}

// handleClearAPIKeyCache clears a specific API key from cache
// @Summary Clear API key cache
// @Description Removes a specific API key from the cache
// @Tags api-keys
// @Produce json
// @Param key path string true "API key to clear from cache"
// @Success 200 {object} JobsResponse
// @Failure 401 {object} JobsResponse
// @Failure 500 {object} JobsResponse
// @Security ApiKeyAuth
// @Router /api-keys/{key}/cache [delete]
func (s *Server) handleClearAPIKeyCache(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		respondError(c, http.StatusBadRequest, "API key parameter is required")
		return
	}

	if err := s.apiKeyService.ClearAPIKeyCache(c.Request.Context(), key); err != nil {
		respondError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to clear cache: %v", err))
		return
	}

	c.JSON(http.StatusOK, JobsResponse{
		Success:     true,
		Message:     fmt.Sprintf("API key cache cleared for: %s", maskAPIKey(key)),
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	})
}

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, JobsResponse{
		Success:     false,
		Message:     message,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	})
}

func isContextCanceled(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return true
	}
	return status.Code(err) == codes.Canceled
}
