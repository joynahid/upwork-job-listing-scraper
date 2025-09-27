package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"

	"upwork-job-api/server"
)

func main() {
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		log.Fatal("FIREBASE_PROJECT_ID is required")
	}
	serviceAccountPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH")
	if serviceAccountPath == "" {
		log.Fatal("FIREBASE_SERVICE_ACCOUNT_PATH is required")
	}
	collection := os.Getenv("FIRESTORE_COLLECTION")
	if collection == "" {
		collection = "individual_jobs"
	}

	limit := flag.Int("limit", 3, "number of documents to dump")
	rawDump := flag.Bool("raw", false, "dump raw document data instead of transformed")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	iter := client.Collection(collection).Documents(ctx)

	found := 0
	for found < *limit {
		doc, err := iter.Next()
		if err != nil {
			if err.Error() == "no more items in iterator" {
				log.Println("no more documents available")
				break
			}
			log.Fatalf("failed to fetch document: %v", err)
		}

		payload := map[string]any{
			"document_id": doc.Ref.ID,
		}

		if *rawDump {
			payload["raw"] = doc.Data()
		} else {
			records, err := server.DebugTransformDocument(doc)
			if err != nil {
				log.Printf("transform error for %s: %v", doc.Ref.ID, err)
				continue
			}
			payload["records"] = records
		}

		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			log.Fatalf("marshal error: %v", err)
		}

		fmt.Println(string(data))
		found++
	}
}
