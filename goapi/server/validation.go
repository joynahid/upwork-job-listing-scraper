package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationError represents a detailed validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Example string `json:"example,omitempty"`
}

// ValidationErrorResponse is the structured error response
type ValidationErrorResponse struct {
	Success bool              `json:"success"`
	Error   string            `json:"error"`
	Details []ValidationError `json:"details"`
}

// JobsQueryParams defines the validated query parameters for /jobs endpoint
type JobsQueryParams struct {
	Limit                      int     `form:"limit" binding:"omitempty,min=1,max=50"`
	Offset                     int     `form:"offset" binding:"omitempty,min=0"`
	PaymentVerified            *bool   `form:"payment_verified" binding:"omitempty"`
	Category                   string  `form:"category" binding:"omitempty,max=200"`
	CategoryGroup              string  `form:"category_group" binding:"omitempty,max=200"`
	Status                     string  `form:"status" binding:"omitempty,oneof=open closed active inactive archived 1 2"`
	JobType                    string  `form:"job_type" binding:"omitempty,job_type_enum"`
	ContractorTier             string  `form:"contractor_tier" binding:"omitempty,contractor_tier_enum"`
	Country                    string  `form:"country" binding:"omitempty,iso3166_1_alpha2|iso3166_1_alpha3"`
	Tags                       string  `form:"tags" binding:"omitempty,max=500"`
	Skills                     string  `form:"skills" binding:"omitempty,max=500"`
	PostedAfter                string  `form:"posted_after" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	PostedBefore               string  `form:"posted_before" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	LastVisitedAfter           string  `form:"last_visited_after" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	BudgetMin                  float64 `form:"budget_min" binding:"omitempty,min=0"`
	BudgetMax                  float64 `form:"budget_max" binding:"omitempty,min=0,gtefield=BudgetMin"`
	HourlyMin                  float64 `form:"hourly_min" binding:"omitempty,min=0"`
	HourlyMax                  float64 `form:"hourly_max" binding:"omitempty,min=0,gtefield=HourlyMin"`
	DurationLabel              string  `form:"duration_label" binding:"omitempty,max=100"`
	Engagement                 string  `form:"engagement" binding:"omitempty,max=100"`
	BuyerTotalSpentMin         float64 `form:"buyer.total_spent_min" binding:"omitempty,min=0"`
	BuyerTotalSpentMax         float64 `form:"buyer.total_spent_max" binding:"omitempty,min=0,gtefield=BuyerTotalSpentMin"`
	BuyerTotalAssignmentsMin   int     `form:"buyer.total_assignments_min" binding:"omitempty,min=0"`
	BuyerTotalAssignmentsMax   int     `form:"buyer.total_assignments_max" binding:"omitempty,min=0,gtefield=BuyerTotalAssignmentsMin"`
	BuyerTotalJobsWithHiresMin int     `form:"buyer.total_jobs_with_hires_min" binding:"omitempty,min=0"`
	BuyerTotalJobsWithHiresMax int     `form:"buyer.total_jobs_with_hires_max" binding:"omitempty,min=0,gtefield=BuyerTotalJobsWithHiresMin"`
	Workload                   string  `form:"workload" binding:"omitempty,max=100"`
	IsContractToHire           *bool   `form:"is_contract_to_hire" binding:"omitempty"`
	NumberOfPositionsMin       int     `form:"number_of_positions_min" binding:"omitempty,min=0"`
	NumberOfPositionsMax       int     `form:"number_of_positions_max" binding:"omitempty,min=0,gtefield=NumberOfPositionsMin"`
	WasRenewed                 *bool   `form:"was_renewed" binding:"omitempty"`
	Premium                    *bool   `form:"premium" binding:"omitempty"`
	HideBudget                 *bool   `form:"hide_budget" binding:"omitempty"`
	ProposalsTier              string  `form:"proposals_tier" binding:"omitempty,max=100"`
	MinJobSuccessScore         int     `form:"min_job_success_score" binding:"omitempty,min=0,max=100"`
	MinOdeskHours              int     `form:"min_odesk_hours" binding:"omitempty,min=0"`
	PrefEnglishSkill           int     `form:"pref_english_skill" binding:"omitempty,min=0,max=4"`
	RisingTalent               *bool   `form:"rising_talent" binding:"omitempty"`
	ShouldHavePortfolio        *bool   `form:"should_have_portfolio" binding:"omitempty"`
	MinHoursWeek               float64 `form:"min_hours_week" binding:"omitempty,min=0"`
	Sort                       string  `form:"sort" binding:"omitempty,sort_field"`
}

// JobListQueryParams defines the validated query parameters for /job-list endpoint
type JobListQueryParams struct {
	Limit           int     `form:"limit" binding:"omitempty,min=1,max=50"`
	PaymentVerified *bool   `form:"payment_verified" binding:"omitempty"`
	Country         string  `form:"country" binding:"omitempty,iso3166_1_alpha2|iso3166_1_alpha3"`
	Skills          string  `form:"skills" binding:"omitempty,max=500"`
	JobType         string  `form:"job_type" binding:"omitempty,job_type_enum"`
	Duration        string  `form:"duration" binding:"omitempty,max=100"`
	HourlyMin       float64 `form:"hourly_min" binding:"omitempty,min=0"`
	HourlyMax       float64 `form:"hourly_max" binding:"omitempty,min=0,gtefield=HourlyMin"`
	BudgetMin       float64 `form:"budget_min" binding:"omitempty,min=0"`
	BudgetMax       float64 `form:"budget_max" binding:"omitempty,min=0,gtefield=BudgetMin"`
	Search          string  `form:"search" binding:"omitempty,max=500"`
	Sort            string  `form:"sort" binding:"omitempty,sort_field"`
}

// RegisterCustomValidators registers custom validators with gin's validator
func RegisterCustomValidators(v *validator.Validate) {
	v.RegisterValidation("job_type_enum", validateJobType)
	v.RegisterValidation("contractor_tier_enum", validateContractorTier)
	v.RegisterValidation("sort_field", validateSortField)
}

// validateJobType validates job type enum values
func validateJobType(fl validator.FieldLevel) bool {
	value := strings.TrimSpace(strings.ToLower(fl.Field().String()))
	if value == "" {
		return true // empty is valid for optional fields
	}

	// Accept numeric codes
	if value == "1" || value == "2" {
		return true
	}

	// Accept string labels (with variations)
	validLabels := []string{
		"hourly", "hourly-job", "hourlyjob",
		"fixed-price", "fixed price", "fixedprice", "fixed",
	}

	for _, label := range validLabels {
		if value == label {
			return true
		}
	}

	return false
}

// validateContractorTier validates contractor tier enum values
func validateContractorTier(fl validator.FieldLevel) bool {
	value := strings.TrimSpace(strings.ToLower(fl.Field().String()))
	if value == "" {
		return true // empty is valid for optional fields
	}

	// Accept numeric codes
	if value == "1" || value == "2" || value == "3" {
		return true
	}

	// Accept string labels (with variations)
	validLabels := []string{
		"entry", "entry-level", "entrylevel", "beginner",
		"intermediate", "mid", "mid-level", "midlevel",
		"expert", "expert-level", "expertlevel", "advanced",
	}

	for _, label := range validLabels {
		if value == label {
			return true
		}
	}

	return false
}

// validateSortField validates sort field enum values
func validateSortField(fl validator.FieldLevel) bool {
	value := strings.TrimSpace(strings.ToLower(fl.Field().String()))
	if value == "" {
		return true // empty is valid for optional fields
	}

	validSortFields := []string{
		"publish_time_asc", "publish_time_desc",
		"last_visited_asc", "last_visited_desc",
		"budget_asc", "budget_desc",
		"posted_on_asc", "posted_on_desc", // aliases
	}

	for _, field := range validSortFields {
		if value == field {
			return true
		}
	}

	return false
}

// FormatValidationErrors converts validator errors into LLM-friendly messages
func FormatValidationErrors(err error) ValidationErrorResponse {
	response := ValidationErrorResponse{
		Success: false,
		Error:   "Validation failed. Please check the details below and correct your request.",
		Details: []ValidationError{},
	}

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrs {
			validationErr := ValidationError{
				Field:   getFieldName(fieldErr),
				Message: formatFieldError(fieldErr),
				Example: getFieldExample(fieldErr),
			}
			response.Details = append(response.Details, validationErr)
		}
	} else {
		// Handle other types of errors
		response.Details = append(response.Details, ValidationError{
			Field:   "general",
			Message: err.Error(),
			Example: "",
		})
	}

	return response
}

// getFieldName converts the struct field name to the API parameter name
func getFieldName(fieldErr validator.FieldError) string {
	field := fieldErr.Field()

	// Convert common field names to their API parameter names
	fieldMap := map[string]string{
		"Limit":                      "limit",
		"Offset":                     "offset",
		"PaymentVerified":            "payment_verified",
		"Category":                   "category",
		"CategoryGroup":              "category_group",
		"Status":                     "status",
		"JobType":                    "job_type",
		"ContractorTier":             "contractor_tier",
		"Country":                    "country",
		"Tags":                       "tags",
		"Skills":                     "skills",
		"PostedAfter":                "posted_after",
		"PostedBefore":               "posted_before",
		"LastVisitedAfter":           "last_visited_after",
		"BudgetMin":                  "budget_min",
		"BudgetMax":                  "budget_max",
		"HourlyMin":                  "hourly_min",
		"HourlyMax":                  "hourly_max",
		"DurationLabel":              "duration_label",
		"Duration":                   "duration",
		"Engagement":                 "engagement",
		"BuyerTotalSpentMin":         "buyer.total_spent_min",
		"BuyerTotalSpentMax":         "buyer.total_spent_max",
		"BuyerTotalAssignmentsMin":   "buyer.total_assignments_min",
		"BuyerTotalAssignmentsMax":   "buyer.total_assignments_max",
		"BuyerTotalJobsWithHiresMin": "buyer.total_jobs_with_hires_min",
		"BuyerTotalJobsWithHiresMax": "buyer.total_jobs_with_hires_max",
		"Workload":                   "workload",
		"IsContractToHire":           "is_contract_to_hire",
		"NumberOfPositionsMin":       "number_of_positions_min",
		"NumberOfPositionsMax":       "number_of_positions_max",
		"WasRenewed":                 "was_renewed",
		"Premium":                    "premium",
		"HideBudget":                 "hide_budget",
		"ProposalsTier":              "proposals_tier",
		"MinJobSuccessScore":         "min_job_success_score",
		"MinOdeskHours":              "min_odesk_hours",
		"PrefEnglishSkill":           "pref_english_skill",
		"RisingTalent":               "rising_talent",
		"ShouldHavePortfolio":        "should_have_portfolio",
		"MinHoursWeek":               "min_hours_week",
		"Sort":                       "sort",
		"Search":                     "search",
	}

	if apiField, exists := fieldMap[field]; exists {
		return apiField
	}

	return strings.ToLower(field)
}

// formatFieldError creates a human-readable error message for each validation failure
func formatFieldError(fieldErr validator.FieldError) string {
	field := getFieldName(fieldErr)
	tag := fieldErr.Tag()
	param := fieldErr.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("The '%s' field is required but was not provided.", field)
	case "min":
		if fieldErr.Type().Name() == "string" {
			return fmt.Sprintf("The '%s' field must be at least %s characters long.", field, param)
		}
		return fmt.Sprintf("The '%s' field must be at least %s.", field, param)
	case "max":
		if fieldErr.Type().Name() == "string" {
			return fmt.Sprintf("The '%s' field must not exceed %s characters.", field, param)
		}
		return fmt.Sprintf("The '%s' field must not exceed %s.", field, param)
	case "gtefield":
		return fmt.Sprintf("The '%s' field must be greater than or equal to '%s'.", field, param)
	case "datetime":
		return fmt.Sprintf("The '%s' field must be a valid ISO 8601 datetime string (format: %s).", field, param)
	case "iso3166_1_alpha2":
		return fmt.Sprintf("The '%s' field must be a valid 2-letter ISO 3166-1 alpha-2 country code (e.g., 'US', 'GB', 'CA').", field)
	case "iso3166_1_alpha3":
		return fmt.Sprintf("The '%s' field must be a valid 3-letter ISO 3166-1 alpha-3 country code (e.g., 'USA', 'GBR', 'CAN').", field)
	case "oneof":
		return fmt.Sprintf("The '%s' field must be one of: %s.", field, strings.ReplaceAll(param, " ", ", "))
	case "job_type_enum":
		return fmt.Sprintf("The '%s' field must be a valid job type. Accepted values: 'hourly', 'fixed-price', or numeric codes (1=hourly, 2=fixed-price).", field)
	case "contractor_tier_enum":
		return fmt.Sprintf("The '%s' field must be a valid contractor tier. Accepted values: 'entry', 'intermediate', 'expert', or numeric codes (1=entry, 2=intermediate, 3=expert).", field)
	case "sort_field":
		return fmt.Sprintf("The '%s' field must be a valid sort field. Accepted values: 'publish_time_asc', 'publish_time_desc', 'last_visited_asc', 'last_visited_desc', 'budget_asc', 'budget_desc'.", field)
	default:
		return fmt.Sprintf("The '%s' field failed validation: %s.", field, tag)
	}
}

// getFieldExample provides an example value for the field
func getFieldExample(fieldErr validator.FieldError) string {
	field := getFieldName(fieldErr)
	tag := fieldErr.Tag()

	examples := map[string]string{
		"limit":                         "?limit=20",
		"offset":                        "?offset=0",
		"payment_verified":              "?payment_verified=true",
		"category":                      "?category=web-mobile-software-dev",
		"category_group":                "?category_group=it-networking",
		"status":                        "?status=open (or 'closed', 'active', 'inactive', 1, 2)",
		"job_type":                      "?job_type=hourly (or 'fixed-price', 1, 2)",
		"contractor_tier":               "?contractor_tier=intermediate (or 'entry', 'expert', 1, 2, 3)",
		"country":                       "?country=US (or 'USA', 'GB', 'GBR')",
		"tags":                          "?tags=python,django,api",
		"skills":                        "?skills=python,django,api",
		"posted_after":                  "?posted_after=2024-01-01T00:00:00Z",
		"posted_before":                 "?posted_before=2024-12-31T23:59:59Z",
		"last_visited_after":            "?last_visited_after=2024-01-01T00:00:00Z",
		"budget_min":                    "?budget_min=1000",
		"budget_max":                    "?budget_max=5000",
		"hourly_min":                    "?hourly_min=50",
		"hourly_max":                    "?hourly_max=150",
		"duration_label":                "?duration_label=1 to 3 months",
		"duration":                      "?duration=1 to 3 months",
		"engagement":                    "?engagement=less-than-30-hrs-week",
		"buyer.total_spent_min":         "?buyer.total_spent_min=10000",
		"buyer.total_spent_max":         "?buyer.total_spent_max=100000",
		"buyer.total_assignments_min":   "?buyer.total_assignments_min=5",
		"buyer.total_assignments_max":   "?buyer.total_assignments_max=50",
		"buyer.total_jobs_with_hires_min": "?buyer.total_jobs_with_hires_min=3",
		"buyer.total_jobs_with_hires_max": "?buyer.total_jobs_with_hires_max=20",
		"workload":                        "?workload=more-than-30-hrs-week",
		"is_contract_to_hire":             "?is_contract_to_hire=true",
		"number_of_positions_min":         "?number_of_positions_min=1",
		"number_of_positions_max":         "?number_of_positions_max=5",
		"was_renewed":                     "?was_renewed=false",
		"premium":                         "?premium=true",
		"hide_budget":                     "?hide_budget=false",
		"proposals_tier":                  "?proposals_tier=50+",
		"min_job_success_score":           "?min_job_success_score=90",
		"min_odesk_hours":                 "?min_odesk_hours=1000",
		"pref_english_skill":              "?pref_english_skill=3 (0-4: 0=Any, 1=Conversational, 2=Fluent, 3=Native/Bilingual, 4=Professional)",
		"rising_talent":                   "?rising_talent=true",
		"should_have_portfolio":           "?should_have_portfolio=true",
		"min_hours_week":                  "?min_hours_week=30",
		"sort":                            "?sort=publish_time_desc (or 'publish_time_asc', 'last_visited_asc', 'last_visited_desc', 'budget_asc', 'budget_desc')",
		"search":                          "?search=python developer",
	}

	if example, exists := examples[field]; exists {
		return example
	}

	// Provide generic examples based on tag
	switch tag {
	case "datetime":
		return fmt.Sprintf("?%s=%s", field, time.Now().UTC().Format(time.RFC3339))
	case "min", "max":
		if fieldErr.Type().Name() == "int" {
			return fmt.Sprintf("?%s=10", field)
		} else if fieldErr.Type().Name() == "float64" {
			return fmt.Sprintf("?%s=100.50", field)
		}
		return fmt.Sprintf("?%s=example_value", field)
	default:
		return fmt.Sprintf("?%s=<value>", field)
	}
}

// ValidateAndBindJobsQuery validates and binds the jobs query parameters
func ValidateAndBindJobsQuery(c *gin.Context) (*JobsQueryParams, error) {
	var params JobsQueryParams

	if err := c.ShouldBindQuery(&params); err != nil {
		return nil, err
	}

	// Additional cross-field validations
	if params.Offset > 0 && params.Limit > 0 {
		if params.Offset%params.Limit != 0 {
			return nil, fmt.Errorf("offset must be a multiple of limit (offset=%d, limit=%d). Example: ?limit=20&offset=0 or ?limit=20&offset=20", params.Offset, params.Limit)
		}
	}

	// Validate date ranges
	if params.PostedAfter != "" && params.PostedBefore != "" {
		after, err1 := time.Parse(time.RFC3339, params.PostedAfter)
		before, err2 := time.Parse(time.RFC3339, params.PostedBefore)
		if err1 == nil && err2 == nil && after.After(before) {
			return nil, fmt.Errorf("posted_after (%s) must be before posted_before (%s)", params.PostedAfter, params.PostedBefore)
		}
	}

	return &params, nil
}

// ValidateAndBindJobListQuery validates and binds the job-list query parameters
func ValidateAndBindJobListQuery(c *gin.Context) (*JobListQueryParams, error) {
	var params JobListQueryParams

	if err := c.ShouldBindQuery(&params); err != nil {
		return nil, err
	}

	return &params, nil
}

// convertToFilterOptions converts validated JobsQueryParams to internal FilterOptions
func convertToFilterOptions(params *JobsQueryParams) (FilterOptions, error) {
	opts := FilterOptions{
		Limit:         defaultLimit,
		SortField:     DefaultSortField,
		SortAscending: DefaultSortAscending,
	}

	if params.Limit > 0 {
		opts.Limit = params.Limit
	}

	opts.Offset = params.Offset
	opts.PaymentVerified = params.PaymentVerified
	opts.CategorySlug = strings.TrimSpace(params.Category)
	opts.CategoryGroupSlug = strings.TrimSpace(params.CategoryGroup)

	// Parse status
	if params.Status != "" {
		parsed, err := parseEnumFilterValue(params.Status, "status", jobStatusCodeFromLabel, jobStatusAcceptedLabels())
		if err != nil {
			return opts, err
		}
		opts.Status = parsed
	}

	// Parse job_type
	if params.JobType != "" {
		parsed, err := parseEnumFilterValue(params.JobType, "job_type", jobTypeCodeFromLabel, jobTypeAcceptedLabels())
		if err != nil {
			return opts, err
		}
		opts.JobType = parsed
	}

	// Parse contractor_tier
	if params.ContractorTier != "" {
		parsed, err := parseEnumFilterValue(params.ContractorTier, "contractor_tier", contractorTierCodeFromLabel, contractorTierAcceptedLabels())
		if err != nil {
			return opts, err
		}
		opts.ContractorTier = parsed
	}

	opts.Country = strings.TrimSpace(params.Country)

	// Parse tags and skills
	for _, token := range strings.Split(params.Tags, ",") {
		trimmed := strings.TrimSpace(token)
		if trimmed != "" {
			opts.Tags = append(opts.Tags, trimmed)
		}
	}
	for _, token := range strings.Split(params.Skills, ",") {
		trimmed := strings.TrimSpace(token)
		if trimmed != "" {
			opts.Tags = append(opts.Tags, trimmed)
		}
	}

	// Parse timestamps
	if params.PostedAfter != "" {
		t, err := time.Parse(time.RFC3339, params.PostedAfter)
		if err != nil {
			return opts, fmt.Errorf("invalid posted_after format: %w", err)
		}
		opts.PostedAfter = &t
	}

	if params.PostedBefore != "" {
		t, err := time.Parse(time.RFC3339, params.PostedBefore)
		if err != nil {
			return opts, fmt.Errorf("invalid posted_before format: %w", err)
		}
		opts.PostedBefore = &t
	}

	if params.LastVisitedAfter != "" {
		t, err := time.Parse(time.RFC3339, params.LastVisitedAfter)
		if err != nil {
			return opts, fmt.Errorf("invalid last_visited_after format: %w", err)
		}
		opts.LastVisitedAfter = &t
	}

	// Budgets
	if params.BudgetMin > 0 {
		opts.BudgetMin = &params.BudgetMin
	}
	if params.BudgetMax > 0 {
		opts.BudgetMax = &params.BudgetMax
	}
	if params.HourlyMin > 0 {
		opts.HourlyMin = &params.HourlyMin
	}
	if params.HourlyMax > 0 {
		opts.HourlyMax = &params.HourlyMax
	}

	opts.DurationLabel = strings.TrimSpace(params.DurationLabel)
	opts.Engagement = strings.TrimSpace(params.Engagement)

	// Buyer fields
	if params.BuyerTotalSpentMin > 0 {
		opts.BuyerTotalSpentMin = &params.BuyerTotalSpentMin
	}
	if params.BuyerTotalSpentMax > 0 {
		opts.BuyerTotalSpentMax = &params.BuyerTotalSpentMax
	}
	if params.BuyerTotalAssignmentsMin > 0 {
		opts.BuyerTotalAssignmentsMin = &params.BuyerTotalAssignmentsMin
	}
	if params.BuyerTotalAssignmentsMax > 0 {
		opts.BuyerTotalAssignmentsMax = &params.BuyerTotalAssignmentsMax
	}
	if params.BuyerTotalJobsWithHiresMin > 0 {
		opts.BuyerTotalJobsWithHiresMin = &params.BuyerTotalJobsWithHiresMin
	}
	if params.BuyerTotalJobsWithHiresMax > 0 {
		opts.BuyerTotalJobsWithHiresMax = &params.BuyerTotalJobsWithHiresMax
	}

	opts.Workload = strings.TrimSpace(params.Workload)
	opts.IsContractToHire = params.IsContractToHire

	if params.NumberOfPositionsMin > 0 {
		opts.NumberOfPositionsMin = &params.NumberOfPositionsMin
	}
	if params.NumberOfPositionsMax > 0 {
		opts.NumberOfPositionsMax = &params.NumberOfPositionsMax
	}

	opts.WasRenewed = params.WasRenewed
	opts.Premium = params.Premium
	opts.HideBudget = params.HideBudget
	opts.ProposalsTier = strings.TrimSpace(params.ProposalsTier)

	if params.MinJobSuccessScore > 0 {
		opts.MinJobSuccessScore = &params.MinJobSuccessScore
	}
	if params.MinOdeskHours > 0 {
		opts.MinOdeskHours = &params.MinOdeskHours
	}
	if params.PrefEnglishSkill > 0 {
		opts.PrefEnglishSkill = &params.PrefEnglishSkill
	}

	opts.RisingTalent = params.RisingTalent
	opts.ShouldHavePortfolio = params.ShouldHavePortfolio

	if params.MinHoursWeek > 0 {
		opts.MinHoursWeek = &params.MinHoursWeek
	}

	// Parse sort
	if params.Sort != "" {
		sortLower := strings.ToLower(strings.TrimSpace(params.Sort))
		switch sortLower {
		case "publish_time_asc", "posted_on_asc":
			opts.SortField = SortPublishTime
			opts.SortAscending = true
		case "publish_time_desc", "posted_on_desc":
			opts.SortField = SortPublishTime
			opts.SortAscending = false
		case "last_visited_asc":
			opts.SortField = SortLastVisited
			opts.SortAscending = true
		case "last_visited_desc":
			opts.SortField = SortLastVisited
			opts.SortAscending = false
		case "budget_asc":
			opts.SortField = SortBudget
			opts.SortAscending = true
		case "budget_desc":
			opts.SortField = SortBudget
			opts.SortAscending = false
		}
	}

	return opts, nil
}

// convertToJobListFilterOptions converts validated JobListQueryParams to internal JobListFilterOptions
func convertToJobListFilterOptions(params *JobListQueryParams) (JobListFilterOptions, error) {
	opts := JobListFilterOptions{
		Limit:         defaultLimit,
		SortField:     SortLastVisited,
		SortAscending: false,
	}

	if params.Limit > 0 {
		opts.Limit = params.Limit
	}

	opts.PaymentVerified = params.PaymentVerified
	opts.Country = strings.TrimSpace(params.Country)

	// Parse skills
	if params.Skills != "" {
		tokens := strings.Split(params.Skills, ",")
		for _, token := range tokens {
			trimmed := strings.TrimSpace(token)
			if trimmed != "" {
				opts.Skills = append(opts.Skills, trimmed)
			}
		}
	}

	// Parse job_type
	if params.JobType != "" {
		parsed, err := parseEnumFilterValue(params.JobType, "job_type", jobTypeCodeFromLabel, jobTypeAcceptedLabels())
		if err != nil {
			return opts, err
		}
		opts.JobType = parsed
	}

	opts.Duration = strings.TrimSpace(params.Duration)

	if params.HourlyMin > 0 {
		opts.MinHourly = &params.HourlyMin
	}
	if params.HourlyMax > 0 {
		opts.MaxHourly = &params.HourlyMax
	}
	if params.BudgetMin > 0 {
		opts.BudgetMin = &params.BudgetMin
	}
	if params.BudgetMax > 0 {
		opts.BudgetMax = &params.BudgetMax
	}

	opts.Search = strings.TrimSpace(params.Search)

	// Parse sort
	if params.Sort != "" {
		sortLower := strings.ToLower(strings.TrimSpace(params.Sort))
		switch sortLower {
		case "publish_time_asc", "published_on_asc":
			opts.SortField = SortPublishTime
			opts.SortAscending = true
		case "publish_time_desc", "published_on_desc":
			opts.SortField = SortPublishTime
			opts.SortAscending = false
		case "last_visited_asc":
			opts.SortField = SortLastVisited
			opts.SortAscending = true
		case "last_visited_desc":
			opts.SortField = SortLastVisited
			opts.SortAscending = false
		}
	}

	return opts, nil
}

