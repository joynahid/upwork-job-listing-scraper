package server

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultSortField     sortField = SortPublishTime
	DefaultSortAscending bool      = false
)

// FilterOptions describes query parameters accepted by /jobs.
type FilterOptions struct {
	Limit                      int
	Offset                     int
	PaymentVerified            *bool
	CategorySlug               string
	CategoryGroupSlug          string
	Status                     *int
	JobType                    *int
	ContractorTier             *int
	Country                    string
	Tags                       []string
	PostedAfter                *time.Time
	PostedBefore               *time.Time
	LastVisitedAfter           *time.Time
	BudgetMin                  *float64
	BudgetMax                  *float64
	HourlyMin                  *float64
	HourlyMax                  *float64
	DurationLabel              string
	Engagement                 string
	BuyerTotalSpentMin         *float64
	BuyerTotalSpentMax         *float64
	BuyerTotalAssignmentsMin   *int
	BuyerTotalAssignmentsMax   *int
	BuyerTotalJobsWithHiresMin *int
	BuyerTotalJobsWithHiresMax *int
	Workload                   string
	IsContractToHire           *bool
	NumberOfPositionsMin       *int
	NumberOfPositionsMax       *int
	WasRenewed                 *bool
	Premium                    *bool
	HideBudget                 *bool
	ProposalsTier              string
	MinJobSuccessScore         *int
	MinOdeskHours              *int
	PrefEnglishSkill           *int
	RisingTalent               *bool
	ShouldHavePortfolio        *bool
	MinHoursWeek               *float64
	SortField                  sortField
	SortAscending              bool
}

