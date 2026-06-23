package odp

import "testing"

func TestBuildSearchFilters(t *testing.T) {
	// Empty field or no values -> skipped; the rest pass through with copied values.
	in := []PatentSearchFilter{
		{Field: "", Values: []string{"x"}},
		{Field: "status", Values: nil},
		{Field: "applicationMetaData.applicationTypeLabelName", Values: []string{"Utility", "Design"}},
	}
	out := buildSearchFilters(in)
	if len(out) != 1 {
		t.Fatalf("filters = %d, want 1 (two skipped)", len(out))
	}
	if out[0].Name == nil || *out[0].Name != "applicationMetaData.applicationTypeLabelName" {
		t.Errorf("filter name = %v", out[0].Name)
	}
	if out[0].Value == nil || len(*out[0].Value) != 2 || (*out[0].Value)[0] != "Utility" {
		t.Errorf("filter values = %v, want [Utility Design]", out[0].Value)
	}

	// All-skipped input yields an empty (len 0) slice so the caller leaves the
	// request field nil.
	if got := buildSearchFilters([]PatentSearchFilter{{Field: ""}}); len(got) != 0 {
		t.Errorf("all-skipped filters = %v, want empty", got)
	}
}

func TestBuildSearchRanges(t *testing.T) {
	in := []PatentSearchRange{
		{Field: "", From: "a", To: "b"},                                  // no field -> skip
		{Field: "applicationMetaData.filingDate", From: "", To: ""},      // no bound -> skip
		{Field: "applicationMetaData.filingDate", From: "2022-01-01"},    // open upper bound
		{Field: "applicationMetaData.grantDate", To: "2023-12-31"},       // open lower bound
	}
	out := buildSearchRanges(in)
	if len(out) != 2 {
		t.Fatalf("ranges = %d, want 2 (two skipped)", len(out))
	}
	// Open upper bound: ValueFrom set, ValueTo nil.
	if out[0].ValueFrom == nil || *out[0].ValueFrom != "2022-01-01" || out[0].ValueTo != nil {
		t.Errorf("range[0] = %+v, want From=2022-01-01 To=nil", out[0])
	}
	// Open lower bound: ValueTo set, ValueFrom nil.
	if out[1].ValueTo == nil || *out[1].ValueTo != "2023-12-31" || out[1].ValueFrom != nil {
		t.Errorf("range[1] = %+v, want From=nil To=2023-12-31", out[1])
	}
}
