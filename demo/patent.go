package main

import (
	"context"
	"fmt"

	odp "github.com/patent-dev/uspto-odp"
	"github.com/patent-dev/uspto-odp/generated"
)

func demoPatent(ctx context.Context, client *odp.Client, patentNumber string) {
	printHeader("Patent API Demonstrations")
	fmt.Printf("Using patent: %s\n", patentNumber)

	demoSearchPatents(ctx, client)
	demoGetPatent(ctx, client, patentNumber)
	demoGetPatentMetaData(ctx, client, patentNumber)
	demoGetPatentAdjustment(ctx, client, patentNumber)
	demoGetPatentContinuity(ctx, client, patentNumber)
	demoGetPatentDocuments(ctx, client, patentNumber)
	demoGetPatentAssignment(ctx, client, patentNumber)
	demoGetPatentAssociatedDocuments(ctx, client, patentNumber)
	demoGetPatentAttorney(ctx, client, patentNumber)
	demoGetPatentForeignPriority(ctx, client, patentNumber)
	demoGetPatentTransactions(ctx, client, patentNumber)
	demoSearchPatentsDownload(ctx, client)
	demoGetStatusCodes(ctx, client)
}

func demoSearchPatents(ctx context.Context, client *odp.Client) {
	printSubHeader("SearchPatents")

	result, err := client.SearchPatents(ctx, "artificialIntelligence", 0, 5)
	if err != nil {
		printError(err)
		return
	}

	if result.Count != nil {
		fmt.Printf("Total results: %d\n", *result.Count)
	}

	if result.PatentFileWrapperDataBag != nil {
		fmt.Printf("Returned: %d patents\n", len(*result.PatentFileWrapperDataBag))
		for i, patent := range *result.PatentFileWrapperDataBag {
			if i >= 3 {
				break
			}
			fmt.Printf("\n%d. ", i+1)
			if patent.ApplicationNumberText != nil {
				fmt.Printf("App: %s", *patent.ApplicationNumberText)
			}
			if patent.ApplicationMetaData != nil && patent.ApplicationMetaData.InventionTitle != nil {
				fmt.Printf("\n   Title: %s", truncate(*patent.ApplicationMetaData.InventionTitle, 80))
			}
			fmt.Println()
		}
	}
}

func demoGetPatent(ctx context.Context, client *odp.Client, patentNumber string) {
	printSubHeader("GetPatent")

	result, err := client.GetPatent(ctx, patentNumber)
	if err != nil {
		printError(err)
		return
	}

	if result.PatentFileWrapperDataBag != nil && len(*result.PatentFileWrapperDataBag) > 0 {
		patent := (*result.PatentFileWrapperDataBag)[0]
		if patent.ApplicationNumberText != nil {
			printResult("Application Number", *patent.ApplicationNumberText)
		}
		if patent.ApplicationMetaData != nil {
			if patent.ApplicationMetaData.InventionTitle != nil {
				printResult("Title", truncate(*patent.ApplicationMetaData.InventionTitle, 60))
			}
			if patent.ApplicationMetaData.ApplicationTypeCategory != nil {
				printResult("Type", *patent.ApplicationMetaData.ApplicationTypeCategory)
			}
			if patent.ApplicationMetaData.FilingDate != nil {
				printResult("Filing Date", *patent.ApplicationMetaData.FilingDate)
			}
			if patent.ApplicationMetaData.ApplicationStatusDescriptionText != nil {
				printResult("Status", *patent.ApplicationMetaData.ApplicationStatusDescriptionText)
			}
		}
	}
}

func demoGetPatentMetaData(ctx context.Context, client *odp.Client, patentNumber string) {
	printSubHeader("GetPatentMetaData")

	appNumber, err := client.ResolvePatentNumber(ctx, patentNumber)
	if err != nil {
		printError(err)
		return
	}

	result, err := client.GetPatentMetaData(ctx, appNumber)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Metadata: %s\n", formatJSON(result))
}

func demoGetPatentAdjustment(ctx context.Context, client *odp.Client, patentNumber string) {
	printSubHeader("GetPatentAdjustment")

	appNumber, err := client.ResolvePatentNumber(ctx, patentNumber)
	if err != nil {
		printError(err)
		return
	}

	result, err := client.GetPatentAdjustment(ctx, appNumber)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Adjustment data: %s\n", formatJSON(result))
}

func demoGetPatentContinuity(ctx context.Context, client *odp.Client, patentNumber string) {
	printSubHeader("GetPatentContinuity")

	appNumber, err := client.ResolvePatentNumber(ctx, patentNumber)
	if err != nil {
		printError(err)
		return
	}

	result, err := client.GetPatentContinuity(ctx, appNumber)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Continuity data: %s\n", formatJSON(result))
}

