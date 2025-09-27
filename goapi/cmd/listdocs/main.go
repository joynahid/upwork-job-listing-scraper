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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	iter := client.Collection(collection).Documents(ctx)
	defer iter.Stop()

	count := 0
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("iteration error: %v", err)
		}
		fmt.Println(doc.Ref.ID)
		count++
	}

	fmt.Printf("total docs: %d\n", count)
}
