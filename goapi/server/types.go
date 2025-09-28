package server

import (
	"fmt"
	"strings"
	"time"
)

type sortField string

const (
	SortLastVisited sortField = "last_visited"
	SortPostedOn    sortField = "posted_on"
	SortBudget      sortField = "budget"
)

var enumKeyReplacer = strings.NewReplacer("-", "", "_", "", " ", "")

var (
	jobTypeLabelByCode = map[int]string{
		1: "hourly",
		2: "fixed-price",
	}
	jobTypeCanonicalLabels = []string{"hourly", "fixed-price"}
	jobTypeCodeByKey       = map[string]int{
		canonicalEnumKey("hourly"):      1,
		canonicalEnumKey("hourly-job"):  1,
		canonicalEnumKey("fixed-price"): 2,
		canonicalEnumKey("fixed price"): 2,
		canonicalEnumKey("fixed"):       2,
	}

	jobStatusLabelByCode = map[int]string{
		1: "open",
		2: "closed",
	}
	jobStatusCanonicalLabels = []string{"open", "closed"}
	jobStatusCodeByKey       = map[string]int{
		canonicalEnumKey("open"):     1,
		canonicalEnumKey("opened"):   1,
		canonicalEnumKey("active"):   1,
		canonicalEnumKey("closed"):   2,
		canonicalEnumKey("inactive"): 2,
		canonicalEnumKey("archived"): 2,
	}

	contractorTierLabelByCode = map[int]string{
		1: "entry",
		2: "intermediate",
		3: "expert",
	}
	contractorTierCanonicalLabels = []string{"entry", "intermediate", "expert"}
	contractorTierCodeByKey       = map[string]int{
		canonicalEnumKey("entry"):        1,
		canonicalEnumKey("entry-level"):  1,
		canonicalEnumKey("beginner"):     1,
		canonicalEnumKey("intermediate"): 2,
		canonicalEnumKey("mid"):          2,
		canonicalEnumKey("mid-level"):    2,
		canonicalEnumKey("expert"):       3,
		canonicalEnumKey("expert-level"): 3,
		canonicalEnumKey("advanced"):     3,
	}
)

// JobRecord represents normalized job data prior to serialization.
type JobRecord struct {
	ID             string
	Title          string
	Description    string
	JobType        *int
	Status         *int
	ContractorTier *int
	Category       *CategoryInfo
	PostedOn       *time.Time
	Budget         *BudgetInfo
	Buyer          *BuyerInfo
	Tags           []string
	URL            string
	LastVisitedAt  *time.Time
	DurationLabel  string
	Engagement     string
	Skills         []string
	HourlyInfo     *HourlyBudget
	ClientActivity *ClientActivity
	Location       *JobLocation
	IsPrivate      bool
	PrivacyReason  string
}

type JobSummaryRecord struct {
	ID            string
	Title         string
	Description   string
	JobType       *int
	DurationLabel string
	Engagement    string
	Skills        []string
	HourlyInfo    *HourlyBudget
	FixedBudget   *BudgetInfo
	WeeklyBudget  *BudgetInfo
	Client        *JobSummaryClient
	Ciphertext    string
	URL           string
	PublishedOn   *time.Time
	RenewedOn     *time.Time
	LastVisitedAt *time.Time
}

// JobDTO is the API response schema.
type JobDTO struct {
	ID               string          `json:"id"`
	Title            string          `json:"title,omitempty"`
	Description      string          `json:"description,omitempty"`
	JobType          string          `json:"job_type,omitempty"`
	Status           string          `json:"status,omitempty"`
	ContractorTier   string          `json:"contractor_tier,omitempty"`
	PostedOn         string          `json:"posted_on,omitempty"`
	Category         *CategoryInfo   `json:"category,omitempty"`
	Budget           *BudgetInfo     `json:"budget,omitempty"`
	Buyer            *BuyerInfo      `json:"buyer,omitempty"`
	Tags             []string        `json:"tags,omitempty"`
	URL              string          `json:"url,omitempty"`
	LastVisitedAt    string          `json:"last_visited_at,omitempty"`
	PostedOnRelative string          `json:"posted_on_relative,omitempty"`
	DurationLabel    string          `json:"duration_label,omitempty"`
	Engagement       string          `json:"engagement,omitempty"`
	Skills           []string        `json:"skills,omitempty"`
	HourlyInfo       *HourlyBudget   `json:"hourly_budget,omitempty"`
	ClientActivity   *ClientActivity `json:"client_activity,omitempty"`
	Location         *JobLocation    `json:"location,omitempty"`
	IsPrivate        bool            `json:"is_private,omitempty"`
	PrivacyReason    string          `json:"privacy_reason,omitempty"`
}

