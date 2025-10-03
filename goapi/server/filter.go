package server

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

const (
	DefaultSortField     sortField = SortPublishTime
	DefaultSortAscending bool      = false
)

type NumericRange struct {
	Min *float64
	Max *float64
}

func (r NumericRange) contains(value float64) bool {
	if r.Min != nil && value < *r.Min {
		return false
	}
	if r.Max != nil && value > *r.Max {
		return false
	}
	return true
}

func (r NumericRange) String() string {
	switch {
	case r.Min != nil && r.Max != nil:
		return fmt.Sprintf("%.2f-%.2f", *r.Min, *r.Max)
	case r.Min != nil:
		return fmt.Sprintf("%.2f-", *r.Min)
	case r.Max != nil:
		return fmt.Sprintf("-%.2f", *r.Max)
	default:
		return ""
	}
}

type IntRange struct {
	Min *int
	Max *int
}

func (r IntRange) contains(value int) bool {
	if r.Min != nil && value < *r.Min {
		return false
	}
	if r.Max != nil && value > *r.Max {
		return false
	}
	return true
}

func (r IntRange) String() string {
	switch {
	case r.Min != nil && r.Max != nil:
		return fmt.Sprintf("%d-%d", *r.Min, *r.Max)
	case r.Min != nil:
		return fmt.Sprintf("%d-", *r.Min)
	case r.Max != nil:
		return fmt.Sprintf("-%d", *r.Max)
	default:
		return ""
	}
}

// FilterOptions describes the supported /jobs filters.
type FilterOptions struct {
	Limit               int
	Offset              int
	PaymentVerified     *bool
	ContractorTierCodes []int
	JobTypeCodes        []int
	DurationLabels      []string
	WorkloadValues      []string
	ContractToHire      *bool
	BudgetRanges        []NumericRange
	HourlyRanges        []NumericRange
	ClientHiresRanges   []IntRange
	LocationRegions     []string
	Timezones           []string
	Proposals           []string
	PreviousClients     string
	CategoryGroupIDs    []string
	SortField           sortField
	SortAscending       bool
	SearchQuery         string
	SearchExpression    *SearchExpression
	UpworkURL           string
}

