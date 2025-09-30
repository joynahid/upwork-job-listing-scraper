package server

import (
	"strconv"
	"strings"
	"time"
)

type sortField string

const (
	SortLastVisited sortField = "last_visited"
	SortPublishTime sortField = "publish_time"
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
	ID                    string
	Title                 string
	Description           string
	JobType               *int
	Status                *int
	ContractorTier        *int
	Category              *CategoryInfo
	PostedOn              *time.Time
	CreatedOn             *time.Time
	PublishTime           *time.Time
	Budget                *BudgetInfo
	Buyer                 *BuyerInfo
	Tags                  []string
	URL                   string
	LastVisitedAt         *time.Time
	DurationLabel         string
	Engagement            string
	Skills                []string
	HourlyInfo            *HourlyBudget
	ClientActivity        *ClientActivity
	Location              *JobLocation
	IsPrivate             bool
	PrivacyReason         string
	Ciphertext            string
	Workload              string
	IsContractToHire      *bool
	NumberOfPositions     *int
	WasRenewed            *bool
	Premium               *bool
	HideBudget            *bool
	ProposalsTier         string
	TierText              string
	Qualifications        *JobQualifications
	WeeklyRetainerBudget  *BudgetInfo
	Occupations           []string
	Recno                 *int64
}

type JobSummaryRecord struct {
	ID                   string
	Title                string
	Description          string
	JobType              *int
	DurationLabel        string
	Engagement           string
	Skills               []string
	HourlyInfo           *HourlyBudget
	FixedBudget          *BudgetInfo
	WeeklyBudget         *BudgetInfo
	Client               *JobSummaryClient
	Ciphertext           string
	URL                  string
	PublishedOn          *time.Time
	RenewedOn            *time.Time
	LastVisitedAt        *time.Time
	Workload             string
	IsContractToHire     *bool
	NumberOfPositions    *int
	WasRenewed           *bool
	Premium              *bool
	HideBudget           *bool
	ProposalsTier        string
	Qualifications       *JobQualifications
	WeeklyRetainerBudget *BudgetInfo
	Occupations          []string
	Recno                *int64
}

// JobDTO is the API response schema.
type JobDTO struct {
	ID                   string             `json:"id"`
	Title                string             `json:"title,omitempty"`
	Description          string             `json:"description,omitempty"`
	JobType              string             `json:"job_type,omitempty"`
	Status               string             `json:"status,omitempty"`
	ContractorTier       string             `json:"contractor_tier,omitempty"`
	PostedOn             string             `json:"posted_on,omitempty"`
	CreatedOn            string             `json:"created_on,omitempty"`
	PublishTime          string             `json:"publish_time,omitempty"`
	PublishTimeRelative  string             `json:"publish_time_relative,omitempty"`
	Category             *CategoryInfo      `json:"category,omitempty"`
	Budget               *BudgetInfo        `json:"budget,omitempty"`
	Buyer                *BuyerDTO          `json:"buyer,omitempty"`
	Tags                 []string           `json:"tags,omitempty"`
	URL                  string             `json:"url,omitempty"`
	LastVisitedAt        string             `json:"last_visited_at,omitempty"`
	DurationLabel        string             `json:"duration_label,omitempty"`
	Engagement           string             `json:"engagement,omitempty"`
	Skills               []string           `json:"skills,omitempty"`
	HourlyInfo           *HourlyBudget      `json:"hourly_budget,omitempty"`
	ClientActivity       *ClientActivity    `json:"client_activity,omitempty"`
	Location             *JobLocation       `json:"location,omitempty"`
	IsPrivate            bool               `json:"is_private,omitempty"`
	PrivacyReason        string             `json:"privacy_reason,omitempty"`
	Ciphertext           string             `json:"ciphertext,omitempty"`
	Workload             string             `json:"workload,omitempty"`
	IsContractToHire     *bool              `json:"is_contract_to_hire,omitempty"`
	NumberOfPositions    *int               `json:"number_of_positions,omitempty"`
	WasRenewed           *bool              `json:"was_renewed,omitempty"`
	Premium              *bool              `json:"premium,omitempty"`
	HideBudget           *bool              `json:"hide_budget,omitempty"`
	ProposalsTier        string             `json:"proposals_tier,omitempty"`
	TierText             string             `json:"tier_text,omitempty"`
	Qualifications       *JobQualifications `json:"qualifications,omitempty"`
	WeeklyRetainerBudget *BudgetInfo        `json:"weekly_retainer_budget,omitempty"`
	Occupations          []string           `json:"occupations,omitempty"`
	Recno                *int64             `json:"recno,omitempty"`
}

