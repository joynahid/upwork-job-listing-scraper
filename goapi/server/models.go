package server

import (
	"time"
)

// APIKey represents an API key with metadata
type APIKey struct {
	Key        string    `json:"key" firestore:"key"`
	ExpiryTime time.Time `json:"expiry_time" firestore:"expiry_time"`
	Source     string    `json:"source" firestore:"source"`
	CreatedAt  time.Time `json:"created_at" firestore:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" firestore:"updated_at"`
	IsActive   bool      `json:"is_active" firestore:"is_active"`
}

// IsExpired checks if the API key has expired
func (ak *APIKey) IsExpired() bool {
	return time.Now().UTC().After(ak.ExpiryTime)
}

// IsValid checks if the API key is valid (active and not expired)
func (ak *APIKey) IsValid() bool {
	return ak.IsActive && !ak.IsExpired()
}

// APIAccess represents the api_accesses document structure in Firestore
type APIAccess struct {
	APIKeys   []APIKey  `json:"api_keys" firestore:"api_keys"`
	UpdatedAt time.Time `json:"updated_at" firestore:"updated_at"`
}
