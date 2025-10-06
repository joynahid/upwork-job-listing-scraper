package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"upwork-job-api/server"
)

type NormalizationStats struct {
	TotalDocs       int
	ValidDocs       int
	EmptyDocs       int
	InvalidDocs     int
	DeletedDocs     int
	NormalizedDocs  int
	UpdatedDocs     int
	SkippedDocs     int
	ErrorDocs       int
}

// NormalizedJobData represents the normalized structure to save in Firestore
type NormalizedJobData struct {
	// Core fields
	ID                   string                 `firestore:"id"`
	Title                string                 `firestore:"title,omitempty"`
	Description          string                 `firestore:"description,omitempty"`
	JobType              *int                   `firestore:"job_type,omitempty"`
	Status               *int                   `firestore:"status,omitempty"`
	ContractorTier       *int                   `firestore:"contractor_tier,omitempty"`

	// Category
	Category             *CategoryInfo          `firestore:"category,omitempty"`

	// Timestamps
	PostedOn             *time.Time             `firestore:"posted_on,omitempty"`
	CreatedOn            *time.Time             `firestore:"created_on,omitempty"`
	PublishTime          *time.Time             `firestore:"publish_time,omitempty"`

	// Budget
	Budget               *BudgetInfo            `firestore:"budget,omitempty"`
	HourlyInfo           *HourlyBudget          `firestore:"hourly_budget,omitempty"`
	WeeklyRetainerBudget *BudgetInfo            `firestore:"weekly_retainer_budget,omitempty"`
	HideBudget           *bool                  `firestore:"hide_budget,omitempty"`

	// Buyer/Client
	Buyer                *BuyerInfo             `firestore:"buyer,omitempty"`
	ClientActivity       *ClientActivity        `firestore:"client_activity,omitempty"`

	// Location
	Location             *JobLocation           `firestore:"location,omitempty"`

	// Skills and Tags
	Skills               []string               `firestore:"skills,omitempty"`
	Tags                 []string               `firestore:"tags,omitempty"`
	Occupations          []string               `firestore:"occupations,omitempty"`

	// Job Details
	URL                  string                 `firestore:"url,omitempty"`
	Ciphertext           string                 `firestore:"ciphertext,omitempty"`
	DurationLabel        string                 `firestore:"duration_label,omitempty"`
	Engagement           string                 `firestore:"engagement,omitempty"`
	Workload             string                 `firestore:"workload,omitempty"`

	// Flags
	IsPrivate            bool                   `firestore:"is_private,omitempty"`
	PrivacyReason        string                 `firestore:"privacy_reason,omitempty"`
	IsContractToHire     *bool                  `firestore:"is_contract_to_hire,omitempty"`
	WasRenewed           *bool                  `firestore:"was_renewed,omitempty"`
	Premium              *bool                  `firestore:"premium,omitempty"`

	// Additional Info
	NumberOfPositions    *int                   `firestore:"number_of_positions,omitempty"`
	ProposalsTier        string                 `firestore:"proposals_tier,omitempty"`
	TierText             string                 `firestore:"tier_text,omitempty"`
	Qualifications       *JobQualifications     `firestore:"qualifications,omitempty"`
	Recno                *int64                 `firestore:"recno,omitempty"`

	// Metadata
	LastVisitedAt        *time.Time             `firestore:"last_visited_at,omitempty"`
	ScrapeMetadata       map[string]interface{} `firestore:"scrape_metadata,omitempty"`

	// Normalized flag
	IsNormalized         bool                   `firestore:"is_normalized"`
	NormalizedAt         time.Time              `firestore:"normalized_at"`
}

// Mirror types from server package
type CategoryInfo struct {
	Name      string `firestore:"name,omitempty"`
	Slug      string `firestore:"slug,omitempty"`
	Group     string `firestore:"group,omitempty"`
	GroupSlug string `firestore:"group_slug,omitempty"`
}

type BudgetInfo struct {
	FixedAmount *float64 `firestore:"fixed_amount,omitempty"`
	Currency    string   `firestore:"currency,omitempty"`
}

type HourlyBudget struct {
	Min      *float64 `firestore:"min,omitempty"`
	Max      *float64 `firestore:"max,omitempty"`
	Currency string   `firestore:"currency,omitempty"`
}