func parseFilterOptions(values url.Values) (FilterOptions, error) {
	opts := FilterOptions{
		Limit:         defaultLimit,
		SortField:     DefaultSortField,
		SortAscending: DefaultSortAscending,
	}

	opts.UpworkURL = strings.TrimSpace(firstQuery(values, "upwork_url"))

	if raw := firstQuery(values, "search"); raw != "" {
		if err := opts.ApplySearchQuery(raw); err != nil {
			return opts, err
		}
	} else if raw := firstQuery(values, "q"); raw != "" {
		if err := opts.ApplySearchQuery(raw); err != nil {
			return opts, err
		}
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
		parsed, err := parseFlexibleBool(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid payment_verified parameter")
		}
		opts.PaymentVerified = &parsed
	}

	if raw := firstQuery(values, "contract_to_hire"); raw != "" {
		parsed, err := parseFlexibleBool(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid contract_to_hire parameter")
		}
		opts.ContractToHire = &parsed
	}

	if raw := firstQuery(values, "contractor_tier"); raw != "" {
		tiers, err := parseContractorTierList(raw)
		if err != nil {
			return opts, err
		}
		opts.ContractorTierCodes = tiers
	}

	if raw := firstQuery(values, "t"); raw != "" {
		types, err := parseJobTypeList(raw)
		if err != nil {
			return opts, err
		}
		opts.JobTypeCodes = types
	} else if raw := firstQuery(values, "job_type"); raw != "" {
		types, err := parseJobTypeList(raw)
		if err != nil {
			return opts, err
		}
		opts.JobTypeCodes = types
	}

	if raw := firstQuery(values, "duration_v3"); raw != "" {
		opts.DurationLabels = parseDurationTokens(raw)
	}

	if raw := firstQuery(values, "workload"); raw != "" {
		opts.WorkloadValues = parseCSVLower(raw)
	}

	if raw := firstQuery(values, "amount"); raw != "" {
		ranges, err := parseNumericRanges(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid amount parameter: %w", err)
		}
		opts.BudgetRanges = ranges
	}

	if raw := firstQuery(values, "hourly_rate"); raw != "" {
		ranges, err := parseNumericRanges(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid hourly_rate parameter: %w", err)
		}
		opts.HourlyRanges = ranges
	}

	if raw := firstQuery(values, "client_hires"); raw != "" {
		ranges, err := parseIntRanges(raw)
		if err != nil {
			return opts, fmt.Errorf("invalid client_hires parameter: %w", err)
		}
		opts.ClientHiresRanges = ranges
	}

	if raw := firstQuery(values, "location"); raw != "" {
		opts.LocationRegions = parseCSVLower(raw)
	}

	if raw := firstQuery(values, "timezone"); raw != "" {
		opts.Timezones = parseCSV(raw)
	}

	if raw := firstQuery(values, "proposals"); raw != "" {
		opts.Proposals = parseCSVNormalized(raw)
	}

	if raw := firstQuery(values, "previous_clients"); raw != "" {
		opts.PreviousClients = strings.ToLower(strings.TrimSpace(raw))
	}

	if raw := firstQuery(values, "subcategory2_uid"); raw != "" {
		opts.CategoryGroupIDs = parseCSVNormalized(raw)
	}

	if raw := firstQuery(values, "sort"); raw != "" {
		applySortParam(&opts, raw)
	}

	if raw := firstQuery(values, "upwork_url"); raw != "" {
		opts.UpworkURL = strings.TrimSpace(raw)
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
	if len(opts.ContractorTierCodes) > 0 {
		parts = append(parts, fmt.Sprintf("contractor_tier=%s", joinTierLabels(opts.ContractorTierCodes)))
	}
	if len(opts.JobTypeCodes) > 0 {
		parts = append(parts, fmt.Sprintf("job_type=%s", joinJobTypeLabels(opts.JobTypeCodes)))
	}
	if len(opts.DurationLabels) > 0 {
		parts = append(parts, fmt.Sprintf("duration_v3=%s", strings.Join(opts.DurationLabels, ",")))
	}
	if len(opts.WorkloadValues) > 0 {
		parts = append(parts, fmt.Sprintf("workload=%s", strings.Join(opts.WorkloadValues, ",")))
	}
	if opts.ContractToHire != nil {
		parts = append(parts, fmt.Sprintf("contract_to_hire=%t", *opts.ContractToHire))
	}
	if len(opts.BudgetRanges) > 0 {
		parts = append(parts, fmt.Sprintf("amount=%s", joinNumericRanges(opts.BudgetRanges)))
	}
	if len(opts.HourlyRanges) > 0 {
		parts = append(parts, fmt.Sprintf("hourly_rate=%s", joinNumericRanges(opts.HourlyRanges)))
	}
	if len(opts.ClientHiresRanges) > 0 {
		parts = append(parts, fmt.Sprintf("client_hires=%s", joinIntRanges(opts.ClientHiresRanges)))
	}
	if len(opts.LocationRegions) > 0 {
		parts = append(parts, fmt.Sprintf("location=%s", strings.Join(opts.LocationRegions, ",")))
	}
	if len(opts.Timezones) > 0 {
		parts = append(parts, fmt.Sprintf("timezone=%s", strings.Join(opts.Timezones, ",")))
	}
	if len(opts.Proposals) > 0 {
		parts = append(parts, fmt.Sprintf("proposals=%s", strings.Join(opts.Proposals, ",")))
	}
	if opts.PreviousClients != "" {
		parts = append(parts, fmt.Sprintf("previous_clients=%s", opts.PreviousClients))
	}
	if len(opts.CategoryGroupIDs) > 0 {
		parts = append(parts, fmt.Sprintf("subcategory2_uid=%s", strings.Join(opts.CategoryGroupIDs, ",")))
	}
	if opts.SearchQuery != "" {
		parts = append(parts, fmt.Sprintf("search=%q", opts.SearchQuery))
	}
	if opts.UpworkURL != "" {
		parts = append(parts, fmt.Sprintf("upwork_url=%s", opts.UpworkURL))
	}

	sortLabel := "last_visited_desc"
	switch opts.SortField {
	case SortPublishTime:
		if opts.SortAscending {
			sortLabel = "publish_time_asc"
		} else {
			sortLabel = "publish_time_desc"
		}
	case SortBudget:
		if opts.SortAscending {
			sortLabel = "budget_asc"
		} else {
			sortLabel = "budget_desc"
		}
	default:
		if opts.SortAscending {
			sortLabel = "last_visited_asc"
		}
	}
	parts = append(parts, fmt.Sprintf("sort=%s", sortLabel))

	return strings.Join(parts, ", ")
}

func (opts *FilterOptions) ApplySearchQuery(raw string) error {
	if opts == nil {
		return fmt.Errorf("filter options not initialized")
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		opts.SearchQuery = ""
		opts.SearchExpression = nil
		return nil
	}

	expr, err := ParseSearchQuery(trimmed)
	if err != nil {
		return err
	}

	opts.SearchQuery = trimmed
	opts.SearchExpression = expr
	return nil
}

func parseFlexibleBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true, nil
	case "0", "false", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value")
	}
}

