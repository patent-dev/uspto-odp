package odp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/patent-dev/uspto-odp/generated"
)

// TestDefaultConfig tests the DefaultConfig function
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}
	if config.BaseURL != "https://api.uspto.gov" {
		t.Errorf("Expected BaseURL to be https://api.uspto.gov, got %s", config.BaseURL)
	}
	if config.UserAgent != "PatentDev/1.0" {
		t.Errorf("Expected UserAgent to be PatentDev/1.0, got %s", config.UserAgent)
	}
	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries to be 3, got %d", config.MaxRetries)
	}
	if config.RetryDelay != 1 {
		t.Errorf("Expected RetryDelay to be 1, got %d", config.RetryDelay)
	}
	if config.Timeout != 30 {
		t.Errorf("Expected Timeout to be 30, got %d", config.Timeout)
	}
}

// TestClientWithActualResponses tests all client methods with actual API response structures
func TestClientWithActualResponses(t *testing.T) {
	// Create mock server with actual response structures
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/patent/applications/search":
			// Actual response from SearchPatents
			response := map[string]interface{}{
				"count": 32783,
				"patentFileWrapperDataBag": []interface{}{
					map[string]interface{}{
						"eventDataBag": []interface{}{
							map[string]interface{}{
								"eventCode":            "EML_NTR",
								"eventDescriptionText": "Email Notification",
								"eventDate":            "2025-09-11",
							},
							map[string]interface{}{
								"eventCode":            "PGPC",
								"eventDescriptionText": "Sent to Classification Contractor",
								"eventDate":            "2025-09-10",
							},
						},
						"applicationMetaData": map[string]interface{}{
							"firstInventorToFileIndicator": "Y",
							"applicationStatusCode":        17,
							"applicationTypeCode":          "UTL",
							"entityStatusData": map[string]interface{}{
								"smallEntityStatusIndicator":   false,
								"businessEntityStatusCategory": "Small",
							},
							"filingDate": "2025-08-19",
							"inventorBag": []interface{}{
								map[string]interface{}{
									"firstName":        "Hiroshi",
									"lastName":         "ASAHARA",
									"inventorNameText": "Hiroshi ASAHARA",
									"correspondenceAddressBag": []interface{}{
										map[string]interface{}{
											"cityName":              "Tokyo",
											"countryCode":           "JP",
											"nameLineOneText":       "ASAHARA,Hiroshi",
											"countryName":           "JAPAN",
											"postalAddressCategory": "postal",
										},
										map[string]interface{}{
											"cityName":              "Tokyo",
											"countryCode":           "JP",
											"nameLineOneText":       "ASAHARA,Hiroshi",
											"countryName":           "JAPAN",
											"postalAddressCategory": "residence",
										},
									},
								},
								map[string]interface{}{
									"firstName":        "Hiroki",
									"lastName":         "TSUTSUMI",
									"inventorNameText": "Hiroki TSUTSUMI",
									"correspondenceAddressBag": []interface{}{
										map[string]interface{}{
											"cityName":              "Tokyo",
											"countryCode":           "JP",
											"nameLineOneText":       "TSUTSUMI,Hiroki",
											"countryName":           "JAPAN",
											"postalAddressCategory": "postal",
										},
										map[string]interface{}{
											"cityName":              "Tokyo",
											"countryCode":           "JP",
											"nameLineOneText":       "TSUTSUMI,Hiroki",
											"countryName":           "JAPAN",
											"postalAddressCategory": "residence",
										},
									},
								},
							},
							"applicationStatusDescriptionText": "Sent to Classification contractor",
							"applicantBag": []interface{}{
								map[string]interface{}{
									"applicantNameText": "INSTITUTE OF SCIENCE TOKYO",
									"correspondenceAddressBag": []interface{}{
										map[string]interface{}{
											"cityName":              "Tokyo",
											"countryCode":           "JP",
											"nameLineOneText":       "INSTITUTE OF SCIENCE TOKYO",
											"countryName":           "JAPAN",
											"postalAddressCategory": "postal",
										},
									},
								},
							},
							"firstApplicantName":            "INSTITUTE OF SCIENCE TOKYO",
							"customerNumber":                20350,
							"inventionTitle":                "TENDON/LIGAMENT-LIKE ARTIFICIAL TISSUE PRODUCED USING THREE-DIMENSIONAL MECHANOSIGNALING CELL CULTURE SYSTEM",
							"nationalStageIndicator":        true,
							"firstInventorName":             "Hiroshi ASAHARA",
							"applicationConfirmationNumber": 8110,
							"effectiveFilingDate":           "2024-11-05",
							"applicationTypeLabelName":      "Utility",
							"publicationCategoryBag":        []string{"Other"},
							"applicationStatusDate":         "2025-09-10",
							"docketNumber":                  "116514-1473401-100US",
							"applicationTypeCategory":       "REGULAR",
						},
						"parentContinuityBag": []interface{}{
							map[string]interface{}{
								"parentApplicationStatusCode":            17,
								"claimParentageTypeCode":                 "NST",
								"claimParentageTypeCodeDescriptionText":  "is a National Stage Entry of",
								"parentApplicationStatusDescriptionText": "Sent to Classification contractor",
								"parentApplicationNumberText":            "PCTJP2023017788",
								"parentApplicationFilingDate":            "2023-05-11",
								"childApplicationNumberText":             "18863279",
							},
						},
						"lastIngestionDateTime": "2025-09-11T05:00:10",
						"recordAttorney": map[string]interface{}{
							"customerNumberCorrespondenceData": map[string]interface{}{
								"powerOfAttorneyAddressBag": []interface{}{
									map[string]interface{}{
										"cityName":             "Atlanta",
										"geographicRegionName": "GEORGIA",
										"geographicRegionCode": "GA",
										"countryCode":          "US",
										"postalCode":           "30309",
										"nameLineOneText":      "Kilpatrick Townsend & Stockton LLP - West Coast",
										"nameLineTwoText":      "Mailstop: IP Docketing - 22",
										"countryName":          "UNITED STATES",
										"addressLineOneText":   "1100 Peachtree Street",
										"addressLineTwoText":   "Suite 2800",
									},
								},
								"patronIdentifier": 20350,
							},
							"powerOfAttorneyBag": []interface{}{
								map[string]interface{}{
									"activeIndicator":    "ACTIVE",
									"firstName":          "MIKA",
									"lastName":           "ITO",
									"registrationNumber": "71201",
									"attorneyAddressBag": []interface{}{
										map[string]interface{}{
											"cityName":             "SEATTLE",
											"geographicRegionName": "WASHINGTON",
											"geographicRegionCode": "WA",
											"countryCode":          "US",
											"postalCode":           "98101",
											"nameLineOneText":      "KILPATRICK TOWNSEND & STOCKTON LLP",
											"countryName":          "UNITED STATES",
											"addressLineOneText":   "1420 FIFTH AVENUE, SUITE 3700",
										},
									},
									"telecommunicationAddressBag": []interface{}{
										map[string]interface{}{
											"telecommunicationNumber": "206-626-7712",
											"telecomTypeCode":         "TEL",
										},
									},
									"registeredPractitionerCategory": "ATTNY",
								},
								map[string]interface{}{
									"activeIndicator":    "ACTIVE",
									"firstName":          "TYWANDA",
									"lastName":           "HARRIS",
									"registrationNumber": "46758",
									"attorneyAddressBag": []interface{}{
										map[string]interface{}{
											"cityName":             "ATLANTA",
											"geographicRegionName": "GEORGIA",
											"geographicRegionCode": "GA",
											"countryCode":          "US",
											"postalCode":           "30309",
											"nameLineOneText":      "KILPATRICK TOWNSEND & STOCKTON LLP",
											"countryName":          "UNITED STATES",
											"addressLineOneText":   "1100 PEACHTREE ST, SUITE 2800",
										},
									},
									"telecommunicationAddressBag": []interface{}{
										map[string]interface{}{
											"telecommunicationNumber": "404-745-2597",
											"telecomTypeCode":         "TEL",
										},
									},
									"registeredPractitionerCategory": "ATTNY",
								},
							},
							"attorneyBag": []interface{}{
								map[string]interface{}{
									"activeIndicator":    "ACTIVE",
									"firstName":          "JOHN",
									"lastName":           "PRATT",
									"registrationNumber": "29476",
									"attorneyAddressBag": []interface{}{
										map[string]interface{}{
											"cityName":             "ATLANTA",
											"geographicRegionName": "GEORGIA",
											"geographicRegionCode": "GA",
											"countryCode":          "US",
											"postalCode":           "30309-452",
											"nameLineOneText":      "KILAPTRICK TOWNSEND & STOCKTON, LLP",
											"countryName":          "UNITED STATES",
											"addressLineOneText":   "1100 PEACHTREE ST, STE 2800",
										},
									},
									"telecommunicationAddressBag": []interface{}{
										map[string]interface{}{
											"telecommunicationNumber": "404-815-6500",
											"telecomTypeCode":         "TEL",
										},
									},
									"registeredPractitionerCategory": "ATTNY",
								},
								map[string]interface{}{
									"activeIndicator":    "ACTIVE",
									"firstName":          "THEODORE",
									"lastName":           "BROWN",
									"registrationNumber": "31741",
									"attorneyAddressBag": []interface{}{
										map[string]interface{}{
											"cityName":             "MENLO PARK",
											"geographicRegionName": "CALIFORNIA",
											"geographicRegionCode": "CA",
											"countryCode":          "US",
											"postalCode":           "94025",
											"nameLineOneText":      "KILPATRICK TOWNSEND & STOCKTON, LLP",
											"countryName":          "UNITED STATES",
											"addressLineOneText":   "1302 EL CAMINO REAL",
											"addressLineTwoText":   "SUITE 175",
										},
									},
									"nameSuffix": "III",
									"telecommunicationAddressBag": []interface{}{
										map[string]interface{}{
											"telecommunicationNumber": "650-326-2400",
											"telecomTypeCode":         "TEL",
										},
									},
									"registeredPractitionerCategory": "ATTNY",
								},
							},
						},
						"applicationNumberText": "18863279",
						"correspondenceAddressBag": []interface{}{
							map[string]interface{}{
								"cityName":             "Atlanta",
								"geographicRegionName": "GEORGIA",
								"geographicRegionCode": "GA",
								"countryCode":          "US",
								"postalCode":           "30309",
								"nameLineOneText":      "Kilpatrick Townsend & Stockton LLP - West Coast",
								"nameLineTwoText":      "Mailstop: IP Docketing - 22",
								"countryName":          "UNITED STATES",
								"addressLineOneText":   "1100 Peachtree Street",
								"addressLineTwoText":   "Suite 2800",
							},
						},
						"foreignPriorityBag": []interface{}{
							map[string]interface{}{
								"filingDate":            "2022-05-12",
								"applicationNumberText": "2022-078641",
								"ipOfficeName":          "JAPAN",
							},
						},
					},
				},
				"requestIdentifier": "3e733c0b-db90-4b4f-827d-4c405a928ecc",
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/patent/applications/17123456":
			// Actual response from GetPatent
			response := map[string]interface{}{
				"count": 1,
				"patentFileWrapperDataBag": []interface{}{
					map[string]interface{}{
						"grantDocumentMetaData": map[string]interface{}{
							"productIdentifier":  "PTGRXML",
							"zipFileName":        "ipg230711.zip",
							"fileCreateDateTime": "2024-09-30T17:11:29",
							"xmlFileName":        "17123456_11696966.xml",
							"fileLocationURI":    "https://api.uspto.gov/api/v1/datasets/products/files/PTGRXML-SPLT/2023/ipg230711/17123456_11696966.xml",
						},
						"eventDataBag": []interface{}{
							map[string]interface{}{
								"eventCode":            "ELC_RVW",
								"eventDescriptionText": "Electronic Review",
								"eventDate":            "2023-07-11",
							},
							map[string]interface{}{
								"eventCode":            "ELC_RVW",
								"eventDescriptionText": "Electronic Review",
								"eventDate":            "2023-07-11",
							},
						},
						"applicationMetaData": map[string]interface{}{
							"firstInventorToFileIndicator": "Y",
							"applicationStatusCode":        150,
							"applicationTypeCode":          "UTL",
							"entityStatusData": map[string]interface{}{
								"businessEntityStatusCategory": "Micro",
							},
							"filingDate":               "2020-12-16",
							"uspcSymbolText":           "422/292",
							"nationalStageIndicator":   false,
							"firstInventorName":        "Darcy Jackson",
							"cpcClassificationBag":     []string{"A61L   2/22", "A61L2101/32"},
							"effectiveFilingDate":      "2020-12-16",
							"applicationTypeLabelName": "Utility",
							"applicationStatusDate":    "2023-06-21",
							"class":                    "422",
							"applicationTypeCategory":  "REGULAR",
							"inventorBag": []interface{}{
								map[string]interface{}{
									"firstName":        "Darcy",
									"lastName":         "Jackson",
									"inventorNameText": "Darcy Jackson",
									"correspondenceAddressBag": []interface{}{
										map[string]interface{}{
											"cityName":              "Fort Lauderdale",
											"geographicRegionName":  "FLORIDA",
											"geographicRegionCode":  "FL",
											"countryCode":           "US",
											"nameLineOneText":       "Jackson, Darcy",
											"countryName":           "UNITED STATES",
											"postalAddressCategory": "residence",
										},
										map[string]interface{}{
											"cityName":              "Fort Lauderdale",
											"geographicRegionName":  "FLORIDA",
											"geographicRegionCode":  "FL",
											"countryCode":           "US",
											"nameLineOneText":       "Darcy  Jackson",
											"countryName":           "UNITED STATES",
											"postalAddressCategory": "postal",
										},
									},
								},
								map[string]interface{}{
									"firstName":        "Darryl",
									"lastName":         "Rhue",
									"inventorNameText": "Darryl Rhue",
									"correspondenceAddressBag": []interface{}{
										map[string]interface{}{
											"cityName":              "Fort Lauderdale",
											"geographicRegionName":  "FLORIDA",
											"geographicRegionCode":  "FL",
											"countryCode":           "US",
											"nameLineOneText":       "Rhue, Darryl",
											"countryName":           "UNITED STATES",
											"postalAddressCategory": "residence",
										},
										map[string]interface{}{
											"cityName":              "Fort Lauderdale",
											"geographicRegionName":  "FLORIDA",
											"geographicRegionCode":  "FL",
											"countryCode":           "US",
											"nameLineOneText":       "Darryl  Rhue",
											"countryName":           "UNITED STATES",
											"postalAddressCategory": "postal",
										},
									},
								},
							},
							"applicationStatusDescriptionText": "Patented Case",
							"patentNumber":                     "11696966",
							"grantDate":                        "2023-07-11",
							"customerNumber":                   1818,
							"groupArtUnitNumber":               "1759",
							"inventionTitle":                   "DISINFECTANT FOG DISPENSER SYSTEM",
							"applicationConfirmationNumber":    9000,
							"examinerNameText":                 "PEREZ, JELITZA M",
							"subclass":                         "292",
							"publicationCategoryBag":           []string{"Granted/Issued"},
							"docketNumber":                     "430114",
						},
						"patentTermAdjustmentData": map[string]interface{}{
							"applicantDayDelayQuantity":       0,
							"overlappingDayQuantity":          0,
							"ipOfficeAdjustmentDelayQuantity": 0,
							"cDelayQuantity":                  0,
							"adjustmentTotalQuantity":         238,
							"bDelayQuantity":                  0,
							"nonOverlappingDayDelayQuantity":  238,
							"aDelayQuantity":                  238,
							"patentTermAdjustmentHistoryDataBag": []interface{}{
								map[string]interface{}{
									"applicantDayDelayQuantity":      0,
									"eventDescriptionText":           "PTA 36 Months",
									"eventSequenceNumber":            64.5,
									"originatingEventSequenceNumber": 0.5,
									"ptaPTECode":                     "PTA",
									"ipOfficeDayDelayQuantity":       0,
									"eventDate":                      "2023-07-11",
								},
								map[string]interface{}{
									"applicantDayDelayQuantity":      0,
									"eventDescriptionText":           "Patent Issue Date Used in PTA Calculation",
									"eventSequenceNumber":            64.0,
									"originatingEventSequenceNumber": 0.0,
									"ptaPTECode":                     "PTA",
									"ipOfficeDayDelayQuantity":       0,
									"eventDate":                      "2023-07-11",
								},
							},
						},
						"lastIngestionDateTime": "2025-07-18T22:45:35",
						"childContinuityBag": []interface{}{
							map[string]interface{}{
								"childApplicationStatusDescriptionText": "Application Undergoing Preexam Processing",
								"claimParentageTypeCode":                "?",
								"childApplicationStatusCode":            19,
								"claimParentageTypeCodeDescriptionText": "is a no data of",
								"parentApplicationNumberText":           "17123456",
								"childApplicationNumberText":            "PCTUS2297200",
							},
						},
						"recordAttorney": map[string]interface{}{
							"customerNumberCorrespondenceData": map[string]interface{}{
								"powerOfAttorneyAddressBag": []interface{}{
									map[string]interface{}{
										"cityName":             "MIAMI",
										"geographicRegionName": "FLORIDA",
										"geographicRegionCode": "FL",
										"countryCode":          "US",
										"postalCode":           "33134",
										"nameLineOneText":      "SANCHELIMA & ASSOCIATES, P.A.",
										"nameLineTwoText":      "CHRIS SANCHELIMA, ESQ.",
										"countryName":          "UNITED STATES",
										"addressLineOneText":   "235 S.W. LE JEUNE ROAD",
									},
								},
								"patronIdentifier": 1818,
							},
							"powerOfAttorneyBag": []interface{}{
								map[string]interface{}{
									"activeIndicator":    "ACTIVE",
									"firstName":          "CHRISTIAN",
									"lastName":           "SANCHELIMA",
									"registrationNumber": "70150",
									"attorneyAddressBag": []interface{}{
										map[string]interface{}{
											"cityName":             "MIAMI",
											"geographicRegionName": "FLORIDA",
											"geographicRegionCode": "FL",
											"countryCode":          "US",
											"postalCode":           "33134",
											"nameLineOneText":      "SANCHELIMA & ASSOCIATES, P.A.",
											"countryName":          "UNITED STATES",
											"addressLineOneText":   "235 SW 42ND AVENUE",
										},
									},
									"telecommunicationAddressBag": []interface{}{
										map[string]interface{}{
											"telecommunicationNumber": "305-447-1617",
											"telecomTypeCode":         "TEL",
										},
									},
									"registeredPractitionerCategory": "AGENT",
								},
								map[string]interface{}{
									"activeIndicator":    "ACTIVE",
									"firstName":          "JESUS",
									"lastName":           "SANCHELIMA",
									"registrationNumber": "28755",
									"attorneyAddressBag": []interface{}{
										map[string]interface{}{
											"cityName":             "MIAMI",
											"geographicRegionName": "FLORIDA",
											"geographicRegionCode": "FL",
											"countryCode":          "US",
											"postalCode":           "33134",
											"nameLineOneText":      "SANCHELIMA & ASSOCIATES, P.A.",
											"countryName":          "UNITED STATES",
											"addressLineOneText":   "235 S.W. LE JEUNE ROAD",
										},
									},
									"telecommunicationAddressBag": []interface{}{
										map[string]interface{}{
											"telecommunicationNumber": "305-447-1617",
											"telecomTypeCode":         "TEL",
										},
									},
									"registeredPractitionerCategory": "ATTNY",
								},
							},
							"attorneyBag": []interface{}{
								map[string]interface{}{
									"activeIndicator":    "ACTIVE",
									"firstName":          "JESUS",
									"lastName":           "SANCHELIMA",
									"registrationNumber": "28755",
									"attorneyAddressBag": []interface{}{
										map[string]interface{}{
											"cityName":             "MIAMI",
											"geographicRegionName": "FLORIDA",
											"geographicRegionCode": "FL",
											"countryCode":          "US",
											"postalCode":           "33134",
											"nameLineOneText":      "SANCHELIMA & ASSOCIATES, P.A.",
											"countryName":          "UNITED STATES",
											"addressLineOneText":   "235 S.W. LE JEUNE ROAD",
										},
									},
									"telecommunicationAddressBag": []interface{}{
										map[string]interface{}{
											"telecommunicationNumber": "305-447-1617",
											"telecomTypeCode":         "TEL",
										},
									},
									"registeredPractitionerCategory": "ATTNY",
								},
								map[string]interface{}{
									"activeIndicator":    "ACTIVE",
									"firstName":          "CHRISTIAN",
									"lastName":           "SANCHELIMA",
									"registrationNumber": "70150",
									"attorneyAddressBag": []interface{}{
										map[string]interface{}{
											"cityName":             "MIAMI",
											"geographicRegionName": "FLORIDA",
											"geographicRegionCode": "FL",
											"countryCode":          "US",
											"postalCode":           "33134",
											"nameLineOneText":      "SANCHELIMA & ASSOCIATES, P.A.",
											"countryName":          "UNITED STATES",
											"addressLineOneText":   "235 SW 42ND AVENUE",
										},
									},
									"telecommunicationAddressBag": []interface{}{
										map[string]interface{}{
											"telecommunicationNumber": "305-447-1617",
											"telecomTypeCode":         "TEL",
										},
									},
									"registeredPractitionerCategory": "AGENT",
								},
							},
						},
						"applicationNumberText": "17123456",
						"correspondenceAddressBag": []interface{}{
							map[string]interface{}{
								"cityName":             "MIAMI",
								"geographicRegionName": "FLORIDA",
								"geographicRegionCode": "FL",
								"countryCode":          "US",
								"postalCode":           "33134",
								"nameLineOneText":      "SANCHELIMA & ASSOCIATES, P.A.",
								"nameLineTwoText":      "CHRIS SANCHELIMA, ESQ.",
								"countryName":          "UNITED STATES",
								"addressLineOneText":   "235 S.W. LE JEUNE ROAD",
							},
						},
					},
				},
				"requestIdentifier": "937c3c3f-e76e-47c1-96a1-97551935d441",
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/patent/applications/17123456/adjustment":
			// Actual response from GetPatentAdjustment
			response := map[string]interface{}{
				"count": 1,
				"patentFileWrapperDataBag": []interface{}{
					map[string]interface{}{
						"patentTermAdjustmentData": map[string]interface{}{
							"applicantDayDelayQuantity":       0,
							"overlappingDayQuantity":          0,
							"ipOfficeAdjustmentDelayQuantity": 0,
							"cDelayQuantity":                  0,
							"adjustmentTotalQuantity":         238,
							"bDelayQuantity":                  0,
							"nonOverlappingDayDelayQuantity":  238,
							"aDelayQuantity":                  238,
							"patentTermAdjustmentHistoryDataBag": []interface{}{
								map[string]interface{}{
									"applicantDayDelayQuantity":      0,
									"eventDescriptionText":           "PTA 36 Months",
									"eventSequenceNumber":            64.5,
									"originatingEventSequenceNumber": 0.5,
									"ptaPTECode":                     "PTA",
									"ipOfficeDayDelayQuantity":       0,
									"eventDate":                      "2023-07-11",
								},
								map[string]interface{}{
									"applicantDayDelayQuantity":      0,
									"eventDescriptionText":           "Patent Issue Date Used in PTA Calculation",
									"eventSequenceNumber":            64.0,
									"originatingEventSequenceNumber": 0.0,
									"ptaPTECode":                     "PTA",
									"ipOfficeDayDelayQuantity":       0,
									"eventDate":                      "2023-07-11",
								},
							},
						},
						"applicationNumberText": "17123456",
					},
				},
				"requestIdentifier": "c30d8ea6-60b4-4e83-99f1-607d9c4d4e83",
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/patent/applications/17123456/continuity":
			// Actual response from GetPatentContinuity
			response := map[string]interface{}{
				"count": 1,
				"patentFileWrapperDataBag": []interface{}{
					map[string]interface{}{
						"childContinuityBag": []interface{}{
							map[string]interface{}{
								"childApplicationStatusDescriptionText": "Application Undergoing Preexam Processing",
								"claimParentageTypeCode":                "?",
								"childApplicationStatusCode":            19,
								"claimParentageTypeCodeDescriptionText": "is a no data of",
								"parentApplicationNumberText":           "17123456",
								"childApplicationNumberText":            "PCTUS2297200",
							},
						},
						"applicationNumberText": "17123456",
					},
				},
				"requestIdentifier": "a1477612-509d-4bb1-9184-729fd0286854",
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/patent/applications/17123456/documents":
			// Actual response from GetPatentDocuments
			response := map[string]interface{}{
				"count": 65,
				"documentBag": []interface{}{
					map[string]interface{}{
						"applicationNumberText":       "17123456",
						"officialDate":                "2023-07-11T00:00:00.000-0400",
						"documentIdentifier":          "96afb236-1465-43a2-9a65-f120f86b7bcd",
						"documentCode":                "EGRANT.PDF",
						"documentCodeDescriptionText": "Digitally signed official patent eGrant document",
						"directionCategory":           "OUTGOING",
						"downloadOptionBag": []interface{}{
							map[string]interface{}{
								"mimeTypeIdentifier": "PDF",
								"downloadUrl":        "https://api.uspto.gov/api/v1/download/applications/17123456/96afb236-1465-43a2-9a65-f120f86b7bcd/files/11696966_merged.pdf",
							},
						},
					},
					map[string]interface{}{
						"applicationNumberText":       "17123456",
						"officialDate":                "2023-07-11T00:00:00.000-0400",
						"documentIdentifier":          "LJXSUOXHGREENX5",
						"documentCode":                "EGRANT.NTF",
						"documentCodeDescriptionText": "eGrant day-of Notification",
						"directionCategory":           "OUTGOING",
						"downloadOptionBag": []interface{}{
							map[string]interface{}{
								"mimeTypeIdentifier": "PDF",
								"downloadUrl":        "https://api.uspto.gov/api/v1/download/applications/17123456/LJXSUOXHGREENX5.pdf",
								"pageTotalQuantity":  1,
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/patent/status-codes":
			// Actual response from GetStatusCodes
			response := map[string]interface{}{
				"count": 241,
				"statusCodeBag": []interface{}{
					map[string]interface{}{
						"applicationStatusCode":            1,
						"applicationStatusDescriptionText": "Missassigned Application Number",
					},
					map[string]interface{}{
						"applicationStatusCode":            3,
						"applicationStatusDescriptionText": "Proceedings Terminated",
					},
				},
				"requestIdentifier": "5c35a272-dcbc-4229-952c-24afdf9ebfca",
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/datasets/products/search":
			// Actual response from SearchBulkProducts
			response := map[string]interface{}{
				"count": 1,
				"bulkDataProductBag": []interface{}{
					map[string]interface{}{
						"productIdentifier":               "PTGRAPS",
						"productDescriptionText":          "Contains the concatenated full-text of each patent grant document issued weekly (Tuesdays) from January 1, 1976 to present (excludes images/drawings). This is the subset of the Patent Grant Full Text Data with Embedded TIFF Images.",
						"productTitleText":                "Patent Grant Full-Text Data (No Images) - APS",
						"productFrequencyText":            "YEARLY",
						"productLabelArrayText":           []string{"Patent"},
						"productDatasetArrayText":         []string{"Authoritative"},
						"productDatasetCategoryArrayText": []string{"Issued patents (patent grants)"},
						"productFromDate":                 "1976-01-06",
						"productToDate":                   "2001-12-25",
						"productTotalFileSize":            25938442242,
						"productFileTotalQuantity":        1382,
						"lastModifiedDateTime":            "2025-09-23T10:02:00Z",
						"mimeTypeIdentifierArrayText":     []string{"ASCII", "XML"},
					},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/datasets/products/PTGRXML":
			// Actual response from GetBulkProduct
			response := map[string]interface{}{
				"count": 1,
				"bulkDataProductBag": []interface{}{
					map[string]interface{}{
						"productIdentifier":               "PTGRXML",
						"productDescriptionText":          "Provides the bulk zip files that contains the concatenated full-text of each patent grant document issued weekly. This page provides an additional feature called “View Patent Records” which allows user to find or discover the patent grants that are bundled in the zip file. Even though the zip file may contain grants that belong to the applications which were filed before 2001, ODP will only show the grants that belong to the applications that were filed from 2001.",
						"productTitleText":                "Patent Grant Full-Text Data (No Images) - XML",
						"productFrequencyText":            "WEEKLY",
						"daysOfWeekText":                  "TUESDAY",
						"productLabelArrayText":           []string{"Patent"},
						"productDatasetArrayText":         []string{"Authoritative"},
						"productDatasetCategoryArrayText": []string{"Issued patents (patent grants)"},
						"productFromDate":                 "2002-01-01",
						"productToDate":                   "2025-09-23",
						"productTotalFileSize":            120319343938,
						"productFileTotalQuantity":        1267,
						"lastModifiedDateTime":            "2025-09-23T00:57:53Z",
						"mimeTypeIdentifierArrayText":     []string{"ASCII", "XML"},
						"productFileBag": map[string]interface{}{
							"count": 11,
							"fileDataBag": []interface{}{
								map[string]interface{}{
									"fileName":                 "ipg250923.zip",
									"fileSize":                 169944690,
									"fileDataFromDate":         "2025-09-23",
									"fileDataToDate":           "2025-09-23",
									"fileTypeText":             "Data",
									"fileDownloadURI":          "https://api.uspto.gov/api/v1/datasets/products/files/PTGRXML/2025/ipg250923.zip",
									"fileReleaseDate":          "2025-09-23T00:57:53Z",
									"fileLastModifiedDateTime": "2025-09-23T00:57:53Z",
								},
								map[string]interface{}{
									"fileName":                 "ipg250916.zip",
									"fileSize":                 112750995,
									"fileDataFromDate":         "2025-09-16",
									"fileDataToDate":           "2025-09-16",
									"fileTypeText":             "Data",
									"fileDownloadURI":          "https://api.uspto.gov/api/v1/datasets/products/files/PTGRXML/2025/ipg250916.zip",
									"fileReleaseDate":          "2025-09-16 00:57:52",
									"fileLastModifiedDateTime": "2025-09-16 00:57:52",
								},
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/petition/decisions/search":
			// Actual response from SearchPetitions
			response := map[string]interface{}{
				"count":             6,
				"requestIdentifier": "ebc0d202-5e5a-499e-80d9-450be6f821b4",
				"petitionDecisionDataBag": []interface{}{
					map[string]interface{}{
						"applicationNumberText":                "13986179",
						"businessEntityStatusCategory":         "Micro",
						"courtActionIndicator":                 false,
						"decisionDate":                         "2020-11-30",
						"decisionPetitionTypeCode":             502,
						"decisionTypeCode":                     "C",
						"decisionTypeCodeDescriptionText":      "DENIED",
						"finalDecidingOfficeName":              "OFFICE OF PETITIONS",
						"groupArtUnitNumber":                   "3656",
						"inventionTitle":                       "Centrifugal machine",
						"inventorBag":                          []string{"Richard Foore"},
						"lastIngestionDateTime":                "2025-07-16T19:49:16",
						"petitionDecisionRecordIdentifier":     "9dc6b94a-afa0-5e66-beef-f26fa80992b8",
						"petitionIssueConsideredTextBag":       []string{"Revival of an abandoned application"},
						"petitionMailDate":                     "2020-05-15",
						"prosecutionStatusCode":                161,
						"prosecutionStatusCodeDescriptionText": "Abandoned",
						"ruleBag":                              []string{"37 CFR 1.137(a)", "37 CFR 1.137(b)(1)"},
						"statuteBag":                           []string{"35 USC 27"},
						"technologyCenter":                     "3600",
					},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/datasets/products/files/PTGRXML/2025/ipg250916.zip":
			// GetBulkFileURL endpoint - returns 302 redirect
			w.Header().Set("Location", "https://data.uspto.gov/files/PTGRXML/2025/ipg250916.zip?Expires=1758628085&Signature=test&Key-Pair-Id=TEST123")
			w.WriteHeader(http.StatusFound)

		case "/api/v1/datasets/products/files/TEST/test.zip":
			// Mock download for FileDownloadURI validation tests
			zipData := []byte("PK\x03\x04test zip content for validation")
			w.Header().Set("Content-Type", "application/zip")
			w.Header().Set("Content-Length", "35")
			w.Write(zipData)

		case "/api/v1/patent/applications/15000001/assignment":
			// Actual response structure from demo/examples/get_patent_assignment/response.json
			response := map[string]interface{}{
				"count": 1,
				"patentFileWrapperDataBag": []interface{}{
					map[string]interface{}{
						"applicationNumberText": "15000001",
						"assignmentBag": []interface{}{
							map[string]interface{}{
								"assigneeBag": []interface{}{
									map[string]interface{}{
										"assigneeAddress": map[string]interface{}{
											"addressLineOneText":   "129, SAMSUNG-RO, YEONGTONG-GU",
											"cityName":             "SUWON-SI, GYEONGGI-DO",
											"geographicRegionName": "KRX",
											"postalCode":           "16677",
										},
										"assigneeNameText": "SAMSUNG ELECTRONICS CO., LTD",
									},
								},
								"assignmentMailedDate":     "2016-04-20",
								"assignmentReceivedDate":   "2016-04-19",
								"assignmentRecordedDate":   "2016-04-19",
								"conveyanceText":           "ASSIGNMENT OF ASSIGNORS INTEREST (SEE DOCUMENT FOR DETAILS).",
								"frameNumber":              190,
								"reelNumber":               38323,
								"reelAndFrameNumber":       "038323/0190",
								"pageTotalQuantity":        8,
								"imageAvailableStatusCode": false,
								"assignorBag": []interface{}{
									map[string]interface{}{
										"assignorName":  "HEO, JIN-PIL",
										"executionDate": "2016-01-05",
									},
									map[string]interface{}{
										"assignorName":  "JUNG, MIN-HWA",
										"executionDate": "2016-01-05",
									},
								},
								"correspondenceAddress": map[string]interface{}{
									"addressLineOneText":    "P.O. BOX 1213",
									"correspondentNameText": "MUIR PATENT LAW, PLLC",
								},
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/patent/applications/17248024/associated-documents":
			// Actual response structure from demo/examples/get_patent_associated_documents/response.json
			response := map[string]interface{}{
				"count": 1,
				"patentFileWrapperDataBag": []interface{}{
					map[string]interface{}{
						"applicationNumberText": "17248024",
						"grantDocumentMetaData": map[string]interface{}{
							"fileCreateDateTime": "2024-09-30T17:04:44",
							"fileLocationURI":    "https://api.uspto.gov/api/v1/datasets/products/files/PTGRXML-SPLT/2023/ipg230509/17248024_11646472.xml",
							"productIdentifier":  "PTGRXML",
							"xmlFileName":        "17248024_11646472.xml",
							"zipFileName":        "ipg230509.zip",
						},
						"pgpubDocumentMetaData": map[string]interface{}{
							"fileCreateDateTime": "2024-09-27T23:02:09",
							"fileLocationURI":    "https://api.uspto.gov/api/v1/datasets/products/files/APPXML-SPLT/2021/ipa210708/17248024_20210210819.xml",
							"productIdentifier":  "APPXML",
							"xmlFileName":        "17248024_20210210819.xml",
							"zipFileName":        "ipa210708.zip",
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/patent/applications/15000001/foreign-priority":
			// Actual response structure from demo/examples/get_patent_foreign_priority/response.json
			response := map[string]interface{}{
				"count": 1,
				"patentFileWrapperDataBag": []interface{}{
					map[string]interface{}{
						"applicationNumberText": "15000001",
						"foreignPriorityBag": []interface{}{
							map[string]interface{}{
								"applicationNumberText": "10-2014-0009131",
								"filingDate":            "2015-01-20",
								"ipOfficeName":          "REPUBLIC OF KOREA",
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/patent/applications/17248024/meta-data":
			// Actual response structure from demo/examples/get_patent_meta_data/response.json
			response := map[string]interface{}{
				"count": 1,
				"patentFileWrapperDataBag": []interface{}{
					map[string]interface{}{
						"applicationMetaData": map[string]interface{}{
							"applicationConfirmationNumber":    4114,
							"applicationStatusCode":            150,
							"applicationStatusDate":            "2023-04-19",
							"applicationStatusDescriptionText": "Patented Case",
							"applicationTypeCategory":          "REGULAR",
							"applicationTypeCode":              "UTL",
							"applicationTypeLabelName":         "Utility",
							"class":                            "429",
							"customerNumber":                   22434,
							"docketNumber":                     "PLUSP040X1C4US",
							"earliestPublicationDate":          "2021-07-08",
							"earliestPublicationNumber":        "US20210210819A1",
							"effectiveFilingDate":              "2021-01-05",
							"examinerNameText":                 "DOVE, TRACY MAE",
							"filingDate":                       "2021-01-05",
							"firstApplicantName":               "PolyPlus Battery Company",
							"firstInventorName":                "Steven J. Visco",
							"grantDate":                        "2023-05-09",
							"groupArtUnitNumber":               "1723",
							"inventionTitle":                   "Electrode protection using a composite comprising an ionic conducting protective layer",
							"patentNumber":                     "11646472",
							"subclass":                         "303",
							"applicantBag": []interface{}{
								map[string]interface{}{
									"applicantNameText": "PolyPlus Battery Company",
								},
							},
							"cpcClassificationBag": []string{"H01M 50/46", "H01G 11/06"},
						},
						"applicationNumberText": "17248024",
					},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/patent/applications/17248024/transactions":
			// Actual response structure from demo/examples/get_patent_transactions/response.json
			response := map[string]interface{}{
				"count": 1,
				"patentFileWrapperDataBag": []interface{}{
					map[string]interface{}{
						"applicationNumberText": "17248024",
						"eventDataBag": []interface{}{
							map[string]interface{}{
								"eventCode":            "EML_NTR",
								"eventDate":            "2023-10-03",
								"eventDescriptionText": "Email Notification",
							},
							map[string]interface{}{
								"eventCode":            "MOPPT",
								"eventDate":            "2023-10-03",
								"eventDescriptionText": "Mail O.P. Petition Decision",
							},
							map[string]interface{}{
								"eventCode":            "ELC_RVW",
								"eventDate":            "2023-05-09",
								"eventDescriptionText": "Electronic Review",
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/patent/applications/17248024/attorney":
			// Actual response structure from demo/examples/get_patent_attorney/response.json
			response := map[string]interface{}{
				"count": 1,
				"patentFileWrapperDataBag": []interface{}{
					map[string]interface{}{
						"applicationNumberText": "17248024",
						"recordAttorney": map[string]interface{}{
							"attorneyBag": []interface{}{
								map[string]interface{}{
									"activeIndicator": "ACTIVE",
									"attorneyAddressBag": []interface{}{
										map[string]interface{}{
											"addressLineOneText":   "PO BOX 70250",
											"cityName":             "OAKLAND",
											"countryCode":          "US",
											"countryName":          "UNITED STATES",
											"geographicRegionCode": "CA",
											"geographicRegionName": "CALIFORNIA",
											"nameLineOneText":      "WEAVER AUSTIN VILLENEUVE & SAMPSON LLP",
											"postalCode":           "94612-025",
										},
									},
									"firstName":                      "JEFFREY",
									"lastName":                       "WEAVER",
									"registeredPractitionerCategory": "ATTNY",
									"registrationNumber":             "31314",
									"telecommunicationAddressBag": []interface{}{
										map[string]interface{}{
											"telecomTypeCode":         "TEL",
											"telecommunicationNumber": "510-663-1100",
										},
									},
								},
							},
							"customerNumberCorrespondenceData": map[string]interface{}{
								"patronIdentifier": 22434,
								"powerOfAttorneyAddressBag": []interface{}{
									map[string]interface{}{
										"addressLineOneText":   "PO BOX 70250",
										"cityName":             "OAKLAND",
										"countryCode":          "US",
										"countryName":          "UNITED STATES",
										"geographicRegionCode": "CA",
										"geographicRegionName": "CALIFORNIA",
										"nameLineOneText":      "WEAVER AUSTIN VILLENEUVE & SAMPSON LLP",
										"postalCode":           "94612-0250",
									},
								},
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/patent/applications/search/download":
			// Mock response for SearchPatentsDownload - returns CSV data
			w.Header().Set("Content-Type", "text/csv")
			csvData := []byte("Application Number,Filing Date,Title\n16123456,2020-01-15,Machine Learning System\n17234567,2021-02-20,AI Processing Method")
			w.Write(csvData)

		case "/api/v1/petition/decisions/test-petition-id":
			// Mock response for GetPetitionDecision
			response := map[string]interface{}{
				"petitionDecisionData": map[string]interface{}{
					"petitionDecisionRecordIdentifier": "test-petition-id",
					"applicationNumberText":            "13986179",
					"decisionDate":                     "2020-11-30",
					"decisionTypeCode":                 "C",
					"decisionTypeCodeDescriptionText":  "DENIED",
					"finalDecidingOfficeName":          "OFFICE OF PETITIONS",
					"inventionTitle":                   "Centrifugal machine",
					"petitionIssueConsideredTextBag":   []string{"Revival of an abandoned application"},
					"ruleBag":                          []string{"37 CFR 1.137(a)", "37 CFR 1.137(b)(1)"},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/petition/decisions/search/download":
			// Mock response for SearchPetitionsDownload - returns CSV data
			w.Header().Set("Content-Type", "text/csv")
			csvData := []byte("Record ID,Application Number,Decision Date,Decision Type\ntest-id-1,13986179,2020-11-30,DENIED\ntest-id-2,14123456,2021-01-15,GRANTED")
			w.Write(csvData)

		// PTAB API endpoints - Trial Proceedings
		case "/api/v1/patent/trials/proceedings/search":
			response := map[string]interface{}{
				"count": 500,
				"patentTrialProceedingDataBag": []interface{}{
					map[string]interface{}{
						"trialNumber": "IPR2020-00001",
						"trialMetaData": map[string]interface{}{
							"trialTypeCode":        "IPR",
							"trialStatusCategory":  "Terminated",
							"petitionFilingDate":   "2020-01-15",
							"trialInstitutionDate": "2020-07-15",
							"fileDownloadURI":      "https://api.uspto.gov/ptab/files/IPR2020-00001",
						},
						"patentOwnerData": map[string]interface{}{
							"patentNumber":         "10123456",
							"patentOwnerPartyName": "ACME Corp",
						},
						"petitionerPartyName": "Tech Company LLC",
					},
				},
				"requestIdentifier": "ptab-proc-search-123",
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/patent/trials/proceedings/IPR2020-00001":
			response := map[string]interface{}{
				"count": 1,
				"patentTrialProceedingDataBag": []interface{}{
					map[string]interface{}{
						"trialNumber": "IPR2020-00001",
						"trialMetaData": map[string]interface{}{
							"trialTypeCode":       "IPR",
							"trialStatusCategory": "Terminated",
							"petitionFilingDate":  "2020-01-15",
						},
					},
				},
				"requestIdentifier": "ptab-proc-get-123",
			}
			json.NewEncoder(w).Encode(response)

		// PTAB API endpoints - Trial Decisions
		case "/api/v1/patent/trials/decisions/search":
			response := map[string]interface{}{
				"count": 100,
				"patentTrialDocumentDataBag": []interface{}{
					map[string]interface{}{
						"trialNumber": "IPR2020-00001",
						"documentData": map[string]interface{}{
							"documentTitleText":  "Final Written Decision",
							"documentFilingDate": "2021-01-15",
							"documentIdentifier": "doc-123",
						},
					},
				},
				"requestIdentifier": "ptab-decisions-search-123",
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/patent/trials/decisions/doc-123":
			response := map[string]interface{}{
				"count": 1,
				"patentTrialDocumentDataBag": []interface{}{
					map[string]interface{}{
						"trialNumber": "IPR2020-00001",
						"documentData": map[string]interface{}{
							"documentTitleText":  "Final Written Decision",
							"documentFilingDate": "2021-01-15",
							"documentIdentifier": "doc-123",
						},
					},
				},
				"requestIdentifier": "ptab-decision-get-123",
			}
			json.NewEncoder(w).Encode(response)

		// PTAB API endpoints - Trial Documents
		case "/api/v1/patent/trials/documents/search":
			response := map[string]interface{}{
				"count": 200,
				"patentTrialDocumentDataBag": []interface{}{
					map[string]interface{}{
						"trialNumber": "IPR2020-00001",
						"documentData": map[string]interface{}{
							"documentName":       "Petition.pdf",
							"documentCategory":   "PETITION",
							"documentIdentifier": "tdoc-456",
						},
					},
				},
				"requestIdentifier": "ptab-docs-search-123",
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/patent/trials/documents/tdoc-456":
			response := map[string]interface{}{
				"count": 1,
				"patentTrialDocumentDataBag": []interface{}{
					map[string]interface{}{
						"trialNumber": "IPR2020-00001",
						"documentData": map[string]interface{}{
							"documentName":       "Petition.pdf",
							"documentCategory":   "PETITION",
							"documentIdentifier": "tdoc-456",
						},
					},
				},
				"requestIdentifier": "ptab-doc-get-123",
			}
			json.NewEncoder(w).Encode(response)

		// PTAB API endpoints - Appeal Decisions
		case "/api/v1/patent/appeals/decisions/search":
			response := map[string]interface{}{
				"count": 50,
				"patentAppealDataBag": []interface{}{
					map[string]interface{}{
						"appealNumber": "2020-000001",
						"documentData": map[string]interface{}{
							"documentName":       "Decision on Appeal",
							"documentFilingDate": "2020-06-15",
							"documentIdentifier": "appeal-doc-789",
						},
						"appelantData": map[string]interface{}{
							"applicationNumberText": "15123456",
							"inventorName":          "John Smith",
						},
					},
				},
				"requestIdentifier": "ptab-appeal-search-123",
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/patent/appeals/decisions/appeal-doc-789":
			response := map[string]interface{}{
				"count": 1,
				"patentAppealDataBag": []interface{}{
					map[string]interface{}{
						"appealNumber": "2020-000001",
						"documentData": map[string]interface{}{
							"documentName":       "Decision on Appeal",
							"documentFilingDate": "2020-06-15",
							"documentIdentifier": "appeal-doc-789",
						},
					},
				},
				"requestIdentifier": "ptab-appeal-get-123",
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/patent/appeals/2020-000001/decisions":
			response := map[string]interface{}{
				"count": 2,
				"patentAppealDataBag": []interface{}{
					map[string]interface{}{
						"appealNumber": "2020-000001",
						"documentData": map[string]interface{}{
							"documentName":       "Decision on Appeal",
							"documentFilingDate": "2020-06-15",
						},
					},
					map[string]interface{}{
						"appealNumber": "2020-000001",
						"documentData": map[string]interface{}{
							"documentName":       "Rehearing Decision",
							"documentFilingDate": "2020-09-15",
						},
					},
				},
				"requestIdentifier": "ptab-appeal-by-number-123",
			}
			json.NewEncoder(w).Encode(response)

		// PTAB API endpoints - Interference Decisions
		case "/api/v1/patent/interferences/decisions/search":
			response := map[string]interface{}{
				"count":             1811,
				"requestIdentifier": "ptab-interference-search-123",
				"patentInterferenceDataBag": []interface{}{
					map[string]interface{}{
						"interferenceNumber": "106130",
						"documentData": map[string]interface{}{
							"documentTitleText":    "Judgment 37 C.F.R. § 41.127(a)",
							"decisionIssueDate":    "2025-01-28",
							"documentIdentifier":   "229ba0b8d5f70d2e45cc36b79476f56f3faf51bd26c7ccc977208e7b",
							"decisionTypeCategory": "Decision",
						},
						"interferenceMetaData": map[string]interface{}{
							"interferenceStyleName":        "LEE M. KAPLAN v. PATRICE CANI",
							"interferenceLastModifiedDate": "2025-11-13",
						},
						"lastModifiedDateTime": "2025-11-20T03:12:32",
					},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/patent/interferences/decisions/229ba0b8d5f70d2e45cc36b79476f56f3faf51bd26c7ccc977208e7b":
			response := map[string]interface{}{
				"count":             1,
				"requestIdentifier": "ptab-interference-get-123",
				"patentInterferenceDataBag": []interface{}{
					map[string]interface{}{
						"interferenceNumber": "106130",
						"documentData": map[string]interface{}{
							"documentTitleText":    "Judgment 37 C.F.R. § 41.127(a)",
							"decisionIssueDate":    "2025-01-28",
							"documentIdentifier":   "229ba0b8d5f70d2e45cc36b79476f56f3faf51bd26c7ccc977208e7b",
							"decisionTypeCategory": "Decision",
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/patent/interferences/106130/decisions":
			response := map[string]interface{}{
				"count":             2,
				"requestIdentifier": "ptab-interference-by-number-123",
				"patentInterferenceDataBag": []interface{}{
					map[string]interface{}{
						"interferenceNumber": "106130",
						"documentData": map[string]interface{}{
							"documentTitleText": "Judgment 37 C.F.R. § 41.127(a)",
							"decisionIssueDate": "2025-01-28",
						},
					},
					map[string]interface{}{
						"interferenceNumber": "106130",
						"documentData": map[string]interface{}{
							"documentTitleText": "Decision on Priority",
							"decisionIssueDate": "2025-01-28",
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)

		// PTAB API - Trial decisions/documents by trial number
		case "/api/v1/patent/trials/IPR2020-00001/decisions":
			response := map[string]interface{}{
				"count": 2,
				"patentTrialDocumentDataBag": []interface{}{
					map[string]interface{}{
						"trialNumber": "IPR2020-00001",
						"documentData": map[string]interface{}{
							"documentTitleText":  "Institution Decision",
							"documentFilingDate": "2020-07-15",
						},
					},
					map[string]interface{}{
						"trialNumber": "IPR2020-00001",
						"documentData": map[string]interface{}{
							"documentTitleText":  "Final Written Decision",
							"documentFilingDate": "2021-07-15",
						},
					},
				},
				"requestIdentifier": "ptab-trial-decisions-by-number-123",
			}
			json.NewEncoder(w).Encode(response)

		case "/api/v1/patent/trials/IPR2020-00001/documents":
			response := map[string]interface{}{
				"count": 3,
				"patentTrialDocumentDataBag": []interface{}{
					map[string]interface{}{
						"trialNumber": "IPR2020-00001",
						"documentData": map[string]interface{}{
							"documentName":     "Petition.pdf",
							"documentCategory": "PETITION",
						},
					},
					map[string]interface{}{
						"trialNumber": "IPR2020-00001",
						"documentData": map[string]interface{}{
							"documentName":     "Patent Owner Response.pdf",
							"documentCategory": "RESPONSE",
						},
					},
				},
				"requestIdentifier": "ptab-trial-documents-by-number-123",
			}
			json.NewEncoder(w).Encode(response)

		// PTAB API - Download endpoints
		case "/api/v1/patent/trials/proceedings/search/download":
			w.Header().Set("Content-Type", "text/csv")
			csvData := []byte("Trial Number,Status,Filing Date\nIPR2020-00001,Terminated,2020-01-15\nIPR2021-00002,Instituted,2021-03-20")
			w.Write(csvData)

		case "/api/v1/patent/trials/decisions/search/download":
			w.Header().Set("Content-Type", "text/csv")
			csvData := []byte("Trial Number,Decision Type,Decision Date\nIPR2020-00001,Final Written Decision,2021-01-15")
			w.Write(csvData)

		case "/api/v1/patent/trials/documents/search/download":
			w.Header().Set("Content-Type", "text/csv")
			csvData := []byte("Trial Number,Document Name,Category\nIPR2020-00001,Petition.pdf,PETITION")
			w.Write(csvData)

		case "/api/v1/patent/appeals/decisions/search/download":
			w.Header().Set("Content-Type", "text/csv")
			csvData := []byte("Appeal Number,Decision Date,Outcome\n2020-000001,2020-06-15,Affirmed")
			w.Write(csvData)

		case "/api/v1/patent/interferences/decisions/search/download":
			w.Header().Set("Content-Type", "text/csv")
			csvData := []byte("Interference Number,Decision Date,Outcome\n106130,2025-01-28,Judgment")
			w.Write(csvData)

		default:
			// Special handling for bulk file download URLs that include redirect
			if strings.HasPrefix(r.URL.Path, "https://data.uspto.gov/files/") {
				// This is the actual download URL after redirect - return mock ZIP data
				w.Header().Set("Content-Type", "application/zip")
				// Mock ZIP file header (PK signature)
				zipData := []byte("PK\x03\x04MockZipFileContent...")
				w.Write(zipData)
				return
			}
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}
	}))
	defer server.Close()

	// Create client with test configuration
	config := &Config{
		BaseURL:    server.URL,
		APIKey:     "test-key",
		MaxRetries: 1,
		Timeout:    10,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Run("SearchPatents", func(t *testing.T) {
		result, err := client.SearchPatents(ctx, "test", 0, 10)
		if err != nil {
			t.Fatalf("SearchPatents failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		// Basic validation - check that we got data
		jsonData, _ := json.Marshal(result)
		if len(jsonData) < 10 {
			t.Error("Response seems empty")
		}
	})

	t.Run("GetPatent", func(t *testing.T) {
		result, err := client.GetPatent(ctx, "17123456")
		if err != nil {
			t.Fatalf("GetPatent failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		// Basic validation - check that we got data
		jsonData, _ := json.Marshal(result)
		if len(jsonData) < 10 {
			t.Error("Response seems empty")
		}
	})

	t.Run("GetPatentAdjustment", func(t *testing.T) {
		result, err := client.GetPatentAdjustment(ctx, "17123456")
		if err != nil {
			t.Fatalf("GetPatentAdjustment failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.ApplicationNumber != "17123456" {
			t.Errorf("ApplicationNumber = %q, want %q", result.ApplicationNumber, "17123456")
		}
		if result.TotalAdjustmentDays != 238 {
			t.Errorf("TotalAdjustmentDays = %d, want 238", result.TotalAdjustmentDays)
		}
		if result.ADelays != 238 {
			t.Errorf("ADelays = %d, want 238", result.ADelays)
		}
		if result.BDelays != 0 {
			t.Errorf("BDelays = %d, want 0", result.BDelays)
		}
	})

	t.Run("GetPatentContinuity", func(t *testing.T) {
		result, err := client.GetPatentContinuity(ctx, "17123456")
		if err != nil {
			t.Fatalf("GetPatentContinuity failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.ApplicationNumber != "17123456" {
			t.Errorf("ApplicationNumber = %q, want %q", result.ApplicationNumber, "17123456")
		}
		if len(result.Children) != 1 {
			t.Fatalf("Expected 1 child, got %d", len(result.Children))
		}
		ch := result.Children[0]
		if ch.ApplicationNumber != "PCTUS2297200" {
			t.Errorf("Child.ApplicationNumber = %q, want %q", ch.ApplicationNumber, "PCTUS2297200")
		}
		if ch.Status != "Application Undergoing Preexam Processing" {
			t.Errorf("Child.Status = %q", ch.Status)
		}
		if result.Parents == nil {
			t.Error("Parents should be non-nil")
		}
	})

	t.Run("GetPatentDocuments", func(t *testing.T) {
		result, err := client.GetPatentDocuments(ctx, "17123456")
		if err != nil {
			t.Fatalf("GetPatentDocuments failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		// Basic validation - check that we got data
		jsonData, _ := json.Marshal(result)
		if len(jsonData) < 10 {
			t.Error("Response seems empty")
		}
	})

	t.Run("GetStatusCodes", func(t *testing.T) {
		result, err := client.GetStatusCodes(ctx)
		if err != nil {
			t.Fatalf("GetStatusCodes failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		// Basic validation - check that we got data
		jsonData, _ := json.Marshal(result)
		if len(jsonData) < 10 {
			t.Error("Response seems empty")
		}
	})

	t.Run("SearchBulkProducts", func(t *testing.T) {
		result, err := client.SearchBulkProducts(ctx, "patent grant", 0, 10)
		if err != nil {
			t.Fatalf("SearchBulkProducts failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		// Basic validation - check that we got data
		jsonData, _ := json.Marshal(result)
		if len(jsonData) < 10 {
			t.Error("Response seems empty")
		}
	})

	t.Run("GetBulkProduct", func(t *testing.T) {
		result, err := client.GetBulkProduct(ctx, "PTGRXML")
		if err != nil {
			t.Fatalf("GetBulkProduct failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		// Basic validation - check that we got data
		jsonData, _ := json.Marshal(result)
		if len(jsonData) < 10 {
			t.Error("Response seems empty")
		}
	})

	t.Run("SearchPetitions", func(t *testing.T) {
		result, err := client.SearchPetitions(ctx, "revival", 0, 10)
		if err != nil {
			t.Fatalf("SearchPetitions failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		// Basic validation - check that we got data
		jsonData, _ := json.Marshal(result)
		if len(jsonData) < 10 {
			t.Error("Response seems empty")
		}
	})

	t.Run("GetPatentAssignment", func(t *testing.T) {
		// Uses 15000001 (Samsung patent with assignment data)
		result, err := client.GetPatentAssignment(ctx, "15000001")
		if err != nil {
			t.Fatalf("GetPatentAssignment failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
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
		if a.ReelFrame != "038323/0190" {
			t.Errorf("ReelFrame = %q, want %q", a.ReelFrame, "038323/0190")
		}
		if a.Assignor == "" {
			t.Error("Assignor should not be empty")
		}
	})

	t.Run("GetPatentAssociatedDocuments", func(t *testing.T) {
		// Uses 17248024 (PolyPlus patent)
		result, err := client.GetPatentAssociatedDocuments(ctx, "17248024")
		if err != nil {
			t.Fatalf("GetPatentAssociatedDocuments failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("GetPatentAttorney", func(t *testing.T) {
		// Uses 17248024 (PolyPlus patent)
		result, err := client.GetPatentAttorney(ctx, "17248024")
		if err != nil {
			t.Fatalf("GetPatentAttorney failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("GetPatentForeignPriority", func(t *testing.T) {
		// Uses 15000001 (Samsung patent with Korean priority)
		result, err := client.GetPatentForeignPriority(ctx, "15000001")
		if err != nil {
			t.Fatalf("GetPatentForeignPriority failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("GetPatentMetaData", func(t *testing.T) {
		// Uses 17248024 (PolyPlus patent)
		result, err := client.GetPatentMetaData(ctx, "17248024")
		if err != nil {
			t.Fatalf("GetPatentMetaData failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("GetPatentTransactions", func(t *testing.T) {
		// Uses 17248024 (PolyPlus patent)
		result, err := client.GetPatentTransactions(ctx, "17248024")
		if err != nil {
			t.Fatalf("GetPatentTransactions failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.ApplicationNumber != "17248024" {
			t.Errorf("ApplicationNumber = %q, want %q", result.ApplicationNumber, "17248024")
		}
		if len(result.Events) != 3 {
			t.Fatalf("Expected 3 events, got %d", len(result.Events))
		}
		e0 := result.Events[0]
		if e0.Code != "EML_NTR" {
			t.Errorf("Events[0].Code = %q, want %q", e0.Code, "EML_NTR")
		}
		if e0.Date != "2023-10-03" {
			t.Errorf("Events[0].Date = %q, want %q", e0.Date, "2023-10-03")
		}
	})

	t.Run("SearchPatentsDownload", func(t *testing.T) {
		req := generated.PatentDownloadRequest{
			Q: StringPtr("machine learning"),
			Pagination: &generated.Pagination{
				Offset: Int32Ptr(0),
				Limit:  Int32Ptr(10),
			},
		}
		result, err := client.SearchPatentsDownload(ctx, req)
		if err != nil {
			t.Fatalf("SearchPatentsDownload failed: %v", err)
		}
		if len(result) == 0 {
			t.Fatal("Expected data, got empty result")
		}
	})

	t.Run("GetPetitionDecision", func(t *testing.T) {
		result, err := client.GetPetitionDecision(ctx, "test-petition-id", false)
		if err != nil {
			t.Fatalf("GetPetitionDecision failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("SearchPetitionsDownload", func(t *testing.T) {
		req := generated.PetitionDecisionDownloadRequest{
			Q: StringPtr("revival"),
			Pagination: &generated.Pagination{
				Offset: Int32Ptr(0),
				Limit:  Int32Ptr(10),
			},
		}
		result, err := client.SearchPetitionsDownload(ctx, req)
		if err != nil {
			t.Fatalf("SearchPetitionsDownload failed: %v", err)
		}
		if len(result) == 0 {
			t.Fatal("Expected data, got empty result")
		}
	})

	t.Run("DownloadBulkFile_Validation", func(t *testing.T) {
		// Test validation - should reject invalid URLs
		var buf bytes.Buffer
		err := client.DownloadBulkFile(ctx, "https://data.uspto.gov/files/redirect/test.zip", &buf)
		if err == nil {
			t.Fatal("Expected validation error for invalid FileDownloadURI")
		}
		if !strings.Contains(err.Error(), "invalid FileDownloadURI") {
			t.Errorf("Expected validation error, got: %v", err)
		}
		t.Log("FileDownloadURI validation working correctly")
	})

	t.Run("DownloadBulkFileWithProgress_Validation", func(t *testing.T) {
		// Test validation - should reject invalid URLs
		var buf bytes.Buffer
		err := client.DownloadBulkFileWithProgress(ctx, "https://data.uspto.gov/files/redirect/test.zip", &buf, nil)
		if err == nil {
			t.Fatal("Expected validation error for invalid FileDownloadURI")
		}
		if !strings.Contains(err.Error(), "invalid FileDownloadURI") {
			t.Errorf("Expected validation error, got: %v", err)
		}
		t.Log("FileDownloadURI validation working correctly for DownloadBulkFileWithProgress")
	})

	// PTAB API Tests
	t.Run("SearchTrialProceedings", func(t *testing.T) {
		result, err := client.SearchTrialProceedings(ctx, "trialMetaData.trialTypeCode:IPR", 0, 10)
		if err != nil {
			t.Fatalf("SearchTrialProceedings failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		jsonData, _ := json.Marshal(result)
		if len(jsonData) < 10 {
			t.Error("Response seems empty")
		}
	})

	t.Run("GetTrialProceeding", func(t *testing.T) {
		result, err := client.GetTrialProceeding(ctx, "IPR2020-00001")
		if err != nil {
			t.Fatalf("GetTrialProceeding failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("SearchTrialDecisions", func(t *testing.T) {
		result, err := client.SearchTrialDecisions(ctx, "", 0, 10)
		if err != nil {
			t.Fatalf("SearchTrialDecisions failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("GetTrialDecision", func(t *testing.T) {
		result, err := client.GetTrialDecision(ctx, "doc-123")
		if err != nil {
			t.Fatalf("GetTrialDecision failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("SearchTrialDocuments", func(t *testing.T) {
		result, err := client.SearchTrialDocuments(ctx, "", 0, 10)
		if err != nil {
			t.Fatalf("SearchTrialDocuments failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("GetTrialDocument", func(t *testing.T) {
		result, err := client.GetTrialDocument(ctx, "tdoc-456")
		if err != nil {
			t.Fatalf("GetTrialDocument failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("SearchAppealDecisions", func(t *testing.T) {
		result, err := client.SearchAppealDecisions(ctx, "", 0, 10)
		if err != nil {
			t.Fatalf("SearchAppealDecisions failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("GetAppealDecision", func(t *testing.T) {
		result, err := client.GetAppealDecision(ctx, "appeal-doc-789")
		if err != nil {
			t.Fatalf("GetAppealDecision failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("GetAppealDecisionsByAppealNumber", func(t *testing.T) {
		result, err := client.GetAppealDecisionsByAppealNumber(ctx, "2020-000001")
		if err != nil {
			t.Fatalf("GetAppealDecisionsByAppealNumber failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("SearchInterferenceDecisions", func(t *testing.T) {
		result, err := client.SearchInterferenceDecisions(ctx, "", 0, 10)
		if err != nil {
			t.Fatalf("SearchInterferenceDecisions failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("GetInterferenceDecision", func(t *testing.T) {
		result, err := client.GetInterferenceDecision(ctx, "229ba0b8d5f70d2e45cc36b79476f56f3faf51bd26c7ccc977208e7b")
		if err != nil {
			t.Fatalf("GetInterferenceDecision failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("GetInterferenceDecisionsByNumber", func(t *testing.T) {
		result, err := client.GetInterferenceDecisionsByNumber(ctx, "106130")
		if err != nil {
			t.Fatalf("GetInterferenceDecisionsByNumber failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("GetTrialDecisionsByTrialNumber", func(t *testing.T) {
		result, err := client.GetTrialDecisionsByTrialNumber(ctx, "IPR2020-00001")
		if err != nil {
			t.Fatalf("GetTrialDecisionsByTrialNumber failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("GetTrialDocumentsByTrialNumber", func(t *testing.T) {
		result, err := client.GetTrialDocumentsByTrialNumber(ctx, "IPR2020-00001")
		if err != nil {
			t.Fatalf("GetTrialDocumentsByTrialNumber failed: %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("SearchTrialProceedingsDownload", func(t *testing.T) {
		req := generated.DownloadRequest{
			Q: StringPtr(""),
		}
		result, err := client.SearchTrialProceedingsDownload(ctx, req)
		if err != nil {
			t.Fatalf("SearchTrialProceedingsDownload failed: %v", err)
		}
		if len(result) == 0 {
			t.Fatal("Expected data, got empty")
		}
	})

	t.Run("SearchTrialDecisionsDownload", func(t *testing.T) {
		req := generated.DownloadRequest{
			Q: StringPtr(""),
		}
		result, err := client.SearchTrialDecisionsDownload(ctx, req)
		if err != nil {
			t.Fatalf("SearchTrialDecisionsDownload failed: %v", err)
		}
		if len(result) == 0 {
			t.Fatal("Expected data, got empty")
		}
	})

	t.Run("SearchTrialDocumentsDownload", func(t *testing.T) {
		req := generated.DownloadRequest{
			Q: StringPtr(""),
		}
		result, err := client.SearchTrialDocumentsDownload(ctx, req)
		if err != nil {
			t.Fatalf("SearchTrialDocumentsDownload failed: %v", err)
		}
		if len(result) == 0 {
			t.Fatal("Expected data, got empty")
		}
	})

	t.Run("SearchAppealDecisionsDownload", func(t *testing.T) {
		req := generated.DownloadRequest{
			Q: StringPtr(""),
		}
		result, err := client.SearchAppealDecisionsDownload(ctx, req)
		if err != nil {
			t.Fatalf("SearchAppealDecisionsDownload failed: %v", err)
		}
		if len(result) == 0 {
			t.Fatal("Expected data, got empty")
		}
	})

	t.Run("SearchInterferenceDecisionsDownload", func(t *testing.T) {
		req := generated.PatentDownloadRequest{
			Q: StringPtr(""),
		}
		result, err := client.SearchInterferenceDecisionsDownload(ctx, req)
		if err != nil {
			t.Fatalf("SearchInterferenceDecisionsDownload failed: %v", err)
		}
		if len(result) == 0 {
			t.Fatal("Expected data, got empty")
		}
	})

}
