package server

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type JobListFilterOptions struct {
	Limit           int
	PaymentVerified *bool
	Country         string
	Skills          []string
	JobType         *int
	Duration        string
	MinHourly       *float64
	MaxHourly       *float64
	BudgetMin       *float64
	BudgetMax       *float64
	Search          string
	SortField       sortField
	SortAscending   bool
}

func parseJobListFilterOptions(values url.Values) (JobListFilterOptions, error) {
	opts := JobListFilterOptions{
		Limit:         defaultLimit,
		SortField:     SortLastVisited,
		SortAscending: false,
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

	if raw := firstQuery(values, "payment_verified"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid payment_verified parameter")
		}
		opts.PaymentVerified = &parsed
	}

	opts.Country = strings.TrimSpace(firstQuery(values, "country"))

	if raw := firstQuery(values, "skills"); raw != "" {
		tokens := strings.Split(raw, ",")
		for _, token := range tokens {
			token = strings.TrimSpace(token)
			if token != "" {
				opts.Skills = append(opts.Skills, token)
			}
		}
	}

	if raw := firstQuery(values, "job_type"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid job_type parameter")
		}
		opts.JobType = &value
	}

	opts.Duration = strings.TrimSpace(firstQuery(values, "duration"))

	if raw := firstQuery(values, "hourly_min"); raw != "" {
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return opts, fmt.Errorf("invalid hourly_min parameter")
		}
		opts.MinHourly = &value
	}

	if raw := firstQuery(values, "hourly_max"); raw != "" {
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return opts, fmt.Errorf("invalid hourly_max parameter")
		}
		opts.MaxHourly = &value
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

	opts.Search = strings.TrimSpace(firstQuery(values, "search"))

	if raw := strings.ToLower(strings.TrimSpace(firstQuery(values, "sort"))); raw != "" {
		switch raw {
		case "published_on_asc":
			opts.SortField = SortPostedOn
			opts.SortAscending = true
		case "published_on_desc":
			opts.SortField = SortPostedOn
		case "last_visited_asc":
			opts.SortField = SortLastVisited
			opts.SortAscending = true
		case "last_visited_desc":
			opts.SortField = SortLastVisited
		default:
			return opts, fmt.Errorf("invalid sort parameter")
		}
	}

	return opts, nil
}

func formatJobListFilterOptions(opts JobListFilterOptions) string {
	parts := []string{fmt.Sprintf("limit=%d", opts.Limit)}

	if opts.PaymentVerified != nil {
		parts = append(parts, fmt.Sprintf("payment_verified=%t", *opts.PaymentVerified))
	}
	if opts.Country != "" {
		parts = append(parts, fmt.Sprintf("country=%s", strings.ToUpper(opts.Country)))
	}
	if len(opts.Skills) > 0 {
		parts = append(parts, fmt.Sprintf("skills=%s", strings.Join(opts.Skills, ",")))
	}
	if opts.JobType != nil {
		parts = append(parts, fmt.Sprintf("job_type=%d", *opts.JobType))
	}
	if opts.Duration != "" {
		parts = append(parts, fmt.Sprintf("duration=%s", opts.Duration))
	}
	if opts.MinHourly != nil {
		parts = append(parts, fmt.Sprintf("hourly_min=%.2f", *opts.MinHourly))
	}
	if opts.MaxHourly != nil {
		parts = append(parts, fmt.Sprintf("hourly_max=%.2f", *opts.MaxHourly))
	}
	if opts.BudgetMin != nil {
		parts = append(parts, fmt.Sprintf("budget_min=%.2f", *opts.BudgetMin))
	}
	if opts.BudgetMax != nil {
		parts = append(parts, fmt.Sprintf("budget_max=%.2f", *opts.BudgetMax))
	}
	if opts.Search != "" {
		parts = append(parts, fmt.Sprintf("search=%s", opts.Search))
	}

	sortLabel := "last_visited_desc"
	if opts.SortField == SortPostedOn {
		if opts.SortAscending {
			sortLabel = "published_on_asc"
		} else {
			sortLabel = "published_on_desc"
		}
	} else if opts.SortAscending {
		sortLabel = "last_visited_asc"
	}

	parts = append(parts, fmt.Sprintf("sort=%s", sortLabel))

	return strings.Join(parts, ", ")
}