func parseContractorTierList(raw string) ([]int, error) {
	tokens := strings.Split(raw, ",")
	seen := make(map[int]struct{})
	result := make([]int, 0, len(tokens))

	for _, token := range tokens {
		trimmed := strings.TrimSpace(token)
		if trimmed == "" {
			continue
		}
		if code, err := strconv.Atoi(trimmed); err == nil {
			if code < 1 || code > 3 {
				return nil, fmt.Errorf("contractor_tier must be between 1 and 3")
			}
			if _, ok := seen[code]; !ok {
				seen[code] = struct{}{}
				result = append(result, code)
			}
			continue
		}
		if code, ok := contractorTierCodeFromLabel(trimmed); ok {
			if _, exists := seen[code]; !exists {
				seen[code] = struct{}{}
				result = append(result, code)
			}
			continue
		}
		return nil, fmt.Errorf("invalid contractor_tier value: %s", trimmed)
	}

	sort.Ints(result)
	return result, nil
}

func parseJobTypeList(raw string) ([]int, error) {
	tokens := strings.Split(raw, ",")
	seen := make(map[int]struct{})
	result := make([]int, 0, len(tokens))

	for _, token := range tokens {
		trimmed := strings.TrimSpace(token)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		switch lower {
		case "0":
			if _, ok := seen[1]; !ok {
				seen[1] = struct{}{}
				result = append(result, 1)
			}
			continue
		case "1":
			if _, ok := seen[2]; !ok {
				seen[2] = struct{}{}
				result = append(result, 2)
			}
			continue
		}
		if code, ok := jobTypeCodeFromLabel(trimmed); ok {
			if _, exists := seen[code]; !exists {
				seen[code] = struct{}{}
				result = append(result, code)
			}
			continue
		}
		return nil, fmt.Errorf("invalid job type value: %s", trimmed)
	}

	sort.Ints(result)
	return result, nil
}

func parseDurationTokens(raw string) []string {
	tokens := strings.Split(raw, ",")
	result := make([]string, 0, len(tokens))

	for _, token := range tokens {
		trimmed := strings.TrimSpace(token)
		if trimmed == "" {
			continue
		}
		mapped := parseUpworkDuration(trimmed)
		if mapped == "" {
			mapped = trimmed
		}
		if !containsString(result, mapped, true) {
			result = append(result, mapped)
		}
	}

	return result
}

