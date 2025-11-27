package main

import (
	"context"
	"fmt"

	odp "github.com/patent-dev/uspto-odp"
	"github.com/patent-dev/uspto-odp/generated"
)

func demoPetition(ctx context.Context, client *odp.Client) {
	printHeader("Petition API Demonstrations")

	demoSearchPetitions(ctx, client)
	demoGetPetitionDecision(ctx, client)
	demoSearchPetitionsDownload(ctx, client)
}

// demoPetitionWithContext runs all Petition demos with optional example saving
func demoPetitionWithContext(dctx *DemoContext) {
	printHeader("Petition API Demonstrations")

	demoSearchPetitionsCtx(dctx)
	demoGetPetitionDecisionCtx(dctx)
	demoSearchPetitionsDownloadCtx(dctx)
}

func demoSearchPetitions(ctx context.Context, client *odp.Client) {
	printSubHeader("SearchPetitions")

	result, err := client.SearchPetitions(ctx, "", 0, 5)
	if err != nil {
		printError(err)
		return
	}

	printPetitionSearchResult(result)
}

func demoSearchPetitionsCtx(dctx *DemoContext) {
	printSubHeader("SearchPetitions")

	result, err := dctx.Client.SearchPetitions(dctx.Ctx, "", 0, 5)
	if err != nil {
		printError(err)
		return
	}

	dctx.savePatentExample("search_petitions", map[string]string{"query": "", "offset": "0", "limit": "5"}, result)
	printPetitionSearchResult(result)
}

func printPetitionSearchResult(result *generated.PetitionDecisionResponseBag) {
	if result.Count != nil {
		fmt.Printf("Total results: %d\n", *result.Count)
	}

	if result.PetitionDecisionDataBag != nil {
		fmt.Printf("Returned: %d petitions\n", len(*result.PetitionDecisionDataBag))
		for i, petition := range *result.PetitionDecisionDataBag {
			if i >= 3 {
				break
			}
			fmt.Printf("\n%d. ", i+1)
			if petition.PetitionDecisionRecordIdentifier != nil {
				fmt.Printf("Record ID: %s", *petition.PetitionDecisionRecordIdentifier)
			}
			if petition.DecisionPetitionTypeCodeDescriptionText != nil {
				fmt.Printf("\n   Type: %s", *petition.DecisionPetitionTypeCodeDescriptionText)
			}
			if petition.DecisionDate != nil {
				fmt.Printf("\n   Date: %s", *petition.DecisionDate)
			}
			fmt.Println()
		}
	}
}

func demoGetPetitionDecision(ctx context.Context, client *odp.Client) {
	printSubHeader("GetPetitionDecision")

	searchResult, err := client.SearchPetitions(ctx, "", 0, 1)
	if err != nil {
		printError(err)
		return
	}

	var recordID string
	if searchResult.PetitionDecisionDataBag != nil && len(*searchResult.PetitionDecisionDataBag) > 0 {
		petition := (*searchResult.PetitionDecisionDataBag)[0]
		if petition.PetitionDecisionRecordIdentifier != nil {
			recordID = *petition.PetitionDecisionRecordIdentifier
		}
	}

	if recordID == "" {
		fmt.Println("No petition record ID found to demonstrate")
		return
	}

	result, err := client.GetPetitionDecision(ctx, recordID, true)
	if err != nil {
		printError(err)
		return
	}

	printPetitionDecisionResult(result)
}

func demoGetPetitionDecisionCtx(dctx *DemoContext) {
	printSubHeader("GetPetitionDecision")

	// First search to get a valid record ID
	searchResult, err := dctx.Client.SearchPetitions(dctx.Ctx, "", 0, 1)
	if err != nil {
		printError(err)
		return
	}

	var recordID string
	if searchResult.PetitionDecisionDataBag != nil && len(*searchResult.PetitionDecisionDataBag) > 0 {
		petition := (*searchResult.PetitionDecisionDataBag)[0]
		if petition.PetitionDecisionRecordIdentifier != nil {
			recordID = *petition.PetitionDecisionRecordIdentifier
		}
	}

	if recordID == "" {
		fmt.Println("No petition record ID found to demonstrate")
		return
	}

	result, err := dctx.Client.GetPetitionDecision(dctx.Ctx, recordID, true)
	if err != nil {
		printError(err)
		return
	}

	dctx.savePatentExample("get_petition_decision", map[string]string{"recordID": recordID, "includeDocuments": "true"}, result)
	printPetitionDecisionResult(result)
}

func printPetitionDecisionResult(result *generated.PetitionDecisionIdentifierResponseBag) {
	if result.PetitionDecisionDataBag != nil && len(*result.PetitionDecisionDataBag) > 0 {
		petition := (*result.PetitionDecisionDataBag)[0]
		if petition.PetitionDecisionRecordIdentifier != nil {
			printResult("Record ID", *petition.PetitionDecisionRecordIdentifier)
		}
		if petition.DecisionPetitionTypeCodeDescriptionText != nil {
			printResult("Type", *petition.DecisionPetitionTypeCodeDescriptionText)
		}
		if petition.DecisionDate != nil {
			printResult("Date", *petition.DecisionDate)
		}
		if petition.InventionTitle != nil {
			printResult("Title", *petition.InventionTitle)
		}
	}
}

func demoSearchPetitionsDownload(ctx context.Context, client *odp.Client) {
	printSubHeader("SearchPetitionsDownload")

	req := generated.PetitionDecisionDownloadRequest{
		Q: odp.StringPtr(""),
		Pagination: &generated.Pagination{
			Offset: odp.Int32Ptr(0),
			Limit:  odp.Int32Ptr(5),
		},
	}

	data, err := client.SearchPetitionsDownload(ctx, req)
	if err != nil {
		printError(err)
		return
	}

	printDownloadResult(data)
}

func demoSearchPetitionsDownloadCtx(dctx *DemoContext) {
	printSubHeader("SearchPetitionsDownload")

	req := generated.PetitionDecisionDownloadRequest{
		Q: odp.StringPtr(""),
		Pagination: &generated.Pagination{
			Offset: odp.Int32Ptr(0),
			Limit:  odp.Int32Ptr(5),
		},
	}

	data, err := dctx.Client.SearchPetitionsDownload(dctx.Ctx, req)
	if err != nil {
		printError(err)
		return
	}

	if dctx.Saver != nil {
		requestDesc := FormatRequestDescription("search_petitions_download", map[string]string{"query": "", "offset": "0", "limit": "5"})
		format := DetectFormat(data)
		if err := dctx.Saver.SaveExample("search_petitions_download", requestDesc, data, format); err != nil {
			fmt.Printf("Warning: failed to save example: %v\n", err)
		}
	}
	printDownloadResult(data)
}