func parseFilterOptions(values url.Values) (FilterOptions, error) {
	opts := FilterOptions{
		Limit:         defaultLimit,
		SortField:     DefaultSortField,
		SortAscending: DefaultSortAscending,
	}

	if raw := firstQuery(values, "limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit <= 0 {
			return opts, fmt.Errorf("invalid limit parameter")
		}
		if limit > maxLimit {
			limit = maxLimit
		}
		opts.Limit = limit
	}

	if raw := firstQuery(values, "offset"); raw != "" {
		offset, err := strconv.Atoi(raw)
		if err != nil || offset < 0 {
			return opts, fmt.Errorf("invalid offset parameter")
		}
		if opts.Limit > 0 && opts.Limit <= maxLimit && offset%opts.Limit != 0 {
			return opts, fmt.Errorf("invalid offset parameter (must be multiple of limit)")
		}
		opts.Offset = offset
	}

	if raw := firstQuery(values, "payment_verified"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid payment_verified parameter")
		}
		opts.PaymentVerified = &parsed
	}

	opts.CategorySlug = strings.TrimSpace(firstQuery(values, "category"))
	opts.CategoryGroupSlug = strings.TrimSpace(firstQuery(values, "category_group"))

	if raw := firstQuery(values, "status"); raw != "" {
		parsed, err := parseEnumFilterValue(raw, "status", jobStatusCodeFromLabel, jobStatusAcceptedLabels())
		if err != nil {
			return opts, err
		}
		opts.Status = parsed
	}

	if raw := firstQuery(values, "job_type"); raw != "" {
		parsed, err := parseEnumFilterValue(raw, "job_type", jobTypeCodeFromLabel, jobTypeAcceptedLabels())
		if err != nil {
			return opts, err
		}
		opts.JobType = parsed
	}

	if raw := firstQuery(values, "contractor_tier"); raw != "" {
		parsed, err := parseEnumFilterValue(raw, "contractor_tier", contractorTierCodeFromLabel, contractorTierAcceptedLabels())
		if err != nil {
			return opts, err
		}
		opts.ContractorTier = parsed
	}

	opts.Country = strings.TrimSpace(firstQuery(values, "country"))

	appendTokens := func(raw string) {
		if raw == "" {
			return
		}
		for _, token := range strings.Split(raw, ",") {
			trimmed := strings.TrimSpace(token)
			if trimmed != "" {
				opts.Tags = append(opts.Tags, trimmed)
			}
		}
	}

	appendTokens(firstQuery(values, "tags"))
	appendTokens(firstQuery(values, "skills"))

	if raw := firstQuery(values, "posted_after"); raw != "" {
		t, err := parseTimeParam(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid posted_after parameter")
		}
		opts.PostedAfter = &t
	}

	if raw := firstQuery(values, "posted_before"); raw != "" {
		t, err := parseTimeParam(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid posted_before parameter")
		}
		opts.PostedBefore = &t
	}

	if raw := firstQuery(values, "last_visited_after"); raw != "" {
		t, err := parseTimeParam(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid last_visited_after parameter")
		}
		opts.LastVisitedAfter = &t
	}

	if raw := firstQuery(values, "budget_min"); raw != "" {
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return opts, fmt.Errorf("invalid budget_min parameter")
		}
		opts.BudgetMin = &value
	}

	if raw := firstQuery(values, "budget_max"); raw != "" {
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return opts, fmt.Errorf("invalid budget_max parameter")
		}
		opts.BudgetMax = &value
	}

	if raw := firstQuery(values, "hourly_min"); raw != "" {
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return opts, fmt.Errorf("invalid hourly_min parameter")
		}
		opts.HourlyMin = &value
	}

	if raw := firstQuery(values, "hourly_max"); raw != "" {
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return opts, fmt.Errorf("invalid hourly_max parameter")
		}
		opts.HourlyMax = &value
	}

	opts.DurationLabel = strings.TrimSpace(firstQuery(values, "duration_label"))
	opts.Engagement = strings.TrimSpace(firstQuery(values, "engagement"))

	if raw := firstQuery(values, "buyer.total_spent_min"); raw != "" {
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return opts, fmt.Errorf("invalid buyer.total_spent_min parameter")
		}
		opts.BuyerTotalSpentMin = &value
	}

	if raw := firstQuery(values, "buyer.total_spent_max"); raw != "" {
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return opts, fmt.Errorf("invalid buyer.total_spent_max parameter")
		}
		opts.BuyerTotalSpentMax = &value
	}

	if raw := firstQuery(values, "buyer.total_assignments_min"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return opts, fmt.Errorf("invalid buyer.total_assignments_min parameter")
		}
		opts.BuyerTotalAssignmentsMin = &value
	}

	if raw := firstQuery(values, "buyer.total_assignments_max"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return opts, fmt.Errorf("invalid buyer.total_assignments_max parameter")
		}
		opts.BuyerTotalAssignmentsMax = &value
	}

	if raw := firstQuery(values, "buyer.total_jobs_with_hires_min"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return opts, fmt.Errorf("invalid buyer.total_jobs_with_hires_min parameter")
		}
		opts.BuyerTotalJobsWithHiresMin = &value
	}

	if raw := firstQuery(values, "buyer.total_jobs_with_hires_max"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return opts, fmt.Errorf("invalid buyer.total_jobs_with_hires_max parameter")
		}
		opts.BuyerTotalJobsWithHiresMax = &value
	}

	opts.Workload = strings.TrimSpace(firstQuery(values, "workload"))

	if raw := firstQuery(values, "is_contract_to_hire"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid is_contract_to_hire parameter")
		}
		opts.IsContractToHire = &parsed
	}

	if raw := firstQuery(values, "number_of_positions_min"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return opts, fmt.Errorf("invalid number_of_positions_min parameter")
		}
		opts.NumberOfPositionsMin = &value
	}

	if raw := firstQuery(values, "number_of_positions_max"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return opts, fmt.Errorf("invalid number_of_positions_max parameter")
		}
		opts.NumberOfPositionsMax = &value
	}

	if raw := firstQuery(values, "was_renewed"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid was_renewed parameter")
		}
		opts.WasRenewed = &parsed
	}

	if raw := firstQuery(values, "premium"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid premium parameter")
		}
		opts.Premium = &parsed
	}

	if raw := firstQuery(values, "hide_budget"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid hide_budget parameter")
		}
		opts.HideBudget = &parsed
	}

	opts.ProposalsTier = strings.TrimSpace(firstQuery(values, "proposals_tier"))

	if raw := firstQuery(values, "min_job_success_score"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 || value > 100 {
			return opts, fmt.Errorf("invalid min_job_success_score parameter")
		}
		opts.MinJobSuccessScore = &value
	}

	if raw := firstQuery(values, "min_odesk_hours"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return opts, fmt.Errorf("invalid min_odesk_hours parameter")
		}
		opts.MinOdeskHours = &value
	}

	if raw := firstQuery(values, "pref_english_skill"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 || value > 4 {
			return opts, fmt.Errorf("invalid pref_english_skill parameter (must be 0-4)")
		}
		opts.PrefEnglishSkill = &value
	}

	if raw := firstQuery(values, "rising_talent"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid rising_talent parameter")
		}
		opts.RisingTalent = &parsed
	}

	if raw := firstQuery(values, "should_have_portfolio"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid should_have_portfolio parameter")
		}
		opts.ShouldHavePortfolio = &parsed
	}

	if raw := firstQuery(values, "min_hours_week"); raw != "" {
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return opts, fmt.Errorf("invalid min_hours_week parameter")
		}
		opts.MinHoursWeek = &value
	}

	if raw := strings.ToLower(strings.TrimSpace(firstQuery(values, "sort"))); raw != "" {
		switch raw {
		case "publish_time_asc":
			opts.SortField = SortPublishTime
			opts.SortAscending = true
		case "publish_time_desc":
			opts.SortField = SortPublishTime
		case "last_visited_asc":
			opts.SortField = SortLastVisited
			opts.SortAscending = true
		case "last_visited_desc":
			opts.SortField = SortLastVisited
		case "budget_asc":
			opts.SortField = SortBudget
			opts.SortAscending = true
		case "budget_desc":
			opts.SortField = SortBudget
		default:
			return opts, fmt.Errorf("invalid sort parameter")
		}
	}

	return opts, nil
}

