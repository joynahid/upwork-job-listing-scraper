package server

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

const (
	// Collection names
	apiKeysCollection     = "api_keys"
	apiKeysMetaCollection = "api_keys_meta"
	apiKeysMetaDocument   = "metadata"

	// Cache keys
	apiKeyCachePrefix   = "api_key_hash:"
	apiKeysMetaCacheKey = "api_keys_meta"

	// Cache TTL
	apiKeyCacheTTL      = 15 * time.Minute // Individual keys cached longer
	apiKeysMetaCacheTTL = 5 * time.Minute  // Metadata cached shorter

	// Rate limiting
	firestoreQueryLimit = 500 * time.Millisecond // Allow more frequent queries for individual docs
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

	// Generate hash for cache key and document lookup
	keyHash := HashAPIKey(key)
	cacheKey := apiKeyCachePrefix + keyHash

	// Try to get from cache first
	var cachedKey APIKey
	if err := s.redisClient.Get(ctx, cacheKey, &cachedKey); err == nil {
		log.Printf("🔑 API key found in cache: %s", SanitizeAPIKeyForLog(key))
		if cachedKey.IsValid() {
			return &cachedKey, nil
		}
		// Key is cached but invalid, remove from cache
		s.redisClient.Delete(ctx, cacheKey)
		return nil, fmt.Errorf("API key is expired or inactive")
	}

	// Not in cache, fetch from Firestore by document ID (hash)
	apiKey, err := s.fetchAPIKeyByHash(ctx, keyHash)
	if err != nil {
		return nil, err
	}

	if apiKey == nil {
		return nil, fmt.Errorf("API key not found")
	}

	// Verify the actual key matches (hash collision protection)
	if apiKey.Key != key {
		log.Printf("🚨 Hash collision detected for key: %s", SanitizeAPIKeyForLog(key))
		return nil, fmt.Errorf("API key not found")
	}

	// Cache the result
	if apiKey.IsValid() {
		s.redisClient.Set(ctx, cacheKey, apiKey, apiKeyCacheTTL)
		log.Printf("🔑 API key cached: %s", SanitizeAPIKeyForLog(key))
	}

	return apiKey, nil
}

// fetchAPIKeyByHash retrieves API key from Firestore by hash with rate limiting
func (s *APIKeyService) fetchAPIKeyByHash(ctx context.Context, keyHash string) (*APIKey, error) {
	// Rate limiting: ensure we don't overwhelm Firestore
	s.queryMutex.Lock()
	timeSinceLastQuery := time.Since(s.lastFirestoreQuery)
	if timeSinceLastQuery < firestoreQueryLimit {
		sleepTime := firestoreQueryLimit - timeSinceLastQuery
		log.Printf("⏱️ Rate limiting Firestore query, sleeping for %v", sleepTime)
		time.Sleep(sleepTime)
	}
	s.lastFirestoreQuery = time.Now()
	s.queryMutex.Unlock()

	// Direct document lookup by hash (very fast)
	log.Printf("🔥 Querying Firestore for API key by hash: %s", keyHash[:12]+"...")

	doc, err := s.firestoreClient.Collection(apiKeysCollection).Doc(keyHash).Get(ctx)
	if err != nil {
		// Document not found is not an error, just return nil
		if status, ok := err.(interface{ Code() int }); ok && status.Code() == 5 { // NotFound
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch API key document: %w", err)
	}

	var apiKey APIKey
	if err := doc.DataTo(&apiKey); err != nil {
		return nil, fmt.Errorf("failed to parse API key document: %w", err)
	}

	log.Printf("📄 API key found in Firestore: %s", SanitizeAPIKeyForLog(apiKey.Key))
	return &apiKey, nil
}

// AddAPIKey adds a new API key to Firestore and updates metadata
func (s *APIKeyService) AddAPIKey(ctx context.Context, apiKey *APIKey) error {
	// Generate hash for document ID
	apiKey.GenerateKeyHash()

	// Use transaction to ensure consistency
	return s.firestoreClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// Check if key already exists
		docRef := s.firestoreClient.Collection(apiKeysCollection).Doc(apiKey.GetDocumentID())
		_, err := tx.Get(docRef)
		if err == nil {
			return fmt.Errorf("API key already exists")
		}

		// Add the new key
		if err := tx.Set(docRef, apiKey); err != nil {
			return fmt.Errorf("failed to add API key: %w", err)
		}

		// Update metadata
		metaRef := s.firestoreClient.Collection(apiKeysMetaCollection).Doc(apiKeysMetaDocument)
		metaDoc, err := tx.Get(metaRef)

		var metadata APIKeyMetadata
		if err != nil {
			// Create metadata if it doesn't exist
			metadata = APIKeyMetadata{
				TotalKeys:    1,
				ActiveKeys:   1,
				LastUpdated:  time.Now().UTC(),
				LastKeyAdded: time.Now().UTC(),
			}
		} else {
			metaDoc.DataTo(&metadata)
			metadata.TotalKeys++
			if apiKey.IsActive {
				metadata.ActiveKeys++
			}
			metadata.LastUpdated = time.Now().UTC()
			metadata.LastKeyAdded = time.Now().UTC()
		}

		return tx.Set(metaRef, metadata)
	})
}

