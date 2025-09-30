package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
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
	requestTimeout = 20 * time.Second

	// Cache TTLs
	jobsCacheTTL    = 5 * time.Minute
	jobListCacheTTL = 5 * time.Minute
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

	// Cache management endpoints
	group.GET("/cache/stats", s.handleCacheStats)
	group.DELETE("/cache/clear", s.handleClearCache)

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
// @Security ApiKeyAuth
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
	// Generate cache key from query parameters
	cacheKey := generateCacheKey("jobs", c.Request.URL.Query())

	// Try to get from cache
	var cachedResponse JobsResponse
	if err := s.redisClient.Get(c.Request.Context(), cacheKey, &cachedResponse); err == nil {
		s.redisClient.Incr(c.Request.Context(), "cache:stats:hits")
		log.Printf("💚 Cache HIT for /jobs (key: %s)", cacheKey[len(cacheKey)-16:])
		c.JSON(http.StatusOK, cachedResponse)
		return
	}

	// Cache miss - query Firestore
	s.redisClient.Incr(c.Request.Context(), "cache:stats:misses")
	log.Printf("💔 Cache MISS for /jobs (key: %s)", cacheKey[len(cacheKey)-16:])

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

	response := JobsResponse{
		Success:     true,
		Data:        dtos,
		Count:       len(dtos),
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	}

	// Cache the response
	if err := s.redisClient.Set(c.Request.Context(), cacheKey, response, jobsCacheTTL); err != nil {
		log.Printf("⚠️ Failed to cache response: %v", err)
	} else {
		log.Printf("💾 Cached response for %v", jobsCacheTTL)
	}

	c.JSON(http.StatusOK, response)
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
	// Generate cache key from query parameters
	cacheKey := generateCacheKey("job-list", c.Request.URL.Query())

	// Try to get from cache
	var cachedResponse JobListResponse
	if err := s.redisClient.Get(c.Request.Context(), cacheKey, &cachedResponse); err == nil {
		s.redisClient.Incr(c.Request.Context(), "cache:stats:hits")
		log.Printf("💚 Cache HIT for /job-list (key: %s)", cacheKey[len(cacheKey)-16:])
		c.JSON(http.StatusOK, cachedResponse)
		return
	}

	// Cache miss - query Firestore
	s.redisClient.Incr(c.Request.Context(), "cache:stats:misses")
	log.Printf("💔 Cache MISS for /job-list (key: %s)", cacheKey[len(cacheKey)-16:])

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

	response := JobListResponse{
		Success:     true,
		Data:        dtos,
		Count:       len(dtos),
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	}

	// Cache the response
	if err := s.redisClient.Set(c.Request.Context(), cacheKey, response, jobListCacheTTL); err != nil {
		log.Printf("⚠️ Failed to cache response: %v", err)
	} else {
		log.Printf("💾 Cached response for %v", jobListCacheTTL)
	}

	c.JSON(http.StatusOK, response)
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

	// Build Firestore query with native ordering
	query := s.client.Collection(s.collectionName).Query

	// Use Firestore native ordering with flattened fields
	var orderField string
	var orderDir firestore.Direction
	needsInMemorySort := false

	switch opts.SortField {
	case SortPublishTime:
		// Use flattened publishTime field at root level
		orderField = "publishTime"
		if opts.SortAscending {
			orderDir = firestore.Asc
		} else {
			orderDir = firestore.Desc
		}
	case SortLastVisited:
		orderField = "scrape_metadata.last_visited_at"
		if opts.SortAscending {
			orderDir = firestore.Asc
		} else {
			orderDir = firestore.Desc
		}
	case SortBudget:
		// Use flattened budget fields, but still need in-memory sort to handle both fixed and hourly
		orderField = "budgetAmount"
		orderDir = firestore.Desc
		needsInMemorySort = true
	default:
		// Default to publishTime descending for best user experience
		orderField = "publishTime"
		orderDir = firestore.Desc
	}

	query = query.OrderBy(orderField, orderDir)

	// Calculate fetch limit
	fetchLimit := (opts.Limit + opts.Offset) * 3
	if needsInMemorySort {
		fetchLimit = fetchLimit * 2 // Need more for budget sorting
	}
	if fetchLimit < 100 {
		fetchLimit = 100
	}
	if fetchLimit > 500 {
		fetchLimit = 500
	}

	query = query.Limit(fetchLimit)

	iter := query.Documents(ctx)
	defer iter.Stop()

	results := make([]JobRecord, 0, opts.Limit)
	docCount := 0

	for {
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
		docCount++

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

			results = append(results, job)
		}
	}

	log.Printf("📊 Fetched %d docs from Firestore (ordered by %s %v), filtered to %d results", docCount, orderField, orderDir, len(results))

	// In-memory sorting only if needed (budget sorting)
	if needsInMemorySort {
		sortJobs(results, opts)
	}

	if opts.Offset > 0 {
		if opts.Offset >= len(results) {
			return []JobRecord{}, nil
		}
		results = results[opts.Offset:]
	}
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

	// Build Firestore query with native ordering
	query := s.client.Collection(s.jobListCollection).Query

	// Use Firestore native ordering - publishedOn and last_visited_at are at root level
	var orderField string
	var orderDir firestore.Direction

	switch opts.SortField {
	case SortPublishTime:
		orderField = "publishedOn"
		if opts.SortAscending {
			orderDir = firestore.Asc
		} else {
			orderDir = firestore.Desc
		}
	case SortLastVisited:
		orderField = "scrape_metadata.last_visited_at"
		if opts.SortAscending {
			orderDir = firestore.Asc
		} else {
			orderDir = firestore.Desc
		}
	default:
		// Default to last_visited_at descending
		orderField = "scrape_metadata.last_visited_at"
		orderDir = firestore.Desc
	}

	query = query.OrderBy(orderField, orderDir)

	// With native ordering, fetch 2-3x limit to account for filtering
	fetchLimit := opts.Limit * 3
	if fetchLimit < 100 {
		fetchLimit = 100
	}
	if fetchLimit > 500 {
		fetchLimit = 500
	}

	query = query.Limit(fetchLimit)

	iter := query.Documents(ctx)
	defer iter.Stop()

	results := make([]JobSummaryRecord, 0, opts.Limit)
	docCount := 0

	for {
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
		docCount++

		record, err := transformJobListDocument(doc)
		if err != nil {
			log.Printf("Skipping job_list document %s: %v", doc.Ref.ID, err)
			continue
		}

		if !applyJobListFilters(record, opts) {
			continue
		}

		results = append(results, *record)
	}

	log.Printf("📊 Fetched %d docs from Firestore (ordered by %s %v), filtered to %d results", docCount, orderField, orderDir, len(results))

	// Results are already ordered by Firestore, just apply limit
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

