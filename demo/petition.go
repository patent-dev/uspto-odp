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

func demoSearchPetitions(ctx context.Context, client *odp.Client) {
	printSubHeader("SearchPetitions")

	result, err := client.SearchPetitions(ctx, "", 0, 5)
	if err != nil {
		printError(err)
		return
	}

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

	fmt.Printf("Downloaded data size: %d bytes\n", len(data))
	if len(data) > 500 {
		fmt.Printf("Preview:\n%s\n...\n", string(data[:500]))
	} else {
		fmt.Printf("Data:\n%s\n", string(data))
	}
}
