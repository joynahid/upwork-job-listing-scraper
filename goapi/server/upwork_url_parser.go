package server

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ParseUpworkSearchURL converts an Upwork search URL into Go API compatible query parameters.
func ParseUpworkSearchURL(raw string) (url.Values, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("upwork URL must not be empty")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("upwork URL must be absolute")
	}

	query := parsed.Query()
	result := url.Values{}

	for key, values := range query {
		if len(values) == 0 {
			continue
		}
		value := strings.TrimSpace(values[0])
		if value == "" {
			continue
		}

		switch strings.ToLower(key) {
		case "q":
			result.Set("search", value)
		case "payment_verified", "client_payment_verification_status":
			if parsedBool, ok := parseUpworkBool(value); ok {
				result.Set("payment_verified", strconv.FormatBool(parsedBool))
			}
		case "t", "job_type":
			result.Set("t", value)
		case "contractor_tier", "exp_level", "experience_level":
			if tier := parseUpworkContractorTier(value); tier != "" {
				result.Set("contractor_tier", tier)
			} else {
				result.Set("contractor_tier", value)
			}
		case "contract_to_hire":
			if parsedBool, ok := parseUpworkBool(value); ok {
				result.Set("contract_to_hire", strconv.FormatBool(parsedBool))
			}
		case "duration_v3", "duration":
			result.Set("duration_v3", value)
		case "hourly_rate", "hourly":
			result.Set("hourly_rate", value)
		case "amount", "fixed_budget":
			result.Set("amount", value)
		case "client_hires":
			result.Set("client_hires", value)
		case "location":
			result.Set("location", value)
		case "timezone":
			result.Set("timezone", value)
		case "workload":
			result.Set("workload", strings.ToLower(value))
		case "proposals":
			if tier := parseUpworkProposalsTier(value); tier != "" {
				result.Set("proposals", tier)
			} else {
				result.Set("proposals", value)
			}
		case "previous_clients":
			result.Set("previous_clients", value)
		case "sort":
			result.Set("sort", value)
		case "subcategory2_uid", "subcategory":
			result.Set("subcategory2_uid", value)
		default:
			lower := strings.ToLower(key)
			if _, ok := supportedAPIParams[lower]; ok {
				result.Set(lower, value)
			}
		}
	}

	return result, nil
}

var supportedAPIParams = map[string]struct{}{
	"limit":            {},
	"offset":           {},
	"payment_verified": {},
	"amount":           {},
	"client_hires":     {},
	"contract_to_hire": {},
	"contractor_tier":  {},
	"duration_v3":      {},
	"hourly_rate":      {},
	"location":         {},
	"previous_clients": {},
	"proposals":        {},
	"sort":             {},
	"subcategory2_uid": {},
	"t":                {},
	"timezone":         {},
	"workload":         {},
	"search":           {},
	"q":                {},
}

func parseUpworkBool(value string) (bool, bool) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "1", "true", "yes":
		return true, true
	case "0", "false", "no":
		return false, true
	default:
		return false, false
	}
}

func parseUpworkJobType(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "0", "hourly", "hourly-job", "hourlyjob":
		return "hourly"
	case "1", "fixed", "fixed-price", "fixedprice", "fixed_price":
		return "fixed-price"
	case "0,1", "1,0", "hourly,fixed", "fixed,hourly":
		return ""
	default:
		return normalized
	}
}

func parseUpworkContractorTier(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return ""
	}
	if strings.Contains(normalized, ",") {
		normalized = strings.Split(normalized, ",")[0]
	}

	tierMap := map[string]string{
		"1":            "entry",
		"entry":        "entry",
		"entry-level":  "entry",
		"beginner":     "entry",
		"2":            "intermediate",
		"intermediate": "intermediate",
		"mid":          "intermediate",
		"mid-level":    "intermediate",
		"3":            "expert",
		"expert":       "expert",
		"expert-level": "expert",
		"advanced":     "expert",
	}

	return tierMap[normalized]
}

func parseUpworkStatus(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	statusMap := map[string]string{
		"open":     "open",
		"opened":   "open",
		"1":        "open",
		"closed":   "closed",
		"inactive": "closed",
		"archived": "closed",
		"2":        "closed",
	}
	return statusMap[normalized]
}

