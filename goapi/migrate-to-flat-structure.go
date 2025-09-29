package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

// Old structure (array-based)
type OldAPIKey struct {
	Key        string    `json:"key" firestore:"key"`
	ExpiryTime time.Time `json:"expiry_time" firestore:"expiry_time"`
	Source     string    `json:"source" firestore:"source"`
	CreatedAt  time.Time `json:"created_at" firestore:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" firestore:"updated_at"`
	IsActive   bool      `json:"is_active" firestore:"is_active"`
}

type OldAPIAccess struct {
	APIKeys   []OldAPIKey `json:"api_keys" firestore:"api_keys"`
	UpdatedAt time.Time   `json:"updated_at" firestore:"updated_at"`
}

// New structure (flat)
type NewAPIKey struct {
	Key        string    `json:"key" firestore:"key"`
	ExpiryTime time.Time `json:"expiry_time" firestore:"expiry_time"`
	Source     string    `json:"source" firestore:"source"`
	CreatedAt  time.Time `json:"created_at" firestore:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" firestore:"updated_at"`
	IsActive   bool      `json:"is_active" firestore:"is_active"`
	KeyHash    string    `json:"key_hash" firestore:"key_hash"`
}

type APIKeyMetadata struct {
	TotalKeys    int       `json:"total_keys" firestore:"total_keys"`
	ActiveKeys   int       `json:"active_keys" firestore:"active_keys"`
	LastUpdated  time.Time `json:"last_updated" firestore:"last_updated"`
	LastKeyAdded time.Time `json:"last_key_added" firestore:"last_key_added"`
}

const (
	// Old collection/document
	oldCollection = "api_accesses"
	oldDocument   = "document"

	// New collections
	apiKeysCollection     = "api_keys"
	apiKeysMetaCollection = "api_keys_meta"
	apiKeysMetaDocument   = "metadata"
)

func main() {
	ctx := context.Background()

	// Initialize Firestore client
	serviceAccountPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH")
	if serviceAccountPath == "" {
		log.Fatal("FIREBASE_SERVICE_ACCOUNT_PATH environment variable is required")
	}

	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		log.Fatal("FIREBASE_PROJECT_ID environment variable is required")
	}

	client, err := firestore.NewClient(ctx, projectID,
		option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer client.Close()

	fmt.Println("🔄 Starting migration from array-based to flat structure...")

	// Step 1: Read the old document
	fmt.Println("📖 Reading old API keys document...")
	oldDoc, err := client.Collection(oldCollection).Doc(oldDocument).Get(ctx)
	if err != nil {
		log.Fatalf("Failed to read old document: %v", err)
	}

	var oldData OldAPIAccess
	if err := oldDoc.DataTo(&oldData); err != nil {
		log.Fatalf("Failed to parse old document: %v", err)
	}

	fmt.Printf("✅ Found %d API keys to migrate\n", len(oldData.APIKeys))

	// Step 2: Migrate each API key to individual documents
	fmt.Println("🚀 Migrating API keys to flat structure...")

	var migratedKeys []NewAPIKey
	activeCount := 0

	for i, oldKey := range oldData.APIKeys {
		newKey := NewAPIKey{
			Key:        oldKey.Key,
			ExpiryTime: oldKey.ExpiryTime,
			Source:     oldKey.Source,
			CreatedAt:  oldKey.CreatedAt,
			UpdatedAt:  oldKey.UpdatedAt,
			IsActive:   oldKey.IsActive,
			KeyHash:    hashAPIKey(oldKey.Key),
		}

		// Create individual document
		docRef := client.Collection(apiKeysCollection).Doc(newKey.KeyHash)
		if _, err := docRef.Set(ctx, newKey); err != nil {
			log.Printf("❌ Failed to migrate key %d: %v", i+1, err)
			continue
		}

		migratedKeys = append(migratedKeys, newKey)
		if newKey.IsActive {
			activeCount++
		}

		fmt.Printf("✅ Migrated key %d/%d: %s\n", i+1, len(oldData.APIKeys), sanitizeAPIKey(newKey.Key))
	}

	// Step 3: Create metadata document
	fmt.Println("📊 Creating metadata document...")

	var lastKeyAdded time.Time
	for _, key := range migratedKeys {
		if key.CreatedAt.After(lastKeyAdded) {
			lastKeyAdded = key.CreatedAt
		}
	}

	metadata := APIKeyMetadata{
		TotalKeys:    len(migratedKeys),
		ActiveKeys:   activeCount,
		LastUpdated:  time.Now().UTC(),
		LastKeyAdded: lastKeyAdded,
	}

	metaRef := client.Collection(apiKeysMetaCollection).Doc(apiKeysMetaDocument)
	if _, err := metaRef.Set(ctx, metadata); err != nil {
		log.Printf("❌ Failed to create metadata: %v", err)
	} else {
		fmt.Println("✅ Created metadata document")
	}

	// Step 4: Backup old document (rename it)
	fmt.Println("💾 Backing up old document...")

	backupData := map[string]interface{}{
		"original_data": oldData,
		"migrated_at":   time.Now().UTC(),
		"migrated_keys": len(migratedKeys),
	}

	backupRef := client.Collection(oldCollection).Doc(oldDocument + "_backup_" + time.Now().Format("20060102_150405"))
	if _, err := backupRef.Set(ctx, backupData); err != nil {
		log.Printf("⚠️ Warning: Failed to create backup: %v", err)
	} else {
		fmt.Println("✅ Created backup document")
	}

	// Step 5: Summary
	fmt.Println("\n🎉 Migration completed!")
	fmt.Printf("📊 Summary:\n")
	fmt.Printf("   • Migrated: %d/%d keys\n", len(migratedKeys), len(oldData.APIKeys))
	fmt.Printf("   • Active keys: %d\n", activeCount)
	fmt.Printf("   • Inactive keys: %d\n", len(migratedKeys)-activeCount)
	fmt.Printf("   • New collection: %s\n", apiKeysCollection)
	fmt.Printf("   • Metadata collection: %s\n", apiKeysMetaCollection)

	fmt.Println("\n⚠️ Next steps:")
	fmt.Println("   1. Test the new API endpoints")
	fmt.Println("   2. Update your application to use the new structure")
	fmt.Printf("   3. Delete the old document: %s/%s\n", oldCollection, oldDocument)
	fmt.Println("   4. Update any scripts or tools to use the new structure")

	fmt.Println("\n🔧 Test commands:")
	fmt.Println("   go run goapi/cmd/manage-keys/main.go -action=list")
	fmt.Println("   curl -H \"X-API-KEY: your-key\" http://localhost:8080/api-keys/refresh-cache")
}

// hashAPIKey generates a SHA256 hash for an API key string
func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// sanitizeAPIKey returns a sanitized version of the API key for logging
func sanitizeAPIKey(key string) string {
	if len(key) <= 12 {
		return "***"
	}
	return key[:8] + "****" + key[len(key)-4:]
}
