package odp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A bare grant number without a kind code (e.g. "US11646472") is genuinely ambiguous:
// 11,646,472 is a valid grant number AND 11646472 is a valid application number that
// belongs to an unrelated patent. The old resolver defaulted to the application lookup,
// which 404'd for some numbers (US10000000) and - worse - silently returned a wrong
// patent for others (US11646472 -> a 2006 head-tracking application).
//
// The resolver must instead:
//   - auto-resolve when only one interpretation exists, and
//   - report both candidates (number + title) when a grant and a different application
//     both exist, so the caller can present a "did you mean" choice.

// ambiguityMock serves the two endpoints the resolver probes: the grant search and the
// direct application lookup. grantApp/grantTitle describe the grant hit (empty grantApp
// => no grant match -> 404). appExists/appTitle describe the direct application hit.
type ambiguityMock struct {
	grantApp, grantTitle string
	appExists            bool
	appTitle             string
	appNumber            string
}

func newAmbiguityClient(t *testing.T, m ambiguityMock) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v1/patent/applications/search":
			body, _ := io.ReadAll(r.Body)
			if m.grantApp == "" || !strings.Contains(string(body), "patentNumber") {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"no results"}`))
				return
			}
			writeWrapperBag(w, m.grantApp, m.grantTitle)
		case strings.HasPrefix(r.URL.Path, "/api/v1/patent/applications/"):
			if !m.appExists {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"not found"}`))
				return
			}
			writeWrapperBag(w, m.appNumber, m.appTitle)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	cfg := DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.APIKey = "test"
	cfg.MaxRetries = 0
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

func writeWrapperBag(w http.ResponseWriter, appNumber, title string) {
	resp := map[string]any{
		"count": 1,
		"patentFileWrapperDataBag": []any{
			map[string]any{
				"applicationNumberText": appNumber,
				"applicationMetaData": map[string]any{
					"inventionTitle": title,
				},
			},
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func TestNormalizePatentNumber_AmbiguousFlag(t *testing.T) {
	tests := []struct {
		input         string
		wantAmbiguous bool
		wantType      PatentNumberType
	}{
		{"11646472", true, PatentNumberTypeApplication},    // grant 11,646,472 vs app series 11
		{"US11646472", true, PatentNumberTypeApplication},  // same, with US prefix
		{"10000000", true, PatentNumberTypeApplication},    // first 8-digit grant
		{"12999999", true, PatentNumberTypeApplication},    // top of the grant range
		{"13000000", false, PatentNumberTypeApplication},   // app series 13: not yet a grant
		{"14643719", false, PatentNumberTypeApplication},   // Raytheon application
		{"17248024", false, PatentNumberTypeApplication},   // PolyPlus application
		{"17/248,024", false, PatentNumberTypeApplication}, // slash => unambiguous app
		{"US 11,646,472 B2", false, PatentNumberTypeGrant}, // kind code => unambiguous grant
		{"9123456", false, PatentNumberTypeGrant},          // 7-digit grant
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			pn, err := NormalizePatentNumber(tt.input)
			if err != nil {
				t.Fatalf("NormalizePatentNumber(%q): %v", tt.input, err)
			}
			if pn.Ambiguous != tt.wantAmbiguous {
				t.Errorf("Ambiguous = %v, want %v", pn.Ambiguous, tt.wantAmbiguous)
			}
			if pn.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", pn.Type, tt.wantType)
			}
		})
	}
}

func TestResolvePatentNumber_Ambiguous_ReportsBothCandidates(t *testing.T) {
	client := newAmbiguityClient(t, ambiguityMock{
		grantApp:   "17248024",
		grantTitle: "MAKING LITHIUM METAL - SEAWATER BATTERY CELLS HAVING PROTECTED LITHIUM ELECTRODES",
		appExists:  true,
		appNumber:  "11646472",
		appTitle:   "METHOD AND APPARATUS FOR TRACKING LISTENER'S HEAD POSITION FOR VIRTUAL STEREO ACOUSTICS",
	})

	_, err := client.ResolvePatentNumber(context.Background(), "US11646472")
	if err == nil {
		t.Fatal("expected an ambiguity error, got nil")
	}
	var ambErr *AmbiguousPatentNumberError
	if !errors.As(err, &ambErr) {
		t.Fatalf("expected *AmbiguousPatentNumberError, got %T: %v", err, err)
	}
	if len(ambErr.Candidates) != 2 {
		t.Fatalf("want 2 candidates, got %d: %+v", len(ambErr.Candidates), ambErr.Candidates)
	}

	var gotGrant, gotApp bool
	for _, c := range ambErr.Candidates {
		switch c.Type {
		case PatentNumberTypeGrant:
			gotGrant = true
			if c.ApplicationNumber != "17248024" || !strings.Contains(c.Title, "LITHIUM") {
				t.Errorf("grant candidate wrong: %+v", c)
			}
		case PatentNumberTypeApplication:
			gotApp = true
			if c.ApplicationNumber != "11646472" || !strings.Contains(c.Title, "HEAD POSITION") {
				t.Errorf("application candidate wrong: %+v", c)
			}
		}
	}
	if !gotGrant || !gotApp {
		t.Errorf("expected one grant and one application candidate, got %+v", ambErr.Candidates)
	}
}

