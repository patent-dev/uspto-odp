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

// Assignor is a transferring party on an assignment recordation. The USPTO
// schema only carries name + execution date for assignors (no address).
type Assignor struct {
	Name          string
	ExecutionDate string
}

// Assignee is the receiving party on an assignment recordation. The USPTO
// schema carries the assignee's mailing address; ExecutionDate is not part
// of the assignee record.
type Assignee struct {
	Name             string
	City             string
	GeographicRegion string
	PostalCode       string
	CountryName      string
}

// AssignmentEntry represents a single patent assignment record.
type AssignmentEntry struct {
	Assignors           []Assignor
	Assignees           []Assignee
	RecordedDate        string
	Conveyance          string
	ReelFrame           string
	MailedDate          string
	ReceivedDate        string
	DocumentLocationURI string
}

// AssignmentResponse contains patent assignment/ownership data.
type AssignmentResponse struct {
	ApplicationNumber string
	Assignments       []AssignmentEntry
}

// AdjustmentResponse contains patent term adjustment data.
type AdjustmentResponse struct {
	ApplicationNumber   string
	TotalAdjustmentDays int
	ADelays             int
	BDelays             int
	CDelays             int
}

// CorrespondenceAddress represents a single mailing-address record on an
// applicant or inventor.
type CorrespondenceAddress struct {
	NameLineOne           string
	NameLineTwo           string
	CityName              string
	CountryCode           string
	CountryName           string
	GeographicRegionCode  string
	GeographicRegionName  string
	PostalAddressCategory string
}

// Applicant identifies a party listed as an applicant on the application.
type Applicant struct {
	ApplicantNameText        string
	FirstName                string
	MiddleName               string
	LastName                 string
	NamePrefix               string
	NameSuffix               string
	PreferredName            string
	CountryCode              string
	CorrespondenceAddressBag []CorrespondenceAddress
}

// Inventor identifies a party listed as an inventor on the application.
type Inventor struct {
	InventorNameText         string
	FirstName                string
	MiddleName               string
	LastName                 string
	NamePrefix               string
	NameSuffix               string
	PreferredName            string
	CountryCode              string
	CorrespondenceAddressBag []CorrespondenceAddress
}

// EntityStatus indicates the applicant's fee-entity status.
type EntityStatus struct {
	BusinessEntityStatusCategory string
	SmallEntityStatusIndicator   *bool
}

// MetaDataResponse contains the full patent application meta-data response.
// MetaDataResponse pointer fields use *T (rather than T) when the API
// distinguishes "field absent" from the zero value -- for status codes,
// confirmation numbers, indicators. nil means USPTO did not return the
// field for this application.
type MetaDataResponse struct {
	ApplicationNumber                        string
	ApplicationConfirmationNumber            *int
	PatentNumber                             string
	InventionTitle                           string
	FilingDate                               string
	GrantDate                                string
	EffectiveFilingDate                      string
	EarliestPublicationNumber                string
	EarliestPublicationDate                  string
	ApplicationStatusCode                    *int
	ApplicationStatusDescriptionText         string
	ApplicationStatusDate                    string
	ApplicationTypeCategory                  string
	ApplicationTypeCode                      string
	ApplicationTypeLabelName                 string
	ExaminerNameText                         string
	GroupArtUnitNumber                       string
	DocketNumber                             string
	CustomerNumber                           *int
	FirstApplicantName                       string
	FirstInventorName                        string
	FirstInventorToFileIndicator             string
	NationalStageIndicator                   *bool
	PctPublicationNumber                     string
	PctPublicationDate                       string
	InternationalRegistrationNumber          string
	InternationalRegistrationPublicationDate string
	CpcClassificationBag                     []string
	PublicationCategoryBag                   []string
	PublicationDateBag                       []string
	PublicationSequenceNumberBag             []string
	UspcSymbolText                           string
	Class                                    string
	Subclass                                 string
	EntityStatus                             *EntityStatus
	Applicants                               []Applicant
	Inventors                                []Inventor
}

// ForeignPriorityClaim represents a single foreign priority claim.
type ForeignPriorityClaim struct {
	ApplicationNumber string
	FilingDate        string
	IPOfficeName      string
}

// ForeignPriorityResponse contains foreign priority claims for an application.
type ForeignPriorityResponse struct {
	ApplicationNumber string
	Claims            []ForeignPriorityClaim
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

// float32ToIntPtr converts a *float32 to *int, returning nil if the source
// is nil. Used to coerce upstream-typed numeric fields that USPTO defines
// as `number` in the swagger but populates with whole-integer values.
func float32ToIntPtr(p *float32) *int {
	if p == nil {
		return nil
	}
	v := int(*p)
	return &v
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

// mapRelationshipType converts a claim parentage type code to human-readable
// form. Best-effort: the four common codes (CON, DIV, CIP, PRO) get a
// canonical English label; other codes (REI, REX, NST, historical codes)
// pass through the API's description text and finally the raw code. Prefer
// the description text when available.
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
