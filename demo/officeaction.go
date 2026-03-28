package main

import (
	"fmt"
)

func demoOfficeAction(dctx *DemoContext) {
	printHeader("Office Action API Demonstrations")

	demoSearchOfficeActions(dctx)
	demoGetOfficeActionFields(dctx)
	demoSearchOfficeActionCitations(dctx)
	demoGetOfficeActionCitationFields(dctx)
	demoSearchOfficeActionRejections(dctx)
	demoGetOfficeActionRejectionFields(dctx)
	demoSearchEnrichedCitations(dctx)
	demoGetEnrichedCitationFields(dctx)
}

func demoSearchOfficeActions(dctx *DemoContext) {
	printSubHeader("SearchOfficeActions (Text Retrieval)")

	result, err := dctx.Client.SearchOfficeActions(dctx.Ctx, "patentApplicationNumber:17248024", 0, 3)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Total records: %d\n", result.Response.NumFound)
	fmt.Printf("Returned: %d\n", len(result.Response.Docs))

	for i, doc := range result.Response.Docs {
		if i >= 3 {
			break
		}
		fmt.Printf("\n%d. App: %v", i+1, doc["patentApplicationNumber"])
		if v, ok := doc["actionType"]; ok {
			fmt.Printf("\n   Action Type: %v", v)
		}
		if v, ok := doc["mailedDate"]; ok {
			fmt.Printf("\n   Mailed: %v", v)
		}
		fmt.Println()
	}

	dctx.saveExample("search_office_actions", map[string]string{"criteria": "patentApplicationNumber:17248024"}, result)
}

func demoGetOfficeActionFields(dctx *DemoContext) {
	printSubHeader("GetOfficeActionFields")

	result, err := dctx.Client.GetOfficeActionFields(dctx.Ctx)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("API: %s (v%s)\n", result.APIKey, result.APIVersionNumber)
	fmt.Printf("Status: %s\n", result.APIStatus)
	fmt.Printf("Fields (%d):\n", result.FieldCount)
	for i, field := range result.Fields {
		if i >= 10 {
			fmt.Printf("  ... and %d more\n", len(result.Fields)-10)
			break
		}
		fmt.Printf("  - %s\n", field)
	}

	dctx.saveExample("get_office_action_fields", map[string]string{}, result)
}

func demoSearchOfficeActionCitations(dctx *DemoContext) {
	printSubHeader("SearchOfficeActionCitations")

	result, err := dctx.Client.SearchOfficeActionCitations(dctx.Ctx, "patentApplicationNumber:16123456", 0, 5)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Total citation records: %d\n", result.Response.NumFound)
	fmt.Printf("Returned: %d\n", len(result.Response.Docs))

	for i, doc := range result.Response.Docs {
		if i >= 3 {
			break
		}
		fmt.Printf("\n%d. App: %v", i+1, doc["patentApplicationNumber"])
		if v, ok := doc["legalSectionCode"]; ok {
			fmt.Printf("  Section: %v", v)
		}
		if v, ok := doc["referenceIdentifier"]; ok {
			fmt.Printf("  Ref: %v", v)
		}
		fmt.Println()
	}

	dctx.saveExample("search_office_action_citations", map[string]string{"criteria": "patentApplicationNumber:16123456"}, result)
}

func demoSearchOfficeActionRejections(dctx *DemoContext) {
	printSubHeader("SearchOfficeActionRejections")

	result, err := dctx.Client.SearchOfficeActionRejections(dctx.Ctx, "patentApplicationNumber:12190351", 0, 3)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Total rejection records: %d\n", result.Response.NumFound)
	fmt.Printf("Returned: %d\n", len(result.Response.Docs))

	for i, doc := range result.Response.Docs {
		if i >= 3 {
			break
		}
		fmt.Printf("\n%d. App: %v  GAU: %v", i+1, doc["patentApplicationNumber"], doc["groupArtUnitNumber"])
		fmt.Printf("\n   101: %v  102: %v  103: %v  112: %v  DP: %v",
			doc["hasRej101"], doc["hasRej102"], doc["hasRej103"], doc["hasRej112"], doc["hasRejDP"])
		if alice, ok := doc["aliceIndicator"]; ok {
			fmt.Printf("\n   Alice: %v  Bilski: %v  Mayo: %v  Myriad: %v",
				alice, doc["bilskiIndicator"], doc["mayoIndicator"], doc["myriadIndicator"])
		}
		fmt.Println()
	}

	dctx.saveExample("search_office_action_rejections", map[string]string{"criteria": "patentApplicationNumber:12190351"}, result)
}

func demoSearchEnrichedCitations(dctx *DemoContext) {
	printSubHeader("SearchEnrichedCitations")

	result, err := dctx.Client.SearchEnrichedCitations(dctx.Ctx, "*:*", 0, 3)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Total enriched citation records: %d\n", result.Response.NumFound)
	fmt.Printf("Returned: %d\n", len(result.Response.Docs))

	for i, doc := range result.Response.Docs {
		if i >= 3 {
			break
		}
		fmt.Printf("\n%d. App: %v", i+1, doc["patentApplicationNumber"])
		if v, ok := doc["citedPatentNumber"]; ok {
			fmt.Printf("  Cited: %v", v)
		}
		fmt.Println()
	}

	dctx.saveExample("search_enriched_citations", map[string]string{"criteria": "*:*"}, result)
}

func demoGetOfficeActionCitationFields(dctx *DemoContext) {
	printSubHeader("GetOfficeActionCitationFields")

	result, err := dctx.Client.GetOfficeActionCitationFields(dctx.Ctx)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("API: %s  Fields: %d\n", result.APIKey, result.FieldCount)
	dctx.saveExample("get_office_action_citation_fields", map[string]string{}, result)
}

func demoGetOfficeActionRejectionFields(dctx *DemoContext) {
	printSubHeader("GetOfficeActionRejectionFields")

	result, err := dctx.Client.GetOfficeActionRejectionFields(dctx.Ctx)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("API: %s  Fields: %d\n", result.APIKey, result.FieldCount)
	dctx.saveExample("get_office_action_rejection_fields", map[string]string{}, result)
}

func demoGetEnrichedCitationFields(dctx *DemoContext) {
	printSubHeader("GetEnrichedCitationFields")

	result, err := dctx.Client.GetEnrichedCitationFields(dctx.Ctx)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("API: %s  Fields: %d\n", result.APIKey, result.FieldCount)
	dctx.saveExample("get_enriched_citation_fields", map[string]string{}, result)
}
