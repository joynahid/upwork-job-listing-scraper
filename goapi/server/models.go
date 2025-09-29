package server

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

// APIKey represents an API key document in Firestore
// Each API key is stored as a separate document in the api_keys collection
type APIKey struct {
	Key        string    `json:"key" firestore:"key"`
	ExpiryTime time.Time `json:"expiry_time" firestore:"expiry_time"`
	Source     string    `json:"source" firestore:"source"`
	CreatedAt  time.Time `json:"created_at" firestore:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" firestore:"updated_at"`
	IsActive   bool      `json:"is_active" firestore:"is_active"`
	// KeyHash is used as the document ID for fast lookups
	KeyHash string `json:"key_hash" firestore:"key_hash"`
}

// IsExpired checks if the API key has expired
func (ak *APIKey) IsExpired() bool {
	return time.Now().UTC().After(ak.ExpiryTime)
}

// IsValid checks if the API key is valid (active and not expired)
func (ak *APIKey) IsValid() bool {
	return ak.IsActive && !ak.IsExpired()
}

// GenerateKeyHash creates a SHA256 hash of the API key for use as document ID
func (ak *APIKey) GenerateKeyHash() {
	hash := sha256.Sum256([]byte(ak.Key))
	ak.KeyHash = hex.EncodeToString(hash[:])
}

// GetDocumentID returns the document ID for this API key
func (ak *APIKey) GetDocumentID() string {
	if ak.KeyHash == "" {
		ak.GenerateKeyHash()
	}
	return ak.KeyHash
}

// APIKeyMetadata represents metadata for the API keys collection
// Stored as a single document for collection-level information
type APIKeyMetadata struct {
	TotalKeys    int       `json:"total_keys" firestore:"total_keys"`
	ActiveKeys   int       `json:"active_keys" firestore:"active_keys"`
	LastUpdated  time.Time `json:"last_updated" firestore:"last_updated"`
	LastKeyAdded time.Time `json:"last_key_added" firestore:"last_key_added"`
}

// APIKeyQueryFilter represents filters for querying API keys
type APIKeyQueryFilter struct {
	IsActive   *bool      `json:"is_active,omitempty"`
	Source     string     `json:"source,omitempty"`
	ExpiryFrom *time.Time `json:"expiry_from,omitempty"`
	ExpiryTo   *time.Time `json:"expiry_to,omitempty"`
	Limit      int        `json:"limit,omitempty"`
}

// HashAPIKey generates a SHA256 hash for an API key string
func HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// SanitizeAPIKeyForLog returns a sanitized version of the API key for logging
func SanitizeAPIKeyForLog(key string) string {
	if len(key) <= 12 {
		return strings.Repeat("*", len(key))
	}
	return key[:8] + "****" + key[len(key)-4:]
}
