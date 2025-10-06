package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	projectID := "upwork-job-scraper-d1f2c"
	serviceAccountPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH")
	if serviceAccountPath == "" {
		log.Fatal("FIREBASE_SERVICE_ACCOUNT_PATH is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Get a few documents and check what fields they have
	fmt.Println("=== Checking document fields ===")
	iter := client.Collection("individual_jobs").Limit(3).Documents(ctx)
	count := 0
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error iterating: %v\n", err)
			break
		}
		count++
		data := doc.Data()

		fmt.Printf("\nDoc ID: %s\n", doc.Ref.ID)
		fmt.Printf("  Firestore UpdateTime: %s (%s ago)\n", doc.UpdateTime.Format(time.RFC3339), time.Since(doc.UpdateTime).Round(time.Hour))

		// Check for various time fields
		timeFields := []string{"publishTime", "published_on", "posted_on", "createdOn", "created_on", "publishedOn"}
		for _, field := range timeFields {
			if val, ok := data[field]; ok {
				if t, ok := val.(time.Time); ok {
					fmt.Printf("  %s: %s (%s ago)\n", field, t.Format(time.RFC3339), time.Since(t).Round(time.Hour))
				} else {
					fmt.Printf("  %s: %v (type: %T)\n", field, val, val)
				}
			}
		}

		if title, ok := data["title"].(string); ok {
			if len(title) > 60 {
				title = title[:60] + "..."
			}
			fmt.Printf("  title: %s\n", title)
		}

		// Show all top-level field names
		fmt.Printf("  Available fields: ")
		fieldNames := make([]string, 0, len(data))
		for k := range data {
			fieldNames = append(fieldNames, k)
		}
		fmt.Printf("%v\n", fieldNames)
	}

	fmt.Printf("\n\nTotal docs checked: %d\n", count)
}
