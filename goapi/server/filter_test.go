package server

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestParseFilterOptionsSuccess(t *testing.T) {
	values := url.Values{}
	values.Set("search", "python AND golang")
	values.Set("limit", "30")
	values.Set("offset", "60")
	values.Set("payment_verified", "true")
	values.Set("contract_to_hire", "1")
	values.Set("contractor_tier", "entry,3")
	values.Set("t", "hourly,1")
	values.Set("duration_v3", "week,ongoing,Custom")
	values.Set("workload", "PartTime,Fulltime,parttime")
	values.Set("amount", "100-200,300-")
	values.Set("hourly_rate", "15-25")
	values.Set("client_hires", "1-5,10-")
	values.Set("location", "United States, Canada")
	values.Set("timezone", "UTC,America/New_York")
	values.Set("proposals", "0-4,5-9")
	values.Set("previous_clients", "Multiple")
	values.Set("subcategory2_uid", "123,456,123")
	values.Set("sort", "budget_asc")
	values.Set("upwork_url", " https://www.upwork.com/nx/jobs/search/?q=python ")

	opts, err := parseFilterOptions(values)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Limit != 30 {
		t.Fatalf("expected limit 30, got %d", opts.Limit)
	}
	if opts.Offset != 60 {
		t.Fatalf("expected offset 60, got %d", opts.Offset)
	}

	if opts.PaymentVerified == nil || !*opts.PaymentVerified {
		t.Fatalf("expected payment_verified to be true, got %+v", opts.PaymentVerified)
	}
	if opts.ContractToHire == nil || !*opts.ContractToHire {
		t.Fatalf("expected contract_to_hire to be true, got %+v", opts.ContractToHire)
	}

	if !reflect.DeepEqual(opts.ContractorTierCodes, []int{1, 3}) {
		t.Fatalf("unexpected contractor tiers: %+v", opts.ContractorTierCodes)
	}
	if !reflect.DeepEqual(opts.JobTypeCodes, []int{1, 2}) {
		t.Fatalf("unexpected job type codes: %+v", opts.JobTypeCodes)
	}

	expectedDurations := []string{"Less than 1 month", "More than 6 months", "Custom"}
	if !reflect.DeepEqual(opts.DurationLabels, expectedDurations) {
		t.Fatalf("unexpected duration labels: %+v", opts.DurationLabels)
	}

	if !reflect.DeepEqual(opts.WorkloadValues, []string{"parttime", "fulltime"}) {
		t.Fatalf("unexpected workload values: %+v", opts.WorkloadValues)
	}

	if len(opts.BudgetRanges) != 2 {
		t.Fatalf("expected 2 budget ranges, got %d", len(opts.BudgetRanges))
	}
	if opts.BudgetRanges[0].Min == nil || *opts.BudgetRanges[0].Min != 100 || opts.BudgetRanges[0].Max == nil || *opts.BudgetRanges[0].Max != 200 {
		t.Fatalf("unexpected first budget range: %+v", opts.BudgetRanges[0])
	}
	if opts.BudgetRanges[1].Min == nil || *opts.BudgetRanges[1].Min != 300 || opts.BudgetRanges[1].Max != nil {
		t.Fatalf("unexpected second budget range: %+v", opts.BudgetRanges[1])
	}

	if len(opts.HourlyRanges) != 1 || opts.HourlyRanges[0].Min == nil || *opts.HourlyRanges[0].Min != 15 || opts.HourlyRanges[0].Max == nil || *opts.HourlyRanges[0].Max != 25 {
		t.Fatalf("unexpected hourly ranges: %+v", opts.HourlyRanges)
	}

	if len(opts.ClientHiresRanges) != 2 {
		t.Fatalf("expected 2 client_hires ranges, got %d", len(opts.ClientHiresRanges))
	}
	if opts.ClientHiresRanges[0].Min == nil || *opts.ClientHiresRanges[0].Min != 1 || opts.ClientHiresRanges[0].Max == nil || *opts.ClientHiresRanges[0].Max != 5 {
		t.Fatalf("unexpected first client_hires range: %+v", opts.ClientHiresRanges[0])
	}
	if opts.ClientHiresRanges[1].Min == nil || *opts.ClientHiresRanges[1].Min != 10 || opts.ClientHiresRanges[1].Max != nil {
		t.Fatalf("unexpected second client_hires range: %+v", opts.ClientHiresRanges[1])
	}

	if !reflect.DeepEqual(opts.LocationRegions, []string{"united states", "canada"}) {
		t.Fatalf("unexpected location regions: %+v", opts.LocationRegions)
	}
	if !reflect.DeepEqual(opts.Timezones, []string{"UTC", "America/New_York"}) {
		t.Fatalf("unexpected timezone values: %+v", opts.Timezones)
	}
	if !reflect.DeepEqual(opts.Proposals, []string{"0-4", "5-9"}) {
		t.Fatalf("unexpected proposals: %+v", opts.Proposals)
	}

	if opts.PreviousClients != "multiple" {
		t.Fatalf("expected previous_clients \"multiple\", got %q", opts.PreviousClients)
	}
	if !reflect.DeepEqual(opts.CategoryGroupIDs, []string{"123", "456"}) {
		t.Fatalf("unexpected subcategory IDs: %+v", opts.CategoryGroupIDs)
	}

	if opts.SortField != SortBudget || !opts.SortAscending {
		t.Fatalf("unexpected sort configuration: field=%v ascending=%v", opts.SortField, opts.SortAscending)
	}

	if opts.SearchQuery != "python AND golang" {
		t.Fatalf("unexpected search query: %q", opts.SearchQuery)
	}
	if opts.SearchExpression == nil {
		t.Fatalf("expected search expression to be parsed")
	}

	if opts.UpworkURL != "https://www.upwork.com/nx/jobs/search/?q=python" {
		t.Fatalf("unexpected upwork URL: %q", opts.UpworkURL)
	}
}

func TestParseFilterOptionsRejectsInvalidBoolean(t *testing.T) {
	values := url.Values{}
	values.Set("contract_to_hire", "maybe")

	if _, err := parseFilterOptions(values); err == nil {
		t.Fatalf("expected error for invalid boolean, got nil")
	} else if !strings.Contains(err.Error(), "invalid contract_to_hire parameter") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseFilterOptionsAppliesLimitClampAndSearchFallback(t *testing.T) {
	values := url.Values{}
	values.Set("limit", "999")
	values.Set("q", "  data science ")
	values.Set("sort", "recency+asc")

	opts, err := parseFilterOptions(values)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.Limit != maxLimit {
		t.Fatalf("expected limit to clamp to %d, got %d", maxLimit, opts.Limit)
	}

	if opts.SearchQuery != "data science" {
		t.Fatalf("unexpected search query: %q", opts.SearchQuery)
	}
	if opts.SearchExpression == nil {
		t.Fatalf("expected search expression when using q parameter")
	}

	if opts.SortField != SortPublishTime || !opts.SortAscending {
		t.Fatalf("unexpected sort configuration: field=%v ascending=%v", opts.SortField, opts.SortAscending)
	}
}