func parseUpworkCountry(value string) string {
	if value == "" {
		return ""
	}
	parts := strings.Split(value, ",")
	primary := strings.TrimSpace(parts[0])
	if primary == "" {
		return ""
	}
	return strings.ToUpper(primary)
}

func normalizeCommaSeparated(value string) string {
	tokens := strings.Split(value, ",")
	cleaned := make([]string, 0, len(tokens))
	for _, token := range tokens {
		trimmed := strings.TrimSpace(token)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return strings.Join(cleaned, ",")
}

func parseUpworkRange(value string) (*float64, *float64) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return nil, nil
	}
	normalized = strings.ReplaceAll(normalized, "+", "-")
	parts := strings.SplitN(normalized, "-", 2)

	var minVal *float64
	if len(parts) > 0 {
		if v, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64); err == nil {
			minVal = &v
		}
	}

	var maxVal *float64
	if len(parts) == 2 {
		if v, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
			maxVal = &v
		}
	}

	return minVal, maxVal
}

func parseUpworkMultiRange(value string) (*float64, *float64) {
	ranges := strings.Split(value, ",")
	var minVal *float64
	var maxVal *float64

	for _, segment := range ranges {
		lo, hi := parseUpworkRange(strings.TrimSpace(segment))
		if lo != nil {
			if minVal == nil || *lo < *minVal {
				minValCopy := *lo
				minVal = &minValCopy
			}
		}
		if hi != nil {
			if maxVal == nil || *hi > *maxVal {
				maxValCopy := *hi
				maxVal = &maxValCopy
			}
		} else if hi == nil {
			// Open upper bound means we treat as unbounded maximum
			maxVal = nil
		}
	}

	return minVal, maxVal
}

func parseUpworkDuration(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	durationMap := map[string]string{
		"week":        "Less than 1 month",
		"weeks":       "Less than 1 month",
		"month":       "1 to 3 months",
		"months":      "1 to 3 months",
		"semester":    "3 to 6 months",
		"ongoing":     "More than 6 months",
		"more_than_6": "More than 6 months",
	}

	if mapped, ok := durationMap[normalized]; ok {
		return mapped
	}
	return value
}

func parseUpworkProposalsTier(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return ""
	}
	// Upwork uses ranges like "0-4", "5-9", "10-15", "15-20", "20-50", "50+"
	normalized = strings.ReplaceAll(normalized, " ", "")
	normalized = strings.ReplaceAll(normalized, "+", "-")
	return normalized
}

func parseUpworkSort(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	sortMap := map[string]string{
		"recency":       "publish_time_desc",
		"relevance":     "publish_time_desc",
		"client_rating": "last_visited_desc",
		"duration":      "publish_time_desc",
		"budget":        "budget_desc",
		"duration_asc":  "publish_time_asc",
		"client_spend":  "budget_desc",
		"client_recent": "last_visited_desc",
	}

	if mapped, ok := sortMap[normalized]; ok {
		return mapped
	}
	return normalized
}

func parseUpworkCreatedTime(value string) string {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	now := time.Now().UTC()

	windowMap := map[string]time.Duration{
		"LAST_24_HOURS": 24 * time.Hour,
		"LAST_3_DAYS":   72 * time.Hour,
		"LAST_7_DAYS":   7 * 24 * time.Hour,
		"LAST_14_DAYS":  14 * 24 * time.Hour,
		"LAST_30_DAYS":  30 * 24 * time.Hour,
		"PAST_24_HOURS": 24 * time.Hour,
		"PAST_WEEK":     7 * 24 * time.Hour,
	}

	if duration, ok := windowMap[normalized]; ok {
		return now.Add(-duration).Format(time.RFC3339)
	}

	// If the value is an ISO timestamp, pass it through
	if _, err := time.Parse(time.RFC3339, value); err == nil {
		return value
	}

	return ""
}

func formatFloat(value float64) string {
	if value == float64(int64(value)) {
		return strconv.FormatInt(int64(value), 10)
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}