type BuyerInfo struct {
	PaymentVerified    *bool      `firestore:"payment_verified,omitempty"`
	Country            string     `firestore:"country,omitempty"`
	City               string     `firestore:"city,omitempty"`
	Timezone           string     `firestore:"timezone,omitempty"`
	TotalSpent         *float64   `firestore:"total_spent,omitempty"`
	TotalAssignments   *int       `firestore:"total_assignments,omitempty"`
	TotalJobsWithHires *int       `firestore:"total_jobs_with_hires,omitempty"`
	ActiveAssignments  *int       `firestore:"active_assignments,omitempty"`
	FeedbackCount      *int       `firestore:"feedback_count,omitempty"`
	TotalHours         *float64   `firestore:"total_hours,omitempty"`
	Score              *float64   `firestore:"score,omitempty"`
	CompanyIndustry    string     `firestore:"company_industry,omitempty"`
	CompanySize        *int       `firestore:"company_size,omitempty"`
	ContractDate       *time.Time `firestore:"contract_date,omitempty"`
	OpenJobsCount      *int       `firestore:"open_jobs_count,omitempty"`
}

type ClientActivity struct {
	TotalApplicants         *int   `firestore:"total_applicants,omitempty"`
	TotalHired              *int   `firestore:"total_hired,omitempty"`
	TotalInvitedToInterview *int   `firestore:"total_invited_to_interview,omitempty"`
	UnansweredInvites       *int   `firestore:"unanswered_invites,omitempty"`
	InvitationsSent         *int   `firestore:"invitations_sent,omitempty"`
	LastBuyerActivity       string `firestore:"last_buyer_activity,omitempty"`
}

type JobLocation struct {
	Country  string `firestore:"country,omitempty"`
	City     string `firestore:"city,omitempty"`
	Timezone string `firestore:"timezone,omitempty"`
}

