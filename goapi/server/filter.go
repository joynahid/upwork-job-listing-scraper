package server

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultSortField     sortField = SortLastVisited
	DefaultSortAscending bool      = false
)

// FilterOptions describes query parameters accepted by /jobs.
type FilterOptions struct {
	Limit             int
	PaymentVerified   *bool
	CategorySlug      string
	CategoryGroupSlug string
	Status            *int
	JobType           *int
	ContractorTier    *int
	Country           string
	Tags              []string
	PostedAfter       *time.Time
	PostedBefore      *time.Time
	BudgetMin         *float64
	BudgetMax         *float64
	SortField         sortField
	SortAscending     bool
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
		value, err := strconv.Atoi(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid status parameter")
		}
		opts.Status = &value
	}

	if raw := firstQuery(values, "job_type"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid job_type parameter")
		}
		opts.JobType = &value
	}

	if raw := firstQuery(values, "contractor_tier"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid contractor_tier parameter")
		}
		opts.ContractorTier = &value
	}

	opts.Country = strings.TrimSpace(firstQuery(values, "country"))

	if raw := firstQuery(values, "tags"); raw != "" {
		tokens := strings.Split(raw, ",")
		for _, token := range tokens {
			trimmed := strings.TrimSpace(token)
			if trimmed != "" {
				opts.Tags = append(opts.Tags, trimmed)
			}
		}
	}

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

	if raw := strings.ToLower(strings.TrimSpace(firstQuery(values, "sort"))); raw != "" {
		switch raw {
		case "posted_on_asc":
			opts.SortField = SortPostedOn
			opts.SortAscending = true
		case "posted_on_desc":
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

func formatFilterOptions(opts FilterOptions) string {
	parts := []string{fmt.Sprintf("limit=%d", opts.Limit)}

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
		parts = append(parts, fmt.Sprintf("status=%d", *opts.Status))
	}
	if opts.JobType != nil {
		parts = append(parts, fmt.Sprintf("job_type=%d", *opts.JobType))
	}
	if opts.ContractorTier != nil {
		parts = append(parts, fmt.Sprintf("contractor_tier=%d", *opts.ContractorTier))
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
	if opts.BudgetMin != nil {
		parts = append(parts, fmt.Sprintf("budget_min=%.2f", *opts.BudgetMin))
	}
	if opts.BudgetMax != nil {
		parts = append(parts, fmt.Sprintf("budget_max=%.2f", *opts.BudgetMax))
	}
	if len(opts.Tags) > 0 {
		parts = append(parts, fmt.Sprintf("tags=%s", strings.Join(opts.Tags, ",")))
	}

	sortLabel := "last_visited_desc"
	if opts.SortField == SortPostedOn {
		if opts.SortAscending {
			sortLabel = "posted_on_asc"
		} else {
			sortLabel = "posted_on_desc"
		}
	} else if opts.SortAscending {
		sortLabel = "last_visited_asc"
	}
	parts = append(parts, fmt.Sprintf("sort=%s", sortLabel))

	return strings.Join(parts, ", ")
}