func parseCSV(raw string) []string {
	tokens := strings.Split(raw, ",")
	result := make([]string, 0, len(tokens))
	for _, token := range tokens {
		trimmed := strings.TrimSpace(token)
		if trimmed != "" && !containsString(result, trimmed, false) {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseCSVLower(raw string) []string {
	tokens := strings.Split(raw, ",")
	result := make([]string, 0, len(tokens))
	for _, token := range tokens {
		trimmed := strings.ToLower(strings.TrimSpace(token))
		if trimmed != "" && !containsString(result, trimmed, false) {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseCSVNormalized(raw string) []string {
	tokens := strings.Split(raw, ",")
	result := make([]string, 0, len(tokens))
	for _, token := range tokens {
		trimmed := strings.TrimSpace(token)
		if trimmed != "" && !containsString(result, trimmed, true) {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseNumericRanges(raw string) ([]NumericRange, error) {
	tokens := strings.Split(raw, ",")
	result := make([]NumericRange, 0, len(tokens))

	for _, token := range tokens {
		trimmed := strings.TrimSpace(token)
		if trimmed == "" {
			continue
		}
		var minPtr, maxPtr *float64
		if strings.Contains(trimmed, "-") {
			parts := strings.SplitN(trimmed, "-", 2)
			if parts[0] != "" {
				value, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
				if err != nil {
					return nil, err
				}
				minPtr = &value
			}
			if len(parts) > 1 && parts[1] != "" {
				value, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				if err != nil {
					return nil, err
				}
				maxPtr = &value
			}
		} else {
			value, err := strconv.ParseFloat(trimmed, 64)
			if err != nil {
				return nil, err
			}
			minPtr = &value
			maxPtr = &value
		}

		if minPtr == nil && maxPtr == nil {
			continue
		}
		minVal := minPtr
		if minPtr != nil {
			copy := *minPtr
			minVal = &copy
		}
		maxVal := maxPtr
		if maxPtr != nil {
			copy := *maxPtr
			maxVal = &copy
		}
		result = append(result, NumericRange{Min: minVal, Max: maxVal})
	}

	return result, nil
}

func parseIntRanges(raw string) ([]IntRange, error) {
	tokens := strings.Split(raw, ",")
	result := make([]IntRange, 0, len(tokens))

	for _, token := range tokens {
		trimmed := strings.TrimSpace(token)
		if trimmed == "" {
			continue
		}
		var minPtr, maxPtr *int
		if strings.Contains(trimmed, "-") {
			parts := strings.SplitN(trimmed, "-", 2)
			if parts[0] != "" {
				value, err := strconv.Atoi(strings.TrimSpace(parts[0]))
				if err != nil {
					return nil, err
				}
				minPtr = &value
			}
			if len(parts) > 1 && parts[1] != "" {
				value, err := strconv.Atoi(strings.TrimSpace(parts[1]))
				if err != nil {
					return nil, err
				}
				maxPtr = &value
			}
		} else {
			value, err := strconv.Atoi(trimmed)
			if err != nil {
				return nil, err
			}
			minPtr = &value
			maxPtr = &value
		}

		minVal := minPtr
		if minPtr != nil {
			copy := *minPtr
			minVal = &copy
		}
		maxVal := maxPtr
		if maxPtr != nil {
			copy := *maxPtr
			maxVal = &copy
		}
		result = append(result, IntRange{Min: minVal, Max: maxVal})
	}

	return result, nil
}

func joinTierLabels(codes []int) string {
	labels := make([]string, 0, len(codes))
	for _, code := range codes {
		labels = append(labels, contractorTierLabelFromCode(code))
	}
	return strings.Join(labels, ",")
}

func joinJobTypeLabels(codes []int) string {
	labels := make([]string, 0, len(codes))
	for _, code := range codes {
		labels = append(labels, jobTypeLabelFromCode(code))
	}
	return strings.Join(labels, ",")
}

func joinNumericRanges(ranges []NumericRange) string {
	parts := make([]string, 0, len(ranges))
	for _, r := range ranges {
		parts = append(parts, r.String())
	}
	return strings.Join(parts, ",")
}

func joinIntRanges(ranges []IntRange) string {
	parts := make([]string, 0, len(ranges))
	for _, r := range ranges {
		parts = append(parts, r.String())
	}
	return strings.Join(parts, ",")
}

func containsString(list []string, candidate string, caseInsensitive bool) bool {
	for _, item := range list {
		if caseInsensitive {
			if strings.EqualFold(item, candidate) {
				return true
			}
		} else if item == candidate {
			return true
		}
	}
	return false
}

func applySortParam(opts *FilterOptions, raw string) {
	if opts == nil {
		return
	}

	normalized := strings.ToLower(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, " ", "")

	switch normalized {
	case "relevance+desc", "relevancedesc", "relevance", "recency", "recency+desc", "recencydesc":
		opts.SortField = SortPublishTime
		opts.SortAscending = false
		return
	case "relevance+asc", "relevanceasc", "recency+asc", "recencyasc":
		opts.SortField = SortPublishTime
		opts.SortAscending = true
		return
	case "publish_time_asc", "posted_on_asc":
		opts.SortField = SortPublishTime
		opts.SortAscending = true
		return
	case "publish_time_desc", "posted_on_desc":
		opts.SortField = SortPublishTime
		opts.SortAscending = false
		return
	case "last_visited_asc":
		opts.SortField = SortLastVisited
		opts.SortAscending = true
		return
	case "last_visited_desc":
		opts.SortField = SortLastVisited
		opts.SortAscending = false
		return
	case "budget_asc":
		opts.SortField = SortBudget
		opts.SortAscending = true
		return
	case "budget_desc":
		opts.SortField = SortBudget
		opts.SortAscending = false
		return
	}

	if mapped := parseUpworkSort(raw); mapped != "" && !strings.EqualFold(mapped, raw) {
		applySortParam(opts, mapped)
	}
}
