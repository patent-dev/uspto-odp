//go:build integration
// +build integration

package odp

import (
	"os"
	"testing"
	"time"
)

// TestIntegrationSearchPatentsSort verifies that SearchPatentsWithOptions sort
// keys reach the real USPTO ODP API and actually order the results: the same
// query sorted by filing date ascending vs descending must return filing dates
// in the corresponding monotonic order. Skipped without USPTO_API_KEY.
func TestIntegrationSearchPatentsSort(t *testing.T) {
	apiKey := os.Getenv("USPTO_API_KEY")
	if apiKey == "" {
		t.Skip("USPTO_API_KEY environment variable is required. Set it before running tests")
	}

	client, err := NewClient(&Config{
		BaseURL:    "https://api.uspto.gov",
		APIKey:     apiKey,
		MaxRetries: 2,
		RetryDelay: 1 * time.Second,
		Timeout:    30 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	ctx := testCtx(t)

	const sortField = "applicationMetaData.filingDate"
	query := "applicationMetaData.inventionTitle:semiconductor"

	filingDates := func(order string) []string {
		t.Helper()
		res, err := client.SearchPatentsWithOptions(ctx, query, 0, 10, &PatentSearchOptions{
			Sort: []PatentSearchSort{{Field: sortField, Order: order}},
		})
		if err != nil {
			t.Fatalf("SearchPatentsWithOptions(order=%s) failed: %v", order, err)
		}
		if res == nil || res.PatentFileWrapperDataBag == nil {
			t.Fatalf("SearchPatentsWithOptions(order=%s) returned no data", order)
		}
		var dates []string
		for _, w := range *res.PatentFileWrapperDataBag {
			if w.ApplicationMetaData != nil && w.ApplicationMetaData.FilingDate != nil {
				dates = append(dates, *w.ApplicationMetaData.FilingDate)
			}
		}
		if len(dates) < 2 {
			t.Fatalf("order=%s: need >=2 dated results to check ordering, got %d", order, len(dates))
		}
		return dates
	}

	asc := filingDates("asc")
	for i := 1; i < len(asc); i++ {
		if asc[i] < asc[i-1] {
			t.Errorf("asc sort not non-decreasing: %v", asc)
			break
		}
	}

	desc := filingDates("desc")
	for i := 1; i < len(desc); i++ {
		if desc[i] > desc[i-1] {
			t.Errorf("desc sort not non-increasing: %v", desc)
			break
		}
	}

	// The two orders must differ, proving the sort key took effect rather than
	// the API returning a single default ordering both times.
	if asc[0] == desc[0] && asc[len(asc)-1] == desc[len(desc)-1] {
		t.Errorf("asc and desc returned the same ordering (%v vs %v) - sort had no effect", asc, desc)
	}
}
