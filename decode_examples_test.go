package odp

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	generated "github.com/patent-dev/uspto-odp/generated"
)

// Strict completeness guard.
//
// Each fixture under testdata/strictdecode/ is a real USPTO ODP API response,
// captured live. We strict-decode each into the typed 200-response struct it
// maps to (via json.Decoder with DisallowUnknownFields). A failure means the
// API returns a field the generated/hand-written types do not model -- i.e. a
// silently dropped field. The fixtures cover the endpoints whose envelope and
// nested fields were recovered (requestIdentifier, count, directionCategory,
// PTA delay quantities, attorneyDocketNumber, petition decision/status codes,
// trial/interference metadata, OA /fields lastDataUpdatedDate).
//
// LIMITATION: a fixture only exercises the fields present in that one sampled
// response. A field the API can return but did not in the sample is not caught.
// The guard is a high-value smoke test against real data, not a proof of total
// completeness.
//
// The Office Action /records (search) endpoints are intentionally NOT covered:
// they use a dynamic Solr schema and are modelled as Docs []map[string]any by
// design (see DSAPIResponse), so DisallowUnknownFields does not apply.
func TestStrictDecodeExamples(t *testing.T) {
	// target returns a fresh pointer to decode into. Using a constructor keeps
	// each case's concrete type local; inline 200 schemas decode into the JSON200
	// field of their generated response wrapper.
	cases := []struct {
		file   string
		target func() any
	}{
		// Patent application + sub-resources.
		{"get_patent.json", func() any { return new(generated.PatentDataResponse) }},
		{"search_patents.json", func() any { return new(generated.PatentDataResponse) }},
		{"get_patent_adjustment.json", func() any {
			return &(&generated.GetApiV1PatentApplicationsApplicationNumberTextAdjustmentResponse{}).JSON200
		}},
		{"get_patent_assignment.json", func() any {
			return &(&generated.GetApiV1PatentApplicationsApplicationNumberTextAssignmentResponse{}).JSON200
		}},
		{"get_patent_documents.json", func() any { return new(generated.DocumentBag) }},
		{"get_patent_continuity.json", func() any {
			return &(&generated.GetApiV1PatentApplicationsApplicationNumberTextContinuityResponse{}).JSON200
		}},
		{"get_patent_foreign_priority.json", func() any {
			return &(&generated.GetApiV1PatentApplicationsApplicationNumberTextForeignPriorityResponse{}).JSON200
		}},
		{"get_patent_transactions.json", func() any {
			return &(&generated.GetApiV1PatentApplicationsApplicationNumberTextTransactionsResponse{}).JSON200
		}},
		{"get_patent_attorney.json", func() any {
			return &(&generated.GetApiV1PatentApplicationsApplicationNumberTextAttorneyResponse{}).JSON200
		}},
		{"get_patent_associated_documents.json", func() any {
			return &(&generated.GetApiV1PatentApplicationsApplicationNumberTextAssociatedDocumentsResponse{}).JSON200
		}},

		// PTAB trials: decisions, documents, proceedings.
		{"search_trial_decisions.json", func() any { return new(generated.DecisionDataResponse) }},
		{"get_trial_decisions.json", func() any { return new(generated.DecisionDataResponse) }},
		{"search_trial_documents.json", func() any { return new(generated.DocumentDataResponse) }},
		{"get_trial_documents.json", func() any { return new(generated.DocumentDataResponse) }},
		{"search_trial_proceedings.json", func() any { return new(generated.ProceedingDataResponse) }},

		// PTAB appeals + interferences.
		{"search_appeal_decisions.json", func() any { return new(generated.AppealDecisionDataResponse) }},
		{"get_appeal_decisions.json", func() any { return new(generated.AppealDecisionDataResponse) }},
		{"search_interference_decisions.json", func() any { return new(generated.InterferenceDecisionDataResponse) }},
		{"get_interference_decisions.json", func() any { return new(generated.InterferenceDecisionDataResponse) }},

		// Petitions.
		{"search_petitions.json", func() any { return new(generated.PetitionDecisionResponseBag) }},
		{"get_petition_decision.json", func() any { return new(generated.PetitionDecisionIdentifierResponseBag) }},

		// Office Action /fields (hand-written DSAPIFieldsResponse).
		{"get_office_action_fields.json", func() any { return new(DSAPIFieldsResponse) }},
	}

	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", "strictdecode", tc.file))
			if err != nil {
				t.Fatalf("reading fixture: %v", err)
			}
			dec := json.NewDecoder(bytes.NewReader(data))
			dec.DisallowUnknownFields()
			if err := dec.Decode(tc.target()); err != nil {
				t.Fatalf("strict decode dropped a field the API returns: %v", err)
			}
		})
	}
}