// UpdateAPIKey updates an existing API key
func (s *APIKeyService) UpdateAPIKey(ctx context.Context, key string, updates map[string]interface{}) error {
	keyHash := HashAPIKey(key)
	docRef := s.firestoreClient.Collection(apiKeysCollection).Doc(keyHash)

	// Add updated_at timestamp
	updates["updated_at"] = time.Now().UTC()

	// Build update array
	var updateFields []firestore.Update
	for field, value := range updates {
		updateFields = append(updateFields, firestore.Update{
			Path:  field,
			Value: value,
		})
	}

	if _, err := docRef.Update(ctx, updateFields); err != nil {
		return fmt.Errorf("failed to update API key: %w", err)
	}

	// Clear cache for this key
	cacheKey := apiKeyCachePrefix + keyHash
	s.redisClient.Delete(ctx, cacheKey)

	log.Printf("✅ Updated API key: %s", SanitizeAPIKeyForLog(key))
	return nil
}

// DeleteAPIKey removes an API key (soft delete by setting inactive)
func (s *APIKeyService) DeleteAPIKey(ctx context.Context, key string) error {
	return s.UpdateAPIKey(ctx, key, map[string]interface{}{
		"is_active": false,
	})
}

// ListAPIKeys returns a list of API keys with optional filtering
func (s *APIKeyService) ListAPIKeys(ctx context.Context, filter APIKeyQueryFilter) ([]APIKey, error) {
	collection := s.firestoreClient.Collection(apiKeysCollection)
	query := collection.Query

	// Apply filters
	if filter.IsActive != nil {
		query = query.Where("is_active", "==", *filter.IsActive)
	}

	if filter.Source != "" {
		query = query.Where("source", "==", filter.Source)
	}

	if filter.ExpiryFrom != nil {
		query = query.Where("expiry_time", ">=", *filter.ExpiryFrom)
	}

	if filter.ExpiryTo != nil {
		query = query.Where("expiry_time", "<=", *filter.ExpiryTo)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	// Execute query
	iter := query.Documents(ctx)
	defer iter.Stop()

	var apiKeys []APIKey
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate API keys: %w", err)
		}

		var apiKey APIKey
		if err := doc.DataTo(&apiKey); err != nil {
			log.Printf("Warning: failed to parse API key document %s: %v", doc.Ref.ID, err)
			continue
		}

		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, nil
}

// RefreshCache clears all API key caches
func (s *APIKeyService) RefreshCache(ctx context.Context) error {
	log.Printf("🔄 Refreshing API key caches")

	// Clear metadata cache
	if err := s.redisClient.Delete(ctx, apiKeysMetaCacheKey); err != nil {
		log.Printf("Warning: failed to delete metadata cache: %v", err)
	}

	// Note: Individual key caches will be cleared when they expire or are accessed
	log.Printf("✅ API key caches refreshed")
	return nil
}

// ClearAPIKeyCache removes a specific API key from cache
func (s *APIKeyService) ClearAPIKeyCache(ctx context.Context, key string) error {
	keyHash := HashAPIKey(key)
	cacheKey := apiKeyCachePrefix + keyHash
	return s.redisClient.Delete(ctx, cacheKey)
}

// GetMetadata returns API key collection metadata
func (s *APIKeyService) GetMetadata(ctx context.Context) (*APIKeyMetadata, error) {
	// Try cache first
	var metadata APIKeyMetadata
	if err := s.redisClient.Get(ctx, apiKeysMetaCacheKey, &metadata); err == nil {
		return &metadata, nil
	}

	// Fetch from Firestore
	doc, err := s.firestoreClient.Collection(apiKeysMetaCollection).Doc(apiKeysMetaDocument).Get(ctx)
	if err != nil {
		// Return default metadata if document doesn't exist
		return &APIKeyMetadata{
			TotalKeys:   0,
			ActiveKeys:  0,
			LastUpdated: time.Now().UTC(),
		}, nil
	}

	if err := doc.DataTo(&metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Cache the metadata
	s.redisClient.Set(ctx, apiKeysMetaCacheKey, &metadata, apiKeysMetaCacheTTL)

	return &metadata, nil
}
