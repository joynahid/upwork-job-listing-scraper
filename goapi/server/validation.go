package server

import (
	"fmt"
	"net/url"
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
	UpworkURL string `form:"upwork_url" binding:"required,url"`

	derivedParams url.Values `form:"-"`
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
		"UpworkURL":       "upwork_url",
		"Limit":           "limit",
		"Offset":          "offset",
		"PaymentVerified": "payment_verified",
		"Country":         "country",
		"Skills":          "skills",
		"JobType":         "job_type",
		"Duration":        "duration",
		"HourlyMin":       "hourly_min",
		"HourlyMax":       "hourly_max",
		"BudgetMin":       "budget_min",
		"BudgetMax":       "budget_max",
		"Search":          "search",
		"Sort":            "sort",
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
		"upwork_url":       "?upwork_url=https://www.upwork.com/nx/search/jobs/?q=python&hourly_rate=20-40",
		"limit":            "?limit=20",
		"offset":           "?offset=0",
		"payment_verified": "?payment_verified=true",
		"country":          "?country=US",
		"skills":           "?skills=python,react",
		"job_type":         "?job_type=hourly",
		"duration":         "?duration=short-term",
		"hourly_min":       "?hourly_min=25",
		"hourly_max":       "?hourly_max=75",
		"budget_min":       "?budget_min=500",
		"budget_max":       "?budget_max=2000",
		"search":           "?search=(python AND automation)",
		"sort":             "?sort=publish_time_desc",
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

	params.UpworkURL = strings.TrimSpace(params.UpworkURL)

	for key := range c.Request.URL.Query() {
		if strings.EqualFold(key, "upwork_url") {
			continue
		}
		return nil, fmt.Errorf("parameter '%s' is not supported. Only 'upwork_url' may be provided.", key)
	}

	derived, err := ParseUpworkSearchURL(params.UpworkURL)
	if err != nil {
		return nil, fmt.Errorf("invalid upwork_url: %w", err)
	}
	params.derivedParams = derived

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
	combined := url.Values{}

	if params.derivedParams != nil {
		for key, values := range params.derivedParams {
			if len(values) == 0 {
				continue
			}
			value := strings.TrimSpace(values[0])
			if value != "" {
				combined.Set(key, value)
			}
		}
	}

	combined.Set("upwork_url", params.UpworkURL)

	opts, err := parseFilterOptions(combined)
	if err != nil {
		return opts, err
	}
	opts.UpworkURL = strings.TrimSpace(params.UpworkURL)
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
