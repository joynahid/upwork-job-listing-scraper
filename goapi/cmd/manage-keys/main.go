package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// APIKey represents an API key document in Firestore
type APIKey struct {
	Key        string    `json:"key" firestore:"key"`
	ExpiryTime time.Time `json:"expiry_time" firestore:"expiry_time"`
	Source     string    `json:"source" firestore:"source"`
	CreatedAt  time.Time `json:"created_at" firestore:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" firestore:"updated_at"`
	IsActive   bool      `json:"is_active" firestore:"is_active"`
	KeyHash    string    `json:"key_hash" firestore:"key_hash"`
}

// APIKeyMetadata represents metadata for the API keys collection
type APIKeyMetadata struct {
	TotalKeys    int       `json:"total_keys" firestore:"total_keys"`
	ActiveKeys   int       `json:"active_keys" firestore:"active_keys"`
	LastUpdated  time.Time `json:"last_updated" firestore:"last_updated"`
	LastKeyAdded time.Time `json:"last_key_added" firestore:"last_key_added"`
}

// Collection names
const (
	apiKeysCollection     = "api_keys"
	apiKeysMetaCollection = "api_keys_meta"
	apiKeysMetaDocument   = "metadata"
)

func main() {
	var (
		action = flag.String("action", "", "Action: add, update, deactivate, activate, list")
		key    = flag.String("key", "", "API key (for update/deactivate/activate)")
		prefix = flag.String("prefix", "ak_live", "Prefix for new key")
		expiry = flag.String("expiry", "2025-12-31T23:59:59Z", "Expiry time")
		source = flag.String("source", "go_script", "Source of the key")
	)
	flag.Parse()

	if *action == "" {
		fmt.Println("Usage: go run main.go -action=<action> [options]")
		fmt.Println("\nActions:")
		fmt.Println("  add        - Add a new API key")
		fmt.Println("  update     - Update expiry time of existing key")
		fmt.Println("  activate   - Activate an existing key")
		fmt.Println("  deactivate - Deactivate an existing key")
		fmt.Println("  list       - List all API keys")
		fmt.Println("\nOptions:")
		fmt.Println("  -key       - API key (required for update/activate/deactivate)")
		fmt.Println("  -prefix    - Prefix for new key (default: ak_live)")
		fmt.Println("  -expiry    - Expiry time in RFC3339 format (default: 2025-12-31T23:59:59Z)")
		fmt.Println("  -source    - Source description (default: go_script)")
		fmt.Println("\nExamples:")
		fmt.Println("  go run main.go -action=add -prefix=ak_prod -source=manual")
		fmt.Println("  go run main.go -action=deactivate -key=ak_live_1234567890abcdef")
		fmt.Println("  go run main.go -action=list")
		os.Exit(1)
	}

	ctx := context.Background()

	// Initialize Firestore
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

	switch *action {
	case "add":
		err = addAPIKey(ctx, client, *prefix, *expiry, *source)
	case "update":
		if *key == "" {
			log.Fatal("Key is required for update action")
		}
		err = updateAPIKey(ctx, client, *key, map[string]interface{}{
			"expiry_time": parseTime(*expiry),
			"updated_at":  time.Now().UTC(),
		})
	case "activate":
		if *key == "" {
			log.Fatal("Key is required for activate action")
		}
		err = updateAPIKey(ctx, client, *key, map[string]interface{}{
			"is_active":  true,
			"updated_at": time.Now().UTC(),
		})
	case "deactivate":
		if *key == "" {
			log.Fatal("Key is required for deactivate action")
		}
		err = updateAPIKey(ctx, client, *key, map[string]interface{}{
			"is_active":  false,
			"updated_at": time.Now().UTC(),
		})
	case "list":
		err = listAPIKeys(ctx, client)
	default:
		log.Fatal("Unknown action. Use: add, update, activate, deactivate, list")
	}

	if err != nil {
		log.Fatalf("Operation failed: %v", err)
	}
}

func addAPIKey(ctx context.Context, client *firestore.Client, prefix, expiry, source string) error {
	newKey := APIKey{
		Key:        generateAPIKey(prefix),
		ExpiryTime: parseTime(expiry),
		Source:     source,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
		IsActive:   true,
	}

	// Generate hash for document ID
	newKey.KeyHash = hashAPIKey(newKey.Key)

	return client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// Check if key already exists
		docRef := client.Collection(apiKeysCollection).Doc(newKey.KeyHash)
		_, err := tx.Get(docRef)
		if err == nil {
			return fmt.Errorf("API key already exists")
		}

		// Add the new key
		if err := tx.Set(docRef, newKey); err != nil {
			return fmt.Errorf("failed to add API key: %w", err)
		}

		// Update metadata
		metaRef := client.Collection(apiKeysMetaCollection).Doc(apiKeysMetaDocument)
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
			if newKey.IsActive {
				metadata.ActiveKeys++
			}
			metadata.LastUpdated = time.Now().UTC()
			metadata.LastKeyAdded = time.Now().UTC()
		}

		if err := tx.Set(metaRef, metadata); err != nil {
			return fmt.Errorf("failed to update metadata: %w", err)
		}

		fmt.Printf("✅ Added new API key: %s\n", newKey.Key)
		fmt.Printf("   Document ID: %s\n", newKey.KeyHash[:12]+"...")
		fmt.Printf("   Expires: %s\n", newKey.ExpiryTime.Format("2006-01-02 15:04:05 UTC"))
		fmt.Printf("   Source: %s\n", newKey.Source)
		return nil
	})
}