type JobSummaryDTO struct {
	ID                   string             `json:"id"`
	Title                string             `json:"title,omitempty"`
	Description          string             `json:"description,omitempty"`
	JobType              string             `json:"job_type,omitempty"`
	DurationLabel        string             `json:"duration_label,omitempty"`
	Engagement           string             `json:"engagement,omitempty"`
	Skills               []string           `json:"skills,omitempty"`
	HourlyInfo           *HourlyBudget      `json:"hourly_budget,omitempty"`
	FixedBudget          *BudgetInfo        `json:"fixed_budget,omitempty"`
	WeeklyBudget         *BudgetInfo        `json:"weekly_budget,omitempty"`
	Client               *JobSummaryClient  `json:"client,omitempty"`
	Ciphertext           string             `json:"ciphertext,omitempty"`
	URL                  string             `json:"url,omitempty"`
	PublishedOn          string             `json:"published_on,omitempty"`
	PublishTimeRelative  string             `json:"publish_time_relative,omitempty"`
	RenewedOn            string             `json:"renewed_on,omitempty"`
	LastVisitedAt        string             `json:"last_visited_at,omitempty"`
	Workload             string             `json:"workload,omitempty"`
	IsContractToHire     *bool              `json:"is_contract_to_hire,omitempty"`
	NumberOfPositions    *int               `json:"number_of_positions,omitempty"`
	WasRenewed           *bool              `json:"was_renewed,omitempty"`
	Premium              *bool              `json:"premium,omitempty"`
	HideBudget           *bool              `json:"hide_budget,omitempty"`
	ProposalsTier        string             `json:"proposals_tier,omitempty"`
	Qualifications       *JobQualifications `json:"qualifications,omitempty"`
	WeeklyRetainerBudget *BudgetInfo        `json:"weekly_retainer_budget,omitempty"`
	Occupations          []string           `json:"occupations,omitempty"`
	Recno                *int64             `json:"recno,omitempty"`
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
	PaymentVerified      *bool
	Country              string
	City                 string
	Timezone             string
	TotalSpent           *float64
	TotalAssignments     *int
	TotalJobsWithHires   *int
	ActiveAssignments    *int
	FeedbackCount        *int
	TotalHours           *float64
	Score                *float64
	CompanyIndustry      string
	CompanySize          *int
	ContractDate         *time.Time
	OpenJobsCount        *int
}

// BuyerDTO is the API response version of BuyerInfo
type BuyerDTO struct {
	PaymentVerified      *bool    `json:"payment_verified,omitempty"`
	Country              string   `json:"country,omitempty"`
	City                 string   `json:"city,omitempty"`
	Timezone             string   `json:"timezone,omitempty"`
	TotalSpent           *float64 `json:"total_spent,omitempty"`
	TotalAssignments     *int     `json:"total_assignments,omitempty"`
	TotalJobsWithHires   *int     `json:"total_jobs_with_hires,omitempty"`
	ActiveAssignments    *int     `json:"active_assignments,omitempty"`
	FeedbackCount        *int     `json:"feedback_count,omitempty"`
	TotalHours           *float64 `json:"total_hours,omitempty"`
	Score                *float64 `json:"score,omitempty"`
	CompanyIndustry      string   `json:"company_industry,omitempty"`
	CompanySize          *int     `json:"company_size,omitempty"`
	ContractDate         string   `json:"contract_date,omitempty"`
	OpenJobsCount        *int     `json:"open_jobs_count,omitempty"`
}

