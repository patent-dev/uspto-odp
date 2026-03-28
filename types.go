package odp

import (
	"strings"

	"github.com/patent-dev/uspto-odp/generated"
)

// AssociatedDocumentsResponse contains patent grant and publication XML file metadata.
type AssociatedDocumentsResponse struct {
	ApplicationNumber     string
	GrantDocumentMetaData *generated.GrantFileMetaData
	PgpubDocumentMetaData *generated.PGPubFileMetaData
}

// ContinuityParent represents a parent application in a continuity chain.
type ContinuityParent struct {
	ApplicationNumber string
	PatentNumber      string
	FilingDate        string
	Status            string
	RelationshipType  string
}

// ContinuityChild represents a child application in a continuity chain.
type ContinuityChild struct {
	ApplicationNumber string
	PatentNumber      string
	FilingDate        string
	Status            string
	RelationshipType  string
}

// ContinuityResponse contains patent continuity data (parent/child application chain).
type ContinuityResponse struct {
	ApplicationNumber string
	Parents           []ContinuityParent
	Children          []ContinuityChild
}

// AssignmentEntry represents a single patent assignment record.
type AssignmentEntry struct {
	Assignor      string
	Assignee      string
	RecordedDate  string
	ExecutionDate string
	Conveyance    string
	ReelFrame     string
}

// AssignmentResponse contains patent assignment/ownership data.
type AssignmentResponse struct {
	ApplicationNumber string
	Assignments       []AssignmentEntry
}

// AdjustmentResponse contains patent term adjustment data.
// Note: FilingDate and GrantDate are not available from the adjustment endpoint;
// they require a separate GetPatentMetaData call to populate.
type AdjustmentResponse struct {
	ApplicationNumber   string
	TotalAdjustmentDays int
	ADelays             int
	BDelays             int
	CDelays             int
	FilingDate          string
	GrantDate           string
}

// TransactionEvent represents a single patent transaction event.
type TransactionEvent struct {
	Date        string
	Code        string
	Description string
}

// TransactionsResponse contains patent transaction history.
type TransactionsResponse struct {
	ApplicationNumber string
	Events            []TransactionEvent
}

// derefStr safely dereferences a *string pointer, returning "" if nil.
func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(*p)
}

// derefInt safely dereferences an *int pointer, returning 0 if nil.
func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// mapRelationshipType converts a claim parentage type code to human-readable form.
func mapRelationshipType(code, description string) string {
	switch code {
	case "CON":
		return "Continuation"
	case "DIV":
		return "Division"
	case "CIP":
		return "Continuation-in-part"
	case "PRO":
		return "Provisional"
	default:
		if description != "" {
			return description
		}
		return code
	}
}
