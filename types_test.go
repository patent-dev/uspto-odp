package odp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// setupFixtureServer creates a mock server that serves a JSON fixture file at the given path.
func setupFixtureServer(t *testing.T, urlPath, fixturePath string) (*Client, func()) {
	t.Helper()
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Failed to read fixture %s: %v", fixturePath, err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == urlPath {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(data)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	config := DefaultConfig()
	config.BaseURL = server.URL
	config.APIKey = "test-key"
	client, err := NewClient(config)
	if err != nil {
		server.Close()
		t.Fatalf("Failed to create client: %v", err)
	}
	return client, server.Close
}

// setupEmptyServer creates a mock server returning an empty patentFileWrapperDataBag.
func setupEmptyServer(t *testing.T, urlPath string) (*Client, func()) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == urlPath {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"count":1,"patentFileWrapperDataBag":[{"applicationNumberText":"00000000"}]}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	config := DefaultConfig()
	config.BaseURL = server.URL
	config.APIKey = "test-key"
	client, err := NewClient(config)
	if err != nil {
		server.Close()
		t.Fatalf("Failed to create client: %v", err)
	}
	return client, server.Close
}

func TestGetPatentContinuity_Fixture(t *testing.T) {
	client, cleanup := setupFixtureServer(t,
		"/api/v1/patent/applications/17248024/continuity",
		"demo/examples/get_patent_continuity/response.json")
	defer cleanup()

	result, err := client.GetPatentContinuity(context.Background(), "17248024")
	if err != nil {
		t.Fatalf("GetPatentContinuity failed: %v", err)
	}
	if result.ApplicationNumber != "17248024" {
		t.Errorf("ApplicationNumber = %q, want %q", result.ApplicationNumber, "17248024")
	}
	if len(result.Parents) != 12 {
		t.Fatalf("Expected 12 parents, got %d", len(result.Parents))
	}
	// First parent
	p0 := result.Parents[0]
	if p0.ApplicationNumber != "16695054" {
		t.Errorf("Parent[0].ApplicationNumber = %q, want %q", p0.ApplicationNumber, "16695054")
	}
	if p0.PatentNumber != "10916753" {
		t.Errorf("Parent[0].PatentNumber = %q, want %q", p0.PatentNumber, "10916753")
	}
	if p0.FilingDate != "2019-11-25" {
		t.Errorf("Parent[0].FilingDate = %q, want %q", p0.FilingDate, "2019-11-25")
	}
	if p0.RelationshipType != "Continuation" {
		t.Errorf("Parent[0].RelationshipType = %q, want %q", p0.RelationshipType, "Continuation")
	}
	// Fifth parent (CIP)
	p4 := result.Parents[4]
	if p4.RelationshipType != "Continuation-in-part" {
		t.Errorf("Parent[4].RelationshipType = %q, want %q", p4.RelationshipType, "Continuation-in-part")
	}
	// No children in this fixture
	if len(result.Children) != 0 {
		t.Errorf("Expected 0 children, got %d", len(result.Children))
	}
	if result.Children == nil {
		t.Error("Children should be non-nil empty slice")
	}
}