// ToDTO converts BuyerInfo to BuyerDTO
func (b *BuyerInfo) ToDTO() *BuyerDTO {
	if b == nil {
		return nil
	}

	dto := &BuyerDTO{
		PaymentVerified:    b.PaymentVerified,
		Country:            b.Country,
		City:               b.City,
		Timezone:           b.Timezone,
		TotalSpent:         b.TotalSpent,
		TotalAssignments:   b.TotalAssignments,
		TotalJobsWithHires: b.TotalJobsWithHires,
		ActiveAssignments:  b.ActiveAssignments,
		FeedbackCount:      b.FeedbackCount,
		TotalHours:         b.TotalHours,
		Score:              b.Score,
		CompanyIndustry:    b.CompanyIndustry,
		CompanySize:        b.CompanySize,
		OpenJobsCount:      b.OpenJobsCount,
	}

	if b.ContractDate != nil {
		dto.ContractDate = b.ContractDate.UTC().Format(time.RFC3339)
	}

	return dto
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

// JobQualifications represents job qualification requirements
type JobQualifications struct {
	MinJobSuccessScore  *int     `json:"min_job_success_score,omitempty"`
	MinOdeskHours       *int     `json:"min_odesk_hours,omitempty"`
	PrefEnglishSkill    *int     `json:"pref_english_skill,omitempty"`
	RisingTalent        *bool    `json:"rising_talent,omitempty"`
	ShouldHavePortfolio *bool    `json:"should_have_portfolio,omitempty"`
	MinHoursWeek        *float64 `json:"min_hours_week,omitempty"`
}

// ToDTO converts a JobRecord into response form.
func (job JobRecord) ToDTO() JobDTO {
	dto := JobDTO{
		ID:                   job.ID,
		Title:                job.Title,
		Description:          job.Description,
		JobType:              normalizeJobType(job.JobType),
		Status:               normalizeJobStatus(job.Status),
		ContractorTier:       normalizeContractorTier(job.ContractorTier),
		Category:             job.Category,
		Budget:               job.Budget,
		Buyer:                job.Buyer.ToDTO(),
		Tags:                 job.Tags,
		URL:                  job.URL,
		DurationLabel:        job.DurationLabel,
		Engagement:           job.Engagement,
		Skills:               job.Skills,
		HourlyInfo:           job.HourlyInfo,
		ClientActivity:       job.ClientActivity,
		Location:             job.Location,
		IsPrivate:            job.IsPrivate,
		PrivacyReason:        job.PrivacyReason,
		Ciphertext:           job.Ciphertext,
		Workload:             job.Workload,
		IsContractToHire:     job.IsContractToHire,
		NumberOfPositions:    job.NumberOfPositions,
		WasRenewed:           job.WasRenewed,
		Premium:              job.Premium,
		HideBudget:           job.HideBudget,
		ProposalsTier:        job.ProposalsTier,
		TierText:             job.TierText,
		Qualifications:       job.Qualifications,
		WeeklyRetainerBudget: job.WeeklyRetainerBudget,
		Occupations:          job.Occupations,
		Recno:                job.Recno,
	}

	if job.PostedOn != nil {
		dto.PostedOn = job.PostedOn.UTC().Format(time.RFC3339)
	}
	if job.CreatedOn != nil {
		dto.CreatedOn = job.CreatedOn.UTC().Format(time.RFC3339)
	}
	if job.PublishTime != nil {
		publishTime := job.PublishTime.UTC()
		dto.PublishTime = publishTime.Format(time.RFC3339)
		dto.PublishTimeRelative = formatRelativeTime(publishTime)
	}
	if job.LastVisitedAt != nil {
		dto.LastVisitedAt = job.LastVisitedAt.UTC().Format(time.RFC3339)
	}

	return dto
}

func formatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	now := time.Now().UTC()
	if t.After(now) {
		return "just now"
	}

	diff := now.Sub(t)
	seconds := int(diff.Seconds())
	minutes := int(diff.Minutes())
	hours := int(diff.Hours())
	days := int(diff.Hours() / 24)

	switch {
	case seconds < 60:
		if seconds <= 1 {
			return "just now"
		}
		return formatTimeUnit(seconds, "second")
	case minutes < 60:
		return formatTimeUnit(minutes, "minute")
	case hours < 24:
		return formatTimeUnit(hours, "hour")
	case days < 7:
		return formatTimeUnit(days, "day")
	case days < 30:
		weeks := days / 7
		return formatTimeUnit(weeks, "week")
	case days < 365:
		months := days / 30
		return formatTimeUnit(months, "month")
	default:
		years := days / 365
		return formatTimeUnit(years, "year")
	}
}

func formatTimeUnit(value int, unit string) string {
	if value == 1 {
		return "1 " + unit + " ago"
	}
	return strconv.Itoa(value) + " " + unit + "s ago"
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
		ID:                   job.ID,
		Title:                job.Title,
		Description:          job.Description,
		JobType:              normalizeJobType(job.JobType),
		DurationLabel:        job.DurationLabel,
		Engagement:           job.Engagement,
		Skills:               job.Skills,
		HourlyInfo:           job.HourlyInfo,
		FixedBudget:          job.FixedBudget,
		WeeklyBudget:         job.WeeklyBudget,
		Client:               job.Client,
		Ciphertext:           job.Ciphertext,
		URL:                  job.URL,
		Workload:             job.Workload,
		IsContractToHire:     job.IsContractToHire,
		NumberOfPositions:    job.NumberOfPositions,
		WasRenewed:           job.WasRenewed,
		Premium:              job.Premium,
		HideBudget:           job.HideBudget,
		ProposalsTier:        job.ProposalsTier,
		Qualifications:       job.Qualifications,
		WeeklyRetainerBudget: job.WeeklyRetainerBudget,
		Occupations:          job.Occupations,
		Recno:                job.Recno,
	}

	if job.PublishedOn != nil {
		publishedOn := job.PublishedOn.UTC()
		dto.PublishedOn = publishedOn.Format(time.RFC3339)
		dto.PublishTimeRelative = formatRelativeTime(publishedOn)
	}
	if job.RenewedOn != nil {
		dto.RenewedOn = job.RenewedOn.UTC().Format(time.RFC3339)
	}
	if job.LastVisitedAt != nil {
		dto.LastVisitedAt = job.LastVisitedAt.UTC().Format(time.RFC3339)
	}

	return dto
}