type JobQualifications struct {
	MinJobSuccessScore  *int     `firestore:"min_job_success_score,omitempty"`
	MinOdeskHours       *int     `firestore:"min_odesk_hours,omitempty"`
	PrefEnglishSkill    *int     `firestore:"pref_english_skill,omitempty"`
	RisingTalent        *bool    `firestore:"rising_talent,omitempty"`
	ShouldHavePortfolio *bool    `firestore:"should_have_portfolio,omitempty"`
	MinHoursWeek        *float64 `firestore:"min_hours_week,omitempty"`
}

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

	dryRun := flag.Bool("dry-run", true, "Run in dry-run mode (no deletions)")
	deleteEmpty := flag.Bool("delete-empty", false, "Delete empty/invalid documents")
	verbose := flag.Bool("verbose", false, "Show detailed logs for each document")
	saveReport := flag.Bool("save-report", true, "Save detailed report to file")
	batchSize := flag.Int("batch-size", 500, "Batch size for processing")
	flag.Parse()

	fmt.Println("🔍 Firestore Data Normalization Tool")
	fmt.Println("=====================================")
	fmt.Printf("Collection: %s\n", collection)
	fmt.Printf("Dry Run: %v\n", *dryRun)
	fmt.Printf("Delete Empty: %v\n", *deleteEmpty)
	fmt.Printf("Batch Size: %d\n", *batchSize)
	fmt.Println()

	ctx := context.Background()

	client, err := firestore.NewClient(ctx, projectID, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer client.Close()

	stats := &NormalizationStats{}
	var emptyDocs []string
	var invalidDocs []map[string]interface{}

	// Process documents
	iter := client.Collection(collection).Documents(ctx)
	defer iter.Stop()

	fmt.Println("📊 Analyzing and normalizing documents...")
	fmt.Println("⏱️  Adding delays to avoid Firestore quota limits...")

	updateBatch := 0
	readBatch := 0
	consecutiveErrors := 0
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error fetching document: %v", err)
			stats.ErrorDocs++
			consecutiveErrors++
			// Exponential backoff on errors
			backoffTime := time.Duration(consecutiveErrors) * 500 * time.Millisecond
			if backoffTime > 10*time.Second {
				backoffTime = 10 * time.Second
			}
			fmt.Printf("  ⏳ Backing off for %v due to errors...\n", backoffTime)
			time.Sleep(backoffTime)
			continue
		}

		consecutiveErrors = 0 // Reset on success
		readBatch++

		// Add delay every 100 reads to respect read quotas
		if readBatch%100 == 0 {
			time.Sleep(1 * time.Second)
		}

		stats.TotalDocs++
		docID := doc.Ref.ID
		rawData := doc.Data()

		// Check if document is completely empty
		if rawData == nil || len(rawData) == 0 {
			stats.EmptyDocs++
			emptyDocs = append(emptyDocs, docID)
			if *verbose {
				fmt.Printf("  ❌ Empty: %s\n", docID)
			}
			continue
		}

		// Try to transform the document using server's logic
		records, err := server.DebugTransformDocument(doc)
		if err != nil || len(records) == 0 {
			stats.InvalidDocs++
			invalidDoc := map[string]interface{}{
				"id":    docID,
				"error": err.Error(),
				"has_state": rawData["state"] != nil,
				"has_job": rawData["job"] != nil,
				"keys": getKeys(rawData),
			}
			invalidDocs = append(invalidDocs, invalidDoc)
			if *verbose {
				fmt.Printf("  ⚠️  Invalid: %s - %v\n", docID, err)
			}
			continue
		}

		// Validate that we got at least one valid job record
		var validRecord *server.JobRecord
		for _, rec := range records {
			if rec.ID != "" && rec.Title != "" {
				validRecord = &rec
				break
			}
		}

		if validRecord == nil {
			stats.InvalidDocs++
			invalidDoc := map[string]interface{}{
				"id":    docID,
				"error": "No valid job records found",
				"records_count": len(records),
			}
			invalidDocs = append(invalidDocs, invalidDoc)
			if *verbose {
				fmt.Printf("  ⚠️  No valid records: %s\n", docID)
			}
			continue
		}

		stats.ValidDocs++

		// Convert to normalized structure
		normalized := convertToNormalized(validRecord, rawData)

		// Update the document with normalized data
		if !*dryRun {
			if err := updateDocument(ctx, client, collection, docID, normalized); err != nil {
				log.Printf("Failed to update document %s: %v", docID, err)
				stats.ErrorDocs++
			} else {
				stats.NormalizedDocs++
				stats.UpdatedDocs++
				updateBatch++
				if *verbose {
					fmt.Printf("  ✅ Normalized: %s\n", docID)
				}

				// Add delay every 50 updates to respect write quotas
				if updateBatch%50 == 0 {
					time.Sleep(1 * time.Second)
				}
			}
		} else {
			stats.NormalizedDocs++
			if *verbose {
				fmt.Printf("  ✅ Valid: %s (would normalize)\n", docID)
			}
		}

		// Progress indicator
		if stats.TotalDocs%100 == 0 {
			fmt.Printf("  Processed %d documents...\n", stats.TotalDocs)
		}
	}

	// Delete empty/invalid documents if requested
	if *deleteEmpty && !*dryRun {
		fmt.Println("\n🗑️  Deleting empty/invalid documents...")

		// Delete empty documents
		for _, docID := range emptyDocs {
			_, err := client.Collection(collection).Doc(docID).Delete(ctx)
			if err != nil {
				log.Printf("Failed to delete empty doc %s: %v", docID, err)
			} else {
				stats.DeletedDocs++
				if *verbose {
					fmt.Printf("  Deleted empty: %s\n", docID)
				}
			}
		}

		// Delete invalid documents
		for _, inv := range invalidDocs {
			docID := inv["id"].(string)
			_, err := client.Collection(collection).Doc(docID).Delete(ctx)
			if err != nil {
				log.Printf("Failed to delete invalid doc %s: %v", docID, err)
			} else {
				stats.DeletedDocs++
				if *verbose {
					fmt.Printf("  Deleted invalid: %s\n", docID)
				}
			}
		}
	}

	// Print summary
	fmt.Println("\n📈 Summary")
	fmt.Println("==========")
	fmt.Printf("Total Documents:      %d\n", stats.TotalDocs)
	fmt.Printf("Valid Documents:      %d (%.1f%%)\n", stats.ValidDocs, percentage(stats.ValidDocs, stats.TotalDocs))
	fmt.Printf("Normalized Documents: %d (%.1f%%)\n", stats.NormalizedDocs, percentage(stats.NormalizedDocs, stats.TotalDocs))
	if !*dryRun {
		fmt.Printf("Updated Documents:    %d\n", stats.UpdatedDocs)
	}
	fmt.Printf("Empty Documents:      %d (%.1f%%)\n", stats.EmptyDocs, percentage(stats.EmptyDocs, stats.TotalDocs))
	fmt.Printf("Invalid Documents:    %d (%.1f%%)\n", stats.InvalidDocs, percentage(stats.InvalidDocs, stats.TotalDocs))
	if *deleteEmpty && !*dryRun {
		fmt.Printf("Deleted Documents:    %d\n", stats.DeletedDocs)
	}
	fmt.Println()

	// Show sample of issues
	if len(emptyDocs) > 0 {
		fmt.Printf("\n📝 Sample Empty Documents (showing up to 10):\n")
		for i, docID := range emptyDocs {
			if i >= 10 {
				fmt.Printf("  ... and %d more\n", len(emptyDocs)-10)
				break
			}
			fmt.Printf("  - %s\n", docID)
		}
	}

	if len(invalidDocs) > 0 {
		fmt.Printf("\n⚠️  Sample Invalid Documents (showing up to 10):\n")
		for i, inv := range invalidDocs {
			if i >= 10 {
				fmt.Printf("  ... and %d more\n", len(invalidDocs)-10)
				break
			}
			fmt.Printf("  - %s: %v\n", inv["id"], inv["error"])
			if keys, ok := inv["keys"].([]string); ok && len(keys) > 0 {
				fmt.Printf("    Keys: %s\n", strings.Join(keys, ", "))
			}
		}
	}

	// Save detailed report
	if *saveReport {
		reportPath := fmt.Sprintf("normalization_report_%s.json", time.Now().Format("20060102_150405"))
		report := map[string]interface{}{
			"timestamp":       time.Now().UTC().Format(time.RFC3339),
			"collection":      collection,
			"stats":           stats,
			"empty_docs":      emptyDocs,
			"invalid_docs":    invalidDocs,
			"dry_run":         *dryRun,
			"delete_empty":    *deleteEmpty,
		}

		reportJSON, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			log.Printf("Failed to marshal report: %v", err)
		} else {
			if err := os.WriteFile(reportPath, reportJSON, 0644); err != nil {
				log.Printf("Failed to write report: %v", err)
			} else {
				fmt.Printf("\n💾 Detailed report saved to: %s\n", reportPath)
			}
		}
	}

	// Recommendations
	fmt.Println("\n💡 Recommendations:")
	if stats.EmptyDocs > 0 || stats.InvalidDocs > 0 {
		fmt.Println("  1. Review the invalid documents to understand why they're failing")
		if *dryRun {
			fmt.Println("  2. Run with -dry-run=false -delete-empty=true to remove empty/invalid docs")
		}
		fmt.Println("  3. Check the scraper to ensure it's saving complete job data")
		fmt.Println("  4. Consider adding validation in the scraper before saving to Firestore")
	} else {
		fmt.Println("  ✅ All documents are valid! No cleanup needed.")
	}

	// Exit code based on results
	if stats.ErrorDocs > 0 {
		os.Exit(1)
	}
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func percentage(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

// convertToNormalized converts a JobRecord to NormalizedJobData
func convertToNormalized(rec *server.JobRecord, rawData map[string]interface{}) *NormalizedJobData {
	normalized := &NormalizedJobData{
		ID:                   rec.ID,
		Title:                rec.Title,
		Description:          rec.Description,
		JobType:              rec.JobType,
		Status:               rec.Status,
		ContractorTier:       rec.ContractorTier,
		PostedOn:             rec.PostedOn,
		CreatedOn:            rec.CreatedOn,
		PublishTime:          rec.PublishTime,
		URL:                  rec.URL,
		Ciphertext:           rec.Ciphertext,
		DurationLabel:        rec.DurationLabel,
		Engagement:           rec.Engagement,
		Workload:             rec.Workload,
		IsPrivate:            rec.IsPrivate,
		PrivacyReason:        rec.PrivacyReason,
		IsContractToHire:     rec.IsContractToHire,
		WasRenewed:           rec.WasRenewed,
		Premium:              rec.Premium,
		HideBudget:           rec.HideBudget,
		NumberOfPositions:    rec.NumberOfPositions,
		ProposalsTier:        rec.ProposalsTier,
		TierText:             rec.TierText,
		Recno:                rec.Recno,
		LastVisitedAt:        rec.LastVisitedAt,
		Skills:               rec.Skills,
		Tags:                 rec.Tags,
		Occupations:          rec.Occupations,
		IsNormalized:         true,
		NormalizedAt:         time.Now().UTC(),
	}

	// Convert category
	if rec.Category != nil {
		normalized.Category = &CategoryInfo{
			Name:      rec.Category.Name,
			Slug:      rec.Category.Slug,
			Group:     rec.Category.Group,
			GroupSlug: rec.Category.GroupSlug,
		}
	}

	// Convert budget
	if rec.Budget != nil {
		normalized.Budget = &BudgetInfo{
			FixedAmount: rec.Budget.FixedAmount,
			Currency:    rec.Budget.Currency,
		}
	}

	// Convert hourly budget
	if rec.HourlyInfo != nil {
		normalized.HourlyInfo = &HourlyBudget{
			Min:      rec.HourlyInfo.Min,
			Max:      rec.HourlyInfo.Max,
			Currency: rec.HourlyInfo.Currency,
		}
	}

	// Convert weekly retainer budget
	if rec.WeeklyRetainerBudget != nil {
		normalized.WeeklyRetainerBudget = &BudgetInfo{
			FixedAmount: rec.WeeklyRetainerBudget.FixedAmount,
			Currency:    rec.WeeklyRetainerBudget.Currency,
		}
	}

	// Convert buyer
	if rec.Buyer != nil {
		normalized.Buyer = &BuyerInfo{
			PaymentVerified:    rec.Buyer.PaymentVerified,
			Country:            rec.Buyer.Country,
			City:               rec.Buyer.City,
			Timezone:           rec.Buyer.Timezone,
			TotalSpent:         rec.Buyer.TotalSpent,
			TotalAssignments:   rec.Buyer.TotalAssignments,
			TotalJobsWithHires: rec.Buyer.TotalJobsWithHires,
			ActiveAssignments:  rec.Buyer.ActiveAssignments,
			FeedbackCount:      rec.Buyer.FeedbackCount,
			TotalHours:         rec.Buyer.TotalHours,
			Score:              rec.Buyer.Score,
			CompanyIndustry:    rec.Buyer.CompanyIndustry,
			CompanySize:        rec.Buyer.CompanySize,
			ContractDate:       rec.Buyer.ContractDate,
			OpenJobsCount:      rec.Buyer.OpenJobsCount,
		}
	}

	// Convert client activity
	if rec.ClientActivity != nil {
		normalized.ClientActivity = &ClientActivity{
			TotalApplicants:         rec.ClientActivity.TotalApplicants,
			TotalHired:              rec.ClientActivity.TotalHired,
			TotalInvitedToInterview: rec.ClientActivity.TotalInvitedToInterview,
			UnansweredInvites:       rec.ClientActivity.UnansweredInvites,
			InvitationsSent:         rec.ClientActivity.InvitationsSent,
			LastBuyerActivity:       rec.ClientActivity.LastBuyerActivity,
		}
	}

	// Convert location
	if rec.Location != nil {
		normalized.Location = &JobLocation{
			Country:  rec.Location.Country,
			City:     rec.Location.City,
			Timezone: rec.Location.Timezone,
		}
	}

	// Convert qualifications
	if rec.Qualifications != nil {
		normalized.Qualifications = &JobQualifications{
			MinJobSuccessScore:  rec.Qualifications.MinJobSuccessScore,
			MinOdeskHours:       rec.Qualifications.MinOdeskHours,
			PrefEnglishSkill:    rec.Qualifications.PrefEnglishSkill,
			RisingTalent:        rec.Qualifications.RisingTalent,
			ShouldHavePortfolio: rec.Qualifications.ShouldHavePortfolio,
			MinHoursWeek:        rec.Qualifications.MinHoursWeek,
		}
	}

	// Preserve scrape metadata if it exists
	if meta, ok := rawData["scrape_metadata"].(map[string]interface{}); ok {
		normalized.ScrapeMetadata = meta
	}

	return normalized
}

// updateDocument replaces the document with normalized data
func updateDocument(ctx context.Context, client *firestore.Client, collection, docID string, normalized *NormalizedJobData) error {
	// Replace the entire document with normalized structure
	_, err := client.Collection(collection).Doc(docID).Set(ctx, normalized)
	return err
}
