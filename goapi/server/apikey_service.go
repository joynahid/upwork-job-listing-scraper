package server

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
)

const (
	// Cache keys
	apiAccessCacheKey = "api_accesses:document"
	apiKeyCachePrefix = "api_key:"

	// Cache TTL
	apiAccessCacheTTL = 5 * time.Minute
	apiKeyCacheTTL    = 10 * time.Minute

	// Rate limiting
	firestoreQueryLimit = time.Second // Query Firestore at most once per second
)

// APIKeyService manages API key validation with caching and rate limiting
type APIKeyService struct {
	firestoreClient    *firestore.Client
	redisClient        *RedisClient
	lastFirestoreQuery time.Time
	queryMutex         sync.RWMutex
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(firestoreClient *firestore.Client, redisClient *RedisClient) *APIKeyService {
	return &APIKeyService{
		firestoreClient: firestoreClient,
		redisClient:     redisClient,
	}
}

// ValidateAPIKey validates an API key with caching and rate limiting
func (s *APIKeyService) ValidateAPIKey(ctx context.Context, key string) (*APIKey, error) {
	if key == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	// Try to get from cache first
	cacheKey := apiKeyCachePrefix + key
	var cachedKey APIKey

	if err := s.redisClient.Get(ctx, cacheKey, &cachedKey); err == nil {
		log.Printf("🔑 API key found in cache: %s", maskAPIKey(key))
		if cachedKey.IsValid() {
			return &cachedKey, nil
		}
		// Key is cached but invalid, remove from cache
		s.redisClient.Delete(ctx, cacheKey)
		return nil, fmt.Errorf("API key is expired or inactive")
	}

	// Not in cache, fetch from Firestore
	apiKey, err := s.fetchAPIKeyFromFirestore(ctx, key)
	if err != nil {
		return nil, err
	}

	if apiKey == nil {
		return nil, fmt.Errorf("API key not found")
	}

	// Cache the result
	if apiKey.IsValid() {
		s.redisClient.Set(ctx, cacheKey, apiKey, apiKeyCacheTTL)
		log.Printf("🔑 API key cached: %s", maskAPIKey(key))
	}

	return apiKey, nil
}

// fetchAPIKeyFromFirestore retrieves API key from Firestore with rate limiting
func (s *APIKeyService) fetchAPIKeyFromFirestore(ctx context.Context, key string) (*APIKey, error) {
	// Rate limiting: ensure we don't query Firestore more than once per second
	s.queryMutex.Lock()
	timeSinceLastQuery := time.Since(s.lastFirestoreQuery)
	if timeSinceLastQuery < firestoreQueryLimit {
		sleepTime := firestoreQueryLimit - timeSinceLastQuery
		log.Printf("⏱️ Rate limiting Firestore query, sleeping for %v", sleepTime)
		time.Sleep(sleepTime)
	}
	s.lastFirestoreQuery = time.Now()
	s.queryMutex.Unlock()

	// Try to get from cache first (api_accesses document)
	var apiAccess APIAccess
	if err := s.redisClient.Get(ctx, apiAccessCacheKey, &apiAccess); err == nil {
		log.Printf("📄 API accesses document found in cache")
		return s.findAPIKeyInDocument(&apiAccess, key), nil
	}

	// Cache miss, query Firestore
	log.Printf("🔥 Querying Firestore for api_accesses document")

	doc, err := s.firestoreClient.Collection("api_accesses").Doc("document").Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch api_accesses document: %w", err)
	}

	if err := doc.DataTo(&apiAccess); err != nil {
		return nil, fmt.Errorf("failed to parse api_accesses document: %w", err)
	}

	// Cache the document
	s.redisClient.Set(ctx, apiAccessCacheKey, &apiAccess, apiAccessCacheTTL)
	log.Printf("📄 API accesses document cached")

	return s.findAPIKeyInDocument(&apiAccess, key), nil
}

// findAPIKeyInDocument searches for an API key in the APIAccess document
func (s *APIKeyService) findAPIKeyInDocument(apiAccess *APIAccess, key string) *APIKey {
	for _, apiKey := range apiAccess.APIKeys {
		if apiKey.Key == key {
			return &apiKey
		}
	}
	return nil
}

// RefreshCache forces a refresh of the API accesses cache
func (s *APIKeyService) RefreshCache(ctx context.Context) error {
	// Delete cached document
	if err := s.redisClient.Delete(ctx, apiAccessCacheKey); err != nil {
		log.Printf("Warning: failed to delete cached api_accesses: %v", err)
	}

	// Query Firestore to refresh cache
	log.Printf("🔄 Refreshing API accesses cache")

	doc, err := s.firestoreClient.Collection("api_accesses").Doc("document").Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh api_accesses document: %w", err)
	}

	var apiAccess APIAccess
	if err := doc.DataTo(&apiAccess); err != nil {
		return fmt.Errorf("failed to parse api_accesses document: %w", err)
	}

	// Cache the refreshed document
	if err := s.redisClient.Set(ctx, apiAccessCacheKey, &apiAccess, apiAccessCacheTTL); err != nil {
		log.Printf("Warning: failed to cache refreshed api_accesses: %v", err)
	}

	log.Printf("✅ API accesses cache refreshed")
	return nil
}

// ClearAPIKeyCache removes a specific API key from cache
func (s *APIKeyService) ClearAPIKeyCache(ctx context.Context, key string) error {
	cacheKey := apiKeyCachePrefix + key
	return s.redisClient.Delete(ctx, cacheKey)
}

// Note: maskAPIKey function is already defined in util.go