func demoGetPatentDocuments(ctx context.Context, client *odp.Client, patentNumber string) {
	printSubHeader("GetPatentDocuments")

	appNumber, err := client.ResolvePatentNumber(ctx, patentNumber)
	if err != nil {
		printError(err)
		return
	}

	result, err := client.GetPatentDocuments(ctx, appNumber)
	if err != nil {
		printError(err)
		return
	}

	if result != nil && result.DocumentBag != nil {
		fmt.Printf("Total documents: %d\n", len(*result.DocumentBag))
		for i, doc := range *result.DocumentBag {
			if i >= 5 {
				fmt.Printf("... and %d more documents\n", len(*result.DocumentBag)-5)
				break
			}
			fmt.Printf("\n%d. ", i+1)
			if doc.DocumentIdentifier != nil {
				fmt.Printf("ID: %s", *doc.DocumentIdentifier)
			}
			if doc.DocumentCodeDescriptionText != nil {
				fmt.Printf("\n   Description: %s", truncate(*doc.DocumentCodeDescriptionText, 60))
			}
			if doc.OfficialDate != nil {
				fmt.Printf("\n   Date: %s", *doc.OfficialDate)
			}
			fmt.Println()
		}
	}
}

func demoGetPatentAssignment(ctx context.Context, client *odp.Client, patentNumber string) {
	printSubHeader("GetPatentAssignment")

	appNumber, err := client.ResolvePatentNumber(ctx, patentNumber)
	if err != nil {
		printError(err)
		return
	}

	result, err := client.GetPatentAssignment(ctx, appNumber)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Assignment data: %s\n", formatJSON(result))
}

func demoGetPatentAssociatedDocuments(ctx context.Context, client *odp.Client, patentNumber string) {
	printSubHeader("GetPatentAssociatedDocuments")

	appNumber, err := client.ResolvePatentNumber(ctx, patentNumber)
	if err != nil {
		printError(err)
		return
	}

	result, err := client.GetPatentAssociatedDocuments(ctx, appNumber)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Associated documents: %s\n", formatJSON(result))
}

func demoGetPatentAttorney(ctx context.Context, client *odp.Client, patentNumber string) {
	printSubHeader("GetPatentAttorney")

	appNumber, err := client.ResolvePatentNumber(ctx, patentNumber)
	if err != nil {
		printError(err)
		return
	}

	result, err := client.GetPatentAttorney(ctx, appNumber)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Attorney data: %s\n", formatJSON(result))
}

func demoGetPatentForeignPriority(ctx context.Context, client *odp.Client, patentNumber string) {
	printSubHeader("GetPatentForeignPriority")

	appNumber, err := client.ResolvePatentNumber(ctx, patentNumber)
	if err != nil {
		printError(err)
		return
	}

	result, err := client.GetPatentForeignPriority(ctx, appNumber)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Foreign priority data: %s\n", formatJSON(result))
}

func demoGetPatentTransactions(ctx context.Context, client *odp.Client, patentNumber string) {
	printSubHeader("GetPatentTransactions")

	appNumber, err := client.ResolvePatentNumber(ctx, patentNumber)
	if err != nil {
		printError(err)
		return
	}

	result, err := client.GetPatentTransactions(ctx, appNumber)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Transaction history: %s\n", formatJSON(result))
}

func demoSearchPatentsDownload(ctx context.Context, client *odp.Client) {
	printSubHeader("SearchPatentsDownload")

	req := generated.PatentDownloadRequest{
		Q: odp.StringPtr("artificialIntelligence"),
		Pagination: &generated.Pagination{
			Offset: odp.Int32Ptr(0),
			Limit:  odp.Int32Ptr(5),
		},
	}

	data, err := client.SearchPatentsDownload(ctx, req)
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

func demoGetStatusCodes(ctx context.Context, client *odp.Client) {
	printSubHeader("GetStatusCodes")

	result, err := client.GetStatusCodes(ctx)
	if err != nil {
		printError(err)
		return
	}

	if result.StatusCodeBag != nil {
		fmt.Printf("Total status codes: %d\n", len(*result.StatusCodeBag))
		for i, code := range *result.StatusCodeBag {
			if i >= 10 {
				fmt.Printf("... and %d more status codes\n", len(*result.StatusCodeBag)-10)
				break
			}
			if code.ApplicationStatusCode != nil {
				fmt.Printf("\n%d", *code.ApplicationStatusCode)
			}
			if code.ApplicationStatusDescriptionText != nil {
				fmt.Printf(": %s", truncate(*code.ApplicationStatusDescriptionText, 60))
			}
			fmt.Println()
		}
	}
}