type JobSummaryDTO struct {
	ID            string            `json:"id"`
	Title         string            `json:"title,omitempty"`
	Description   string            `json:"description,omitempty"`
	JobType       string            `json:"job_type,omitempty"`
	DurationLabel string            `json:"duration_label,omitempty"`
	Engagement    string            `json:"engagement,omitempty"`
	Skills        []string          `json:"skills,omitempty"`
	HourlyInfo    *HourlyBudget     `json:"hourly_budget,omitempty"`
	FixedBudget   *BudgetInfo       `json:"fixed_budget,omitempty"`
	WeeklyBudget  *BudgetInfo       `json:"weekly_budget,omitempty"`
	Client        *JobSummaryClient `json:"client,omitempty"`
	Ciphertext    string            `json:"ciphertext,omitempty"`
	URL           string            `json:"url,omitempty"`
	PublishedOn   string            `json:"published_on,omitempty"`
	RenewedOn     string            `json:"renewed_on,omitempty"`
	LastVisitedAt string            `json:"last_visited_at,omitempty"`
}

// BudgetInfo describes job budget metadata.
type BudgetInfo struct {
	FixedAmount *float64 `json:"fixed_amount,omitempty"`
	Currency    string   `json:"currency,omitempty"`
}

type HourlyBudget struct {
	Min      *float64 `json:"min,omitempty"`
	Max      *float64 `json:"max,omitempty"`
	Currency string   `json:"currency,omitempty"`
}

// JobsResponse is the envelope returned by /jobs and /health endpoints.
type JobsResponse struct {
	Success     bool     `json:"success"`
	Data        []JobDTO `json:"data"`
	Count       int      `json:"count"`
	LastUpdated string   `json:"last_updated"`
	Message     string   `json:"message,omitempty"`
}

type JobListResponse struct {
	Success     bool            `json:"success"`
	Data        []JobSummaryDTO `json:"data"`
	Count       int             `json:"count"`
	LastUpdated string          `json:"last_updated"`
	Message     string          `json:"message,omitempty"`
}

// CategoryInfo provides category context.
type CategoryInfo struct {
	Name      string `json:"name,omitempty"`
	Slug      string `json:"slug,omitempty"`
	Group     string `json:"group,omitempty"`
	GroupSlug string `json:"group_slug,omitempty"`
}

// BuyerInfo captures client/company details.
type BuyerInfo struct {
	PaymentVerified    *bool    `json:"payment_verified,omitempty"`
	Country            string   `json:"country,omitempty"`
	City               string   `json:"city,omitempty"`
	Timezone           string   `json:"timezone,omitempty"`
	TotalSpent         *float64 `json:"total_spent,omitempty"`
	TotalAssignments   *int     `json:"total_assignments,omitempty"`
	TotalJobsWithHires *int     `json:"total_jobs_with_hires,omitempty"`
}

type ClientActivity struct {
	TotalApplicants         *int   `json:"total_applicants,omitempty"`
	TotalHired              *int   `json:"total_hired,omitempty"`
	TotalInvitedToInterview *int   `json:"total_invited_to_interview,omitempty"`
	UnansweredInvites       *int   `json:"unanswered_invites,omitempty"`
	InvitationsSent         *int   `json:"invitations_sent,omitempty"`
	LastBuyerActivity       string `json:"last_buyer_activity,omitempty"`
}

type JobLocation struct {
	Country  string `json:"country,omitempty"`
	City     string `json:"city,omitempty"`
	Timezone string `json:"timezone,omitempty"`
}

type JobSummaryClient struct {
	PaymentVerified *bool  `json:"payment_verified,omitempty"`
	Country         string `json:"country,omitempty"`
}

// ToDTO converts a JobRecord into response form.
func (job JobRecord) ToDTO() JobDTO {
	dto := JobDTO{
		ID:             job.ID,
		Title:          job.Title,
		Description:    job.Description,
		JobType:        normalizeJobType(job.JobType),
		Status:         normalizeJobStatus(job.Status),
		ContractorTier: normalizeContractorTier(job.ContractorTier),
		Category:       job.Category,
		Budget:         job.Budget,
		Buyer:          job.Buyer,
		Tags:           job.Tags,
		URL:            job.URL,
		DurationLabel:  job.DurationLabel,
		Engagement:     job.Engagement,
		Skills:         job.Skills,
		HourlyInfo:     job.HourlyInfo,
		ClientActivity: job.ClientActivity,
		Location:       job.Location,
		IsPrivate:      job.IsPrivate,
		PrivacyReason:  job.PrivacyReason,
	}

	if job.PostedOn != nil {
		posted := job.PostedOn.UTC()
		dto.PostedOn = posted.Format(time.RFC3339)
		dto.PostedOnRelative = formatRelativeTime(posted, time.Now().UTC())
	}
	if job.LastVisitedAt != nil {
		dto.LastVisitedAt = job.LastVisitedAt.UTC().Format(time.RFC3339)
	}

	return dto
}