func updateAPIKey(ctx context.Context, client *firestore.Client, keyToUpdate string, updates map[string]interface{}) error {
	keyHash := hashAPIKey(keyToUpdate)
	docRef := client.Collection(apiKeysCollection).Doc(keyHash)

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

	// Log what was updated
	if expiry, ok := updates["expiry_time"].(time.Time); ok {
		fmt.Printf("📅 Updated expiry to: %s\n", expiry.Format("2006-01-02 15:04:05 UTC"))
	}
	if active, ok := updates["is_active"].(bool); ok {
		status := "activated"
		if !active {
			status = "deactivated"
		}
		fmt.Printf("🔄 Key %s: %s\n", status, sanitizeAPIKey(keyToUpdate))
	}

	fmt.Printf("✅ Updated API key: %s\n", sanitizeAPIKey(keyToUpdate))
	return nil
}

func listAPIKeys(ctx context.Context, client *firestore.Client) error {
	// Get metadata first
	metaDoc, err := client.Collection(apiKeysMetaCollection).Doc(apiKeysMetaDocument).Get(ctx)
	var metadata APIKeyMetadata
	if err == nil {
		metaDoc.DataTo(&metadata)
		fmt.Printf("📊 Metadata (Updated: %s)\n", metadata.LastUpdated.Format("2006-01-02 15:04:05 UTC"))
		fmt.Printf("   Total Keys: %d, Active Keys: %d\n", metadata.TotalKeys, metadata.ActiveKeys)
		fmt.Printf("   Last Key Added: %s\n\n", metadata.LastKeyAdded.Format("2006-01-02 15:04:05 UTC"))
	}

	// List all API keys
	iter := client.Collection(apiKeysCollection).Documents(ctx)
	defer iter.Stop()

	var apiKeys []APIKey
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate API keys: %w", err)
		}

		var apiKey APIKey
		if err := doc.DataTo(&apiKey); err != nil {
			log.Printf("Warning: failed to parse API key document %s: %v", doc.Ref.ID, err)
			continue
		}

		apiKeys = append(apiKeys, apiKey)
	}

	fmt.Printf("📋 Found %d API keys:\n\n", len(apiKeys))

	for i, key := range apiKeys {
		status := "🟢 Active"
		if !key.IsActive {
			status = "🔴 Inactive"
		}
		if time.Now().UTC().After(key.ExpiryTime) {
			if key.IsActive {
				status = "⏰ Expired"
			} else {
				status = "⏰ Expired & Inactive"
			}
		}

		fmt.Printf("%d. Key: %s\n", i+1, sanitizeAPIKey(key.Key))
		fmt.Printf("   Doc ID: %s...\n", key.KeyHash[:16])
		fmt.Printf("   Status: %s\n", status)
		fmt.Printf("   Expires: %s\n", key.ExpiryTime.Format("2006-01-02 15:04:05 UTC"))
		fmt.Printf("   Source: %s\n", key.Source)
		fmt.Printf("   Created: %s\n", key.CreatedAt.Format("2006-01-02 15:04:05 UTC"))
		fmt.Printf("   Updated: %s\n", key.UpdatedAt.Format("2006-01-02 15:04:05 UTC"))
		fmt.Println()
	}

	// Summary
	active := 0
	expired := 0
	inactive := 0
	for _, key := range apiKeys {
		if !key.IsActive {
			inactive++
		} else if time.Now().UTC().After(key.ExpiryTime) {
			expired++
		} else {
			active++
		}
	}

	fmt.Printf("📊 Summary: %d active, %d expired, %d inactive\n", active, expired, inactive)
	return nil
}

func generateAPIKey(prefix string) string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("Failed to generate random bytes: %v", err)
	}
	return prefix + "_" + hex.EncodeToString(bytes)
}

func parseTime(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		log.Fatalf("Invalid time format: %s (use RFC3339 format like 2025-12-31T23:59:59Z)", timeStr)
	}
	return t.UTC()
}

// hashAPIKey generates a SHA256 hash for an API key string
func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// sanitizeAPIKey returns a sanitized version of the API key for logging
func sanitizeAPIKey(key string) string {
	if len(key) <= 12 {
		return strings.Repeat("*", len(key))
	}
	return key[:8] + "****" + key[len(key)-4:]
}