func TestResolvePatentNumber_BareGrantOnly_AutoResolves(t *testing.T) {
	// US10000000: grant exists, no application 10000000 -> auto-resolve as grant.
	client := newAmbiguityClient(t, ambiguityMock{
		grantApp:   "14643719",
		grantTitle: "COHERENT LADAR USING INTRA-PIXEL QUADRATURE DETECTION",
		appExists:  false,
	})

	app, err := client.ResolvePatentNumber(context.Background(), "US10000000")
	if err != nil {
		t.Fatalf("expected auto-resolution, got error: %v", err)
	}
	if app != "14643719" {
		t.Errorf("want application 14643719, got %q", app)
	}
}

// An application number produced by a prior grant lookup can itself be an 8-digit value
// in the grant range that collides with a different grant (e.g. application 11616583 vs
// grant 11,616,583). Fetch methods must resolve it leniently - taking it as the
// application number - rather than re-probing and raising a spurious ambiguity error.
func TestResolveApplicationNumberLenient_DoesNotProbe(t *testing.T) {
	client := newAmbiguityClient(t, ambiguityMock{
		grantApp:   "16984686", // grant 11,616,583 maps to a different application
		grantTitle: "Seamless Integration of Radio Broadcast Audio with Streaming Audio",
		appExists:  true,
		appNumber:  "11616583",
		appTitle:   "FORMATION FLUID SAMPLING APPARATUS AND METHODS",
	})

	app, err := client.resolveApplicationNumberLenient(context.Background(), "11616583")
	if err != nil {
		t.Fatalf("lenient resolve must not error on a colliding application number: %v", err)
	}
	if app != "11616583" {
		t.Errorf("want application 11616583, got %q", app)
	}
}

func TestGetPatent_DoesNotReprobeApplicationNumber(t *testing.T) {
	client := newAmbiguityClient(t, ambiguityMock{
		grantApp:   "16984686",
		grantTitle: "Seamless Integration of Radio Broadcast Audio with Streaming Audio",
		appExists:  true,
		appNumber:  "11616583",
		appTitle:   "FORMATION FLUID SAMPLING APPARATUS AND METHODS",
	})

	resp, err := client.GetPatent(context.Background(), "11616583")
	if err != nil {
		t.Fatalf("GetPatent must not raise ambiguity on an application number: %v", err)
	}
	if resp == nil || resp.PatentFileWrapperDataBag == nil || len(*resp.PatentFileWrapperDataBag) == 0 {
		t.Fatal("GetPatent returned no patent data")
	}
	if got := (*resp.PatentFileWrapperDataBag)[0].ApplicationNumberText; got == nil || *got != "11616583" {
		t.Errorf("GetPatent resolved to the wrong application: %v", got)
	}
}

func TestResolvePatentNumber_BareApplicationOnly_AutoResolves(t *testing.T) {
	// 14643719: no grant has that number, but the application exists -> resolve as app.
	client := newAmbiguityClient(t, ambiguityMock{
		grantApp:  "", // grant search 404s
		appExists: true,
		appNumber: "14643719",
		appTitle:  "COHERENT LADAR USING INTRA-PIXEL QUADRATURE DETECTION",
	})

	app, err := client.ResolvePatentNumber(context.Background(), "14643719")
	if err != nil {
		t.Fatalf("expected auto-resolution, got error: %v", err)
	}
	if app != "14643719" {
		t.Errorf("want application 14643719, got %q", app)
	}
}