func TestGetPatentContinuity_WithChildren(t *testing.T) {
	childJSON := `{"count":1,"patentFileWrapperDataBag":[{"applicationNumberText":"17000000","childContinuityBag":[{"childApplicationNumberText":"18111111","childPatentNumber":"11222333","childApplicationFilingDate":"2024-01-15","childApplicationStatusDescriptionText":"Patented Case","claimParentageTypeCode":"CON","claimParentageTypeCodeDescriptionText":"is a Continuation of"},{"childApplicationNumberText":"18222222","childApplicationFilingDate":"2024-06-01","childApplicationStatusDescriptionText":"Docketed New Case","claimParentageTypeCode":"DIV","claimParentageTypeCodeDescriptionText":"is a Division of"}]}]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(childJSON))
	}))
	defer server.Close()
	config := DefaultConfig()
	config.BaseURL = server.URL
	config.APIKey = "test-key"
	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	result, err := client.GetPatentContinuity(context.Background(), "17000000")
	if err != nil {
		t.Fatalf("GetPatentContinuity failed: %v", err)
	}
	if len(result.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(result.Children))
	}
	ch0 := result.Children[0]
	if ch0.ApplicationNumber != "18111111" {
		t.Errorf("Child[0].ApplicationNumber = %q, want %q", ch0.ApplicationNumber, "18111111")
	}
	if ch0.PatentNumber != "11222333" {
		t.Errorf("Child[0].PatentNumber = %q, want %q", ch0.PatentNumber, "11222333")
	}
	if ch0.RelationshipType != "Continuation" {
		t.Errorf("Child[0].RelationshipType = %q, want %q", ch0.RelationshipType, "Continuation")
	}
	ch1 := result.Children[1]
	if ch1.RelationshipType != "Division" {
		t.Errorf("Child[1].RelationshipType = %q, want %q", ch1.RelationshipType, "Division")
	}
	if ch1.Status != "Docketed New Case" {
		t.Errorf("Child[1].Status = %q, want %q", ch1.Status, "Docketed New Case")
	}
}

func TestGetPatentContinuity_Empty(t *testing.T) {
	client, cleanup := setupEmptyServer(t, "/api/v1/patent/applications/00000000/continuity")
	defer cleanup()

	result, err := client.GetPatentContinuity(context.Background(), "00000000")
	if err != nil {
		t.Fatalf("GetPatentContinuity failed: %v", err)
	}
	if result.Parents == nil {
		t.Error("Parents should be non-nil empty slice")
	}
	if len(result.Parents) != 0 {
		t.Errorf("Expected 0 parents, got %d", len(result.Parents))
	}
	if result.Children == nil {
		t.Error("Children should be non-nil empty slice")
	}
	if len(result.Children) != 0 {
		t.Errorf("Expected 0 children, got %d", len(result.Children))
	}
}

func TestGetPatentAssignment_Fixture(t *testing.T) {
	client, cleanup := setupFixtureServer(t,
		"/api/v1/patent/applications/15000001/assignment",
		"demo/examples/get_patent_assignment/response.json")
	defer cleanup()

	result, err := client.GetPatentAssignment(context.Background(), "15000001")
	if err != nil {
		t.Fatalf("GetPatentAssignment failed: %v", err)
	}
	if result.ApplicationNumber != "15000001" {
		t.Errorf("ApplicationNumber = %q, want %q", result.ApplicationNumber, "15000001")
	}
	if len(result.Assignments) != 1 {
		t.Fatalf("Expected 1 assignment, got %d", len(result.Assignments))
	}
	a := result.Assignments[0]
	if a.Assignee != "SAMSUNG ELECTRONICS CO., LTD" {
		t.Errorf("Assignee = %q, want %q", a.Assignee, "SAMSUNG ELECTRONICS CO., LTD")
	}
	if a.RecordedDate != "2016-04-19" {
		t.Errorf("RecordedDate = %q, want %q", a.RecordedDate, "2016-04-19")
	}
	if a.ExecutionDate != "2016-01-05" {
		t.Errorf("ExecutionDate = %q, want %q", a.ExecutionDate, "2016-01-05")
	}
	if a.Conveyance != "ASSIGNMENT OF ASSIGNORS INTEREST (SEE DOCUMENT FOR DETAILS)." {
		t.Errorf("Conveyance = %q", a.Conveyance)
	}
	if a.ReelFrame != "038323/0190" {
		t.Errorf("ReelFrame = %q, want %q", a.ReelFrame, "038323/0190")
	}
	// 9 assignors comma-joined
	if a.Assignor == "" {
		t.Error("Assignor should not be empty")
	}
	// Check first and last assignor names appear
	if !strings.Contains(a.Assignor, "HEO, JIN-PIL") {
		t.Errorf("Assignor should contain 'HEO, JIN-PIL', got %q", a.Assignor)
	}
	if !strings.Contains(a.Assignor, "YI, IN-SUN") {
		t.Errorf("Assignor should contain 'YI, IN-SUN', got %q", a.Assignor)
	}
}

func TestGetPatentAssignment_Empty(t *testing.T) {
	client, cleanup := setupEmptyServer(t, "/api/v1/patent/applications/00000000/assignment")
	defer cleanup()

	result, err := client.GetPatentAssignment(context.Background(), "00000000")
	if err != nil {
		t.Fatalf("GetPatentAssignment failed: %v", err)
	}
	if result.Assignments == nil {
		t.Error("Assignments should be non-nil empty slice")
	}
	if len(result.Assignments) != 0 {
		t.Errorf("Expected 0 assignments, got %d", len(result.Assignments))
	}
}

func TestGetPatentAdjustment_Fixture(t *testing.T) {
	client, cleanup := setupFixtureServer(t,
		"/api/v1/patent/applications/17248024/adjustment",
		"demo/examples/get_patent_adjustment/response.json")
	defer cleanup()

	result, err := client.GetPatentAdjustment(context.Background(), "17248024")
	if err != nil {
		t.Fatalf("GetPatentAdjustment failed: %v", err)
	}
	if result.ApplicationNumber != "17248024" {
		t.Errorf("ApplicationNumber = %q, want %q", result.ApplicationNumber, "17248024")
	}
	if result.TotalAdjustmentDays != 304 {
		t.Errorf("TotalAdjustmentDays = %d, want %d", result.TotalAdjustmentDays, 304)
	}
	if result.ADelays != 304 {
		t.Errorf("ADelays = %d, want %d", result.ADelays, 304)
	}
	if result.BDelays != 0 {
		t.Errorf("BDelays = %d, want %d", result.BDelays, 0)
	}
	if result.CDelays != 0 {
		t.Errorf("CDelays = %d, want %d", result.CDelays, 0)
	}
}

func TestGetPatentAdjustment_Empty(t *testing.T) {
	client, cleanup := setupEmptyServer(t, "/api/v1/patent/applications/00000000/adjustment")
	defer cleanup()

	result, err := client.GetPatentAdjustment(context.Background(), "00000000")
	if err != nil {
		t.Fatalf("GetPatentAdjustment failed: %v", err)
	}
	if result.TotalAdjustmentDays != 0 {
		t.Errorf("TotalAdjustmentDays = %d, want %d", result.TotalAdjustmentDays, 0)
	}
	if result.ADelays != 0 {
		t.Errorf("ADelays = %d, want %d", result.ADelays, 0)
	}
}

func TestGetPatentTransactions_Fixture(t *testing.T) {
	client, cleanup := setupFixtureServer(t,
		"/api/v1/patent/applications/17248024/transactions",
		"demo/examples/get_patent_transactions/response.json")
	defer cleanup()

	result, err := client.GetPatentTransactions(context.Background(), "17248024")
	if err != nil {
		t.Fatalf("GetPatentTransactions failed: %v", err)
	}
	if result.ApplicationNumber != "17248024" {
		t.Errorf("ApplicationNumber = %q, want %q", result.ApplicationNumber, "17248024")
	}
	if len(result.Events) < 40 {
		t.Errorf("Expected 40+ events, got %d", len(result.Events))
	}
	// First event
	e0 := result.Events[0]
	if e0.Code != "EML_NTR" {
		t.Errorf("Events[0].Code = %q, want %q", e0.Code, "EML_NTR")
	}
	if e0.Date != "2023-10-03" {
		t.Errorf("Events[0].Date = %q, want %q", e0.Date, "2023-10-03")
	}
	if e0.Description != "Email Notification" {
		t.Errorf("Events[0].Description = %q, want %q", e0.Description, "Email Notification")
	}
}

func TestGetPatentTransactions_Empty(t *testing.T) {
	client, cleanup := setupEmptyServer(t, "/api/v1/patent/applications/00000000/transactions")
	defer cleanup()

	result, err := client.GetPatentTransactions(context.Background(), "00000000")
	if err != nil {
		t.Fatalf("GetPatentTransactions failed: %v", err)
	}
	if result.Events == nil {
		t.Error("Events should be non-nil empty slice")
	}
	if len(result.Events) != 0 {
		t.Errorf("Expected 0 events, got %d", len(result.Events))
	}
}

func TestMapRelationshipType(t *testing.T) {
	tests := []struct {
		code, desc, want string
	}{
		{"CON", "is a Continuation of", "Continuation"},
		{"DIV", "is a Division of", "Division"},
		{"CIP", "is a Continuation in-part of", "Continuation-in-part"},
		{"PRO", "Claims Priority from Provisional Application", "Provisional"},
		{"XYZ", "Some unknown type", "Some unknown type"},
		{"UNK", "", "UNK"},
	}
	for _, tt := range tests {
		got := mapRelationshipType(tt.code, tt.desc)
		if got != tt.want {
			t.Errorf("mapRelationshipType(%q, %q) = %q, want %q", tt.code, tt.desc, got, tt.want)
		}
	}
}

func TestDerefStr(t *testing.T) {
	s := "  hello  "
	if got := derefStr(&s); got != "hello" {
		t.Errorf("derefStr(&%q) = %q, want %q", s, got, "hello")
	}
	if got := derefStr(nil); got != "" {
		t.Errorf("derefStr(nil) = %q, want %q", got, "")
	}
}

func TestDerefInt(t *testing.T) {
	n := 42
	if got := derefInt(&n); got != 42 {
		t.Errorf("derefInt(&42) = %d, want 42", got)
	}
	if got := derefInt(nil); got != 0 {
		t.Errorf("derefInt(nil) = %d, want 0", got)
	}
}