// handleCacheStats returns cache hit/miss statistics
// @Summary Get cache statistics
// @Description Returns cache hit/miss ratio and performance metrics
// @Tags cache
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} JobsResponse
// @Failure 500 {object} JobsResponse
// @Security ApiKeyAuth
// @Router /cache/stats [get]
func (s *Server) handleCacheStats(c *gin.Context) {
	stats, err := s.redisClient.GetStats(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to get cache stats: %v", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// handleClearCache clears all response caches
// @Summary Clear all response caches
// @Description Removes all cached responses (does not affect API key cache)
// @Tags cache
// @Produce json
// @Success 200 {object} JobsResponse
// @Failure 401 {object} JobsResponse
// @Failure 500 {object} JobsResponse
// @Security ApiKeyAuth
// @Router /cache/clear [delete]
func (s *Server) handleClearCache(c *gin.Context) {
	// Clear response caches (keys starting with "response:")
	ctx := c.Request.Context()
	iter := s.redisClient.client.Scan(ctx, 0, "response:*", 0).Iterator()
	count := 0
	for iter.Next(ctx) {
		if err := s.redisClient.Delete(ctx, iter.Val()); err != nil {
			log.Printf("Failed to delete cache key %s: %v", iter.Val(), err)
		} else {
			count++
		}
	}
	if err := iter.Err(); err != nil {
		respondError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to clear cache: %v", err))
		return
	}

	log.Printf("🗑️ Cleared %d cache entries", count)
	c.JSON(http.StatusOK, JobsResponse{
		Success:     true,
		Message:     fmt.Sprintf("Cleared %d cache entries", count),
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

// generateCacheKey creates a deterministic cache key from query parameters
func generateCacheKey(endpoint string, queryParams map[string][]string) string {
	// Sort keys for deterministic output
	keys := make([]string, 0, len(queryParams))
	for k := range queryParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build query string
	var parts []string
	for _, k := range keys {
		values := queryParams[k]
		sort.Strings(values) // Sort values too
		for _, v := range values {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
		}
	}

	queryString := strings.Join(parts, "&")
	fullKey := fmt.Sprintf("%s?%s", endpoint, queryString)

	// Hash for shorter key
	hash := sha256.Sum256([]byte(fullKey))
	return fmt.Sprintf("response:%s:%s", endpoint, hex.EncodeToString(hash[:])[:16])
}
