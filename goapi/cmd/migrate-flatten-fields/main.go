package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	serviceAccountPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH")
	if serviceAccountPath == "" {
		log.Fatal("FIREBASE_SERVICE_ACCOUNT_PATH environment variable is required")
	}

	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		log.Println("FIREBASE_PROJECT_ID not set, attempting to load from service account...")
		var err error
		projectID, err = loadProjectID(serviceAccountPath)
		if err != nil {
			log.Fatalf("Failed to load project ID: %v", err)
		}
	}

	collectionName := os.Getenv("FIRESTORE_COLLECTION")
	if collectionName == "" {
		collectionName = "individual_jobs"
	}

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer client.Close()

	log.Printf("🔥 Connected to Firestore: project=%s, collection=%s", projectID, collectionName)
	log.Println("⚠️  This will flatten sortable fields (publishTime, budget, etc.) to document root")
	log.Println("⏳ Starting migration...")

	if err := migrateCollection(ctx, client, collectionName); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("✅ Migration completed successfully!")
}

func migrateCollection(ctx context.Context, client *firestore.Client, collectionName string) error {
	// Limit to 200 docs for safety - can be run multiple times
	iter := client.Collection(collectionName).Limit(200).Documents(ctx)
	defer iter.Stop()

	updated := 0
	skipped := 0
	errors := 0
	batch := client.Batch()
	batchSize := 0
	maxBatchSize := 500 // Firestore batch limit

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("❌ Error reading document: %v", err)
			errors++
			continue
		}

		data := doc.Data()

		// Skip if already migrated (has publishTime at root)
		if _, exists := data["publishTime"]; exists {
			skipped++
			continue
		}

		updates := make(map[string]interface{})

		// Extract nested job data
		state, _ := data["state"].(map[string]interface{})

		var jobObj map[string]interface{}

		// Try jobDetails.job path first
		if jobDetails, ok := state["jobDetails"].(map[string]interface{}); ok {
			if job, ok := jobDetails["job"].(map[string]interface{}); ok {
				jobObj = job
			}
		}

		// Try job.job path as fallback
		if jobObj == nil {
			if jobState, ok := state["job"].(map[string]interface{}); ok {
				if job, ok := jobState["job"].(map[string]interface{}); ok {
					jobObj = job
				}
			}
		}

		if jobObj == nil {
			skipped++
			continue
		}

		// Flatten publishTime (most critical for sorting)
		if publishTime, ok := jobObj["publishTime"]; ok && publishTime != nil {
			updates["publishTime"] = publishTime
		}

		// Flatten postedOn as fallback
		if postedOn, ok := jobObj["postedOn"]; ok && postedOn != nil {
			updates["postedOn"] = postedOn
		}

		// Flatten createdOn
		if createdOn, ok := jobObj["createdOn"]; ok && createdOn != nil {
			updates["createdOn"] = createdOn
		}

		// Flatten budget amount
		if budget, ok := jobObj["budget"].(map[string]interface{}); ok {
			if amount, ok := budget["amount"]; ok && amount != nil {
				updates["budgetAmount"] = amount
			}
		}

		// Flatten fixed amount (alternative budget field)
		if amount, ok := jobObj["amount"].(map[string]interface{}); ok {
			if fixedAmt, ok := amount["amount"]; ok && fixedAmt != nil {
				updates["fixedAmount"] = fixedAmt
			}
		}

		// Flatten hourly budget
		if hourlyMax, ok := jobObj["hourlyBudgetMax"]; ok && hourlyMax != nil {
			updates["hourlyBudgetMax"] = hourlyMax
		}
		if hourlyMin, ok := jobObj["hourlyBudgetMin"]; ok && hourlyMin != nil {
			updates["hourlyBudgetMin"] = hourlyMin
		}

		// Only update if we have fields to flatten
		if len(updates) == 0 {
			skipped++
			continue
		}

		// Add to batch
		batch.Set(doc.Ref, updates, firestore.MergeAll)
		batchSize++
		updated++

		// Commit batch if it reaches max size
		if batchSize >= maxBatchSize {
			if _, err := batch.Commit(ctx); err != nil {
				log.Printf("❌ Failed to commit batch: %v", err)
				errors += batchSize
			} else {
				log.Printf("✅ Committed batch of %d documents", batchSize)
			}
			batch = client.Batch()
			batchSize = 0
		}

		if updated%10 == 0 {
			log.Printf("📊 Progress: %d updated, %d skipped, %d errors", updated, skipped, errors)
		}
	}

	// Commit remaining batch
	if batchSize > 0 {
		if _, err := batch.Commit(ctx); err != nil {
			log.Printf("❌ Failed to commit final batch: %v", err)
			errors += batchSize
		} else {
			log.Printf("✅ Committed final batch of %d documents", batchSize)
		}
	}

	log.Printf("📊 Final stats: %d updated, %d skipped, %d errors", updated, skipped, errors)

	if skipped > 0 && skipped == 200 {
		log.Println("ℹ️  Note: Processed 200 docs limit. Run again to continue migration.")
	}

	return nil
}

func loadProjectID(serviceAccountPath string) (string, error) {
	data, err := os.ReadFile(serviceAccountPath)
	if err != nil {
		return "", fmt.Errorf("failed to read service account file: %w", err)
	}

	var config struct {
		ProjectID string `json:"project_id"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("failed to parse service account JSON: %w", err)
	}

	if config.ProjectID == "" {
		return "", fmt.Errorf("project_id not found in service account file")
	}

	return config.ProjectID, nil
}