func formatRelativeTime(target time.Time, reference time.Time) string {
	if target.IsZero() {
		return ""
	}

	if reference.IsZero() {
		reference = time.Now().UTC()
	}

	if target.After(reference) {
		return "just now"
	}

	diff := reference.Sub(target)

	seconds := int(diff.Seconds())
	if seconds < 1 {
		return "just now"
	}

	minutes := int(diff.Minutes())
	hours := int(diff.Hours())
	days := int(diff.Hours() / 24)
	weeks := days / 7
	months := days / 30

	switch {
	case seconds < 60:
		return relativeLabel(seconds, "second")
	case minutes < 60:
		return relativeLabel(minutes, "minute")
	case hours < 24:
		return relativeLabel(hours, "hour")
	case days < 7:
		return relativeLabel(days, "day")
	case weeks < 5:
		if weeks < 1 {
			weeks = 1
		}
		return relativeLabel(weeks, "week")
	case months >= 1:
		if months < 1 {
			months = 1
		}
		return relativeLabel(months, "month")
	default:
		return relativeLabel(days, "day")
	}
}

func relativeLabel(value int, unit string) string {
	if value < 1 {
		value = 1
	}
	if value == 1 {
		return fmt.Sprintf("1 %s ago", unit)
	}
	return fmt.Sprintf("%d %ss ago", value, unit)
}

func canonicalEnumKey(value string) string {
	if value == "" {
		return ""
	}
	return enumKeyReplacer.Replace(strings.ToLower(strings.TrimSpace(value)))
}

func jobTypeLabelFromCode(code int) string {
	if label, ok := jobTypeLabelByCode[code]; ok {
		return label
	}
	return "unknown"
}

func jobTypeCodeFromLabel(label string) (int, bool) {
	if label == "" {
		return 0, false
	}
	code, ok := jobTypeCodeByKey[canonicalEnumKey(label)]
	return code, ok
}

func jobTypeAcceptedLabels() []string {
	return append([]string(nil), jobTypeCanonicalLabels...)
}

func jobStatusLabelFromCode(code int) string {
	if label, ok := jobStatusLabelByCode[code]; ok {
		return label
	}
	return "unknown"
}

func jobStatusCodeFromLabel(label string) (int, bool) {
	if label == "" {
		return 0, false
	}
	code, ok := jobStatusCodeByKey[canonicalEnumKey(label)]
	return code, ok
}

func jobStatusAcceptedLabels() []string {
	return append([]string(nil), jobStatusCanonicalLabels...)
}

func contractorTierLabelFromCode(code int) string {
	if label, ok := contractorTierLabelByCode[code]; ok {
		return label
	}
	return "unknown"
}

func contractorTierCodeFromLabel(label string) (int, bool) {
	if label == "" {
		return 0, false
	}
	code, ok := contractorTierCodeByKey[canonicalEnumKey(label)]
	return code, ok
}

func contractorTierAcceptedLabels() []string {
	return append([]string(nil), contractorTierCanonicalLabels...)
}

func normalizeJobType(code *int) string {
	if code == nil {
		return ""
	}

	return jobTypeLabelFromCode(*code)
}

func normalizeJobStatus(code *int) string {
	if code == nil {
		return ""
	}

	return jobStatusLabelFromCode(*code)
}

func normalizeContractorTier(code *int) string {
	if code == nil {
		return ""
	}

	return contractorTierLabelFromCode(*code)
}

// ToDTO converts a JobSummaryRecord into response form.
func (job JobSummaryRecord) ToDTO() JobSummaryDTO {
	dto := JobSummaryDTO{
		ID:            job.ID,
		Title:         job.Title,
		Description:   job.Description,
		JobType:       normalizeJobType(job.JobType),
		DurationLabel: job.DurationLabel,
		Engagement:    job.Engagement,
		Skills:        job.Skills,
		HourlyInfo:    job.HourlyInfo,
		FixedBudget:   job.FixedBudget,
		WeeklyBudget:  job.WeeklyBudget,
		Client:        job.Client,
		Ciphertext:    job.Ciphertext,
		URL:           job.URL,
	}

	if job.PublishedOn != nil {
		dto.PublishedOn = job.PublishedOn.UTC().Format(time.RFC3339)
	}
	if job.RenewedOn != nil {
		dto.RenewedOn = job.RenewedOn.UTC().Format(time.RFC3339)
	}
	if job.LastVisitedAt != nil {
		dto.LastVisitedAt = job.LastVisitedAt.UTC().Format(time.RFC3339)
	}

	return dto
}