func formatFilterOptions(opts FilterOptions) string {
	parts := []string{fmt.Sprintf("limit=%d", opts.Limit)}

	if opts.Offset > 0 {
		parts = append(parts, fmt.Sprintf("offset=%d", opts.Offset))
	}
	if opts.PaymentVerified != nil {
		parts = append(parts, fmt.Sprintf("payment_verified=%t", *opts.PaymentVerified))
	}
	if opts.CategorySlug != "" {
		parts = append(parts, fmt.Sprintf("category=%s", opts.CategorySlug))
	}
	if opts.CategoryGroupSlug != "" {
		parts = append(parts, fmt.Sprintf("category_group=%s", opts.CategoryGroupSlug))
	}
	if opts.Status != nil {
		parts = append(parts, fmt.Sprintf("status=%s", enumLabelOrNumber(*opts.Status, jobStatusLabelFromCode)))
	}
	if opts.JobType != nil {
		parts = append(parts, fmt.Sprintf("job_type=%s", enumLabelOrNumber(*opts.JobType, jobTypeLabelFromCode)))
	}
	if opts.ContractorTier != nil {
		parts = append(parts, fmt.Sprintf("contractor_tier=%s", enumLabelOrNumber(*opts.ContractorTier, contractorTierLabelFromCode)))
	}
	if opts.Country != "" {
		parts = append(parts, fmt.Sprintf("country=%s", strings.ToUpper(opts.Country)))
	}
	if opts.PostedAfter != nil {
		parts = append(parts, fmt.Sprintf("posted_after=%s", opts.PostedAfter.Format(time.RFC3339)))
	}
	if opts.PostedBefore != nil {
		parts = append(parts, fmt.Sprintf("posted_before=%s", opts.PostedBefore.Format(time.RFC3339)))
	}
	if opts.LastVisitedAfter != nil {
		parts = append(parts, fmt.Sprintf("last_visited_after=%s", opts.LastVisitedAfter.Format(time.RFC3339)))
	}
	if opts.BudgetMin != nil {
		parts = append(parts, fmt.Sprintf("budget_min=%.2f", *opts.BudgetMin))
	}
	if opts.BudgetMax != nil {
		parts = append(parts, fmt.Sprintf("budget_max=%.2f", *opts.BudgetMax))
	}
	if opts.HourlyMin != nil {
		parts = append(parts, fmt.Sprintf("hourly_min=%.2f", *opts.HourlyMin))
	}
	if opts.HourlyMax != nil {
		parts = append(parts, fmt.Sprintf("hourly_max=%.2f", *opts.HourlyMax))
	}
	if opts.DurationLabel != "" {
		parts = append(parts, fmt.Sprintf("duration_label=%s", opts.DurationLabel))
	}
	if opts.Engagement != "" {
		parts = append(parts, fmt.Sprintf("engagement=%s", opts.Engagement))
	}
	if len(opts.Tags) > 0 {
		parts = append(parts, fmt.Sprintf("tags=%s", strings.Join(opts.Tags, ",")))
	}
	if opts.BuyerTotalSpentMin != nil {
		parts = append(parts, fmt.Sprintf("buyer.total_spent_min=%.2f", *opts.BuyerTotalSpentMin))
	}
	if opts.BuyerTotalSpentMax != nil {
		parts = append(parts, fmt.Sprintf("buyer.total_spent_max=%.2f", *opts.BuyerTotalSpentMax))
	}
	if opts.BuyerTotalAssignmentsMin != nil {
		parts = append(parts, fmt.Sprintf("buyer.total_assignments_min=%d", *opts.BuyerTotalAssignmentsMin))
	}
	if opts.BuyerTotalAssignmentsMax != nil {
		parts = append(parts, fmt.Sprintf("buyer.total_assignments_max=%d", *opts.BuyerTotalAssignmentsMax))
	}
	if opts.BuyerTotalJobsWithHiresMin != nil {
		parts = append(parts, fmt.Sprintf("buyer.total_jobs_with_hires_min=%d", *opts.BuyerTotalJobsWithHiresMin))
	}
	if opts.BuyerTotalJobsWithHiresMax != nil {
		parts = append(parts, fmt.Sprintf("buyer.total_jobs_with_hires_max=%d", *opts.BuyerTotalJobsWithHiresMax))
	}

	sortLabel := "last_visited_desc"
	if opts.SortField == SortPublishTime {
		if opts.SortAscending {
			sortLabel = "publish_time_asc"
		} else {
			sortLabel = "publish_time_desc"
		}
	} else if opts.SortField == SortBudget {
		if opts.SortAscending {
			sortLabel = "budget_asc"
		} else {
			sortLabel = "budget_desc"
		}
	} else if opts.SortAscending {
		sortLabel = "last_visited_asc"
	}
	parts = append(parts, fmt.Sprintf("sort=%s", sortLabel))

	return strings.Join(parts, ", ")
}

func parseEnumFilterValue(raw string, paramName string, resolver func(string) (int, bool), accepted []string) (*int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	if numeric, err := strconv.Atoi(trimmed); err == nil {
		value := numeric
		return &value, nil
	}

	if code, ok := resolver(trimmed); ok {
		value := code
		return &value, nil
	}

	return nil, fmt.Errorf("invalid %s parameter (expected one of %s or an integer code)", paramName, strings.Join(accepted, ", "))
}

func enumLabelOrNumber(code int, labeler func(int) string) string {
	label := labeler(code)
	if label != "unknown" {
		return label
	}
	return strconv.Itoa(code)
}
