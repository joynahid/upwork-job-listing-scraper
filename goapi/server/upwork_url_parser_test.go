package server

import (
	"net/url"
	"testing"
)

func assertURLValuesEqual(t *testing.T, got, want url.Values) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("expected %d parameters, got %d: %+v", len(want), len(got), got)
	}

	for key := range want {
		if got.Get(key) != want.Get(key) {
			t.Fatalf("parameter %q mismatch: expected %q, got %q", key, want.Get(key), got.Get(key))
		}
	}

	for key := range got {
		if _, ok := want[key]; !ok {
			t.Fatalf("unexpected parameter %q present in result", key)
		}
	}
}

func TestParseUpworkSearchURLTransformsKnownParameters(t *testing.T) {
	raw := "https://www.upwork.com/nx/jobs/search/?q=python%20developer&payment_verified=1&t=hourly&contractor_tier=2&contract_to_hire=no&duration=months&hourly_rate=20-50&amount=100-200&client_hires=5&location=United%20States&timezone=UTC&proposals=50%2B&previous_clients=multiple&sort=recency&subcategory=12345&workload=PartTime&limit=40&offset=80"

	got, err := ParseUpworkSearchURL(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := url.Values{}
	want.Set("search", "python developer")
	want.Set("payment_verified", "true")
	want.Set("t", "hourly")
	want.Set("contractor_tier", "intermediate")
	want.Set("contract_to_hire", "false")
	want.Set("duration_v3", "months")
	want.Set("hourly_rate", "20-50")
	want.Set("amount", "100-200")
	want.Set("client_hires", "5")
	want.Set("location", "United States")
	want.Set("timezone", "UTC")
	want.Set("proposals", "50-")
	want.Set("previous_clients", "multiple")
	want.Set("sort", "recency")
	want.Set("subcategory2_uid", "12345")
	want.Set("workload", "parttime")
	want.Set("limit", "40")
	want.Set("offset", "80")

	assertURLValuesEqual(t, got, want)
}

func TestParseUpworkSearchURLAliasesAndCleanup(t *testing.T) {
	raw := "https://www.upwork.com/nx/jobs/search/?client_payment_verification_status=yes&job_type=fixed-price&contractor_tier=beginner,expert&proposals=5%20-%209&workload=Fulltime&unknown=1"

	got, err := ParseUpworkSearchURL(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := url.Values{}
	want.Set("payment_verified", "true")
	want.Set("t", "fixed-price")
	want.Set("contractor_tier", "entry")
	want.Set("proposals", "5-9")
	want.Set("workload", "fulltime")

	assertURLValuesEqual(t, got, want)
}

func TestParseUpworkSearchURLEmptyInput(t *testing.T) {
	if _, err := ParseUpworkSearchURL("   "); err == nil {
		t.Fatalf("expected error for empty input, got nil")
	}
}

func TestParseUpworkSearchURLRequiresAbsoluteURL(t *testing.T) {
	if _, err := ParseUpworkSearchURL("/jobs/search/?q=python"); err == nil {
		t.Fatalf("expected error for relative URL, got nil")
	}
}
