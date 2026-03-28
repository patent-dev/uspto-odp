package main

import (
	"fmt"

	odp "github.com/patent-dev/uspto-odp"
	"github.com/patent-dev/uspto-odp/generated"
)

// demoPatentWithContext runs all Patent demos with optional example saving
func demoPatentWithContext(dctx *DemoContext, patentNumber string) {
	printHeader("Patent API Demonstrations")
	fmt.Printf("Using patent: %s\n", patentNumber)

	// Resolve patent number once for all demos
	appNumber, err := dctx.Client.ResolvePatentNumber(dctx.Ctx, patentNumber)
	if err != nil {
		printError(err)
		return
	}

	// Use different patents for endpoints that need specific data
	// Patent 15000001 has: foreign priority, assignment
	// Patent 17248024 has: adjustment, continuity
	appNumberWithForeignPriority := "15000001"
	appNumberWithAdjustment := "17248024"

	demoSearchPatentsCtx(dctx)
	demoGetPatentCtx(dctx, patentNumber)
	demoGetPatentMetaDataCtx(dctx, appNumber)
	demoGetPatentAdjustmentCtx(dctx, appNumberWithAdjustment)
	demoGetPatentContinuityCtx(dctx, appNumberWithAdjustment)
	demoGetPatentDocumentsCtx(dctx, appNumber)
	demoGetPatentAssignmentCtx(dctx, appNumberWithForeignPriority)
	demoGetPatentAssociatedDocumentsCtx(dctx, appNumber)
	demoGetPatentAttorneyCtx(dctx, appNumber)
	demoGetPatentForeignPriorityCtx(dctx, appNumberWithForeignPriority)
	demoGetPatentTransactionsCtx(dctx, appNumber)
	demoSearchPatentsDownloadCtx(dctx)
	demoGetStatusCodesCtx(dctx)
}

// savePatentExample saves an example if saver is configured
func (dctx *DemoContext) savePatentExample(name string, params map[string]string, response interface{}) {
	if dctx.SkipSave || dctx.Saver == nil {
		return
	}
	requestDesc := FormatRequestDescription(name, params)
	data, err := marshalJSON(response)
	if err != nil {
		fmt.Printf("Warning: failed to marshal response for %s: %v\n", name, err)
		return
	}
	format := DetectFormat(data)
	if err := dctx.Saver.SaveExample(name, requestDesc, data, format); err != nil {
		fmt.Printf("Warning: failed to save example for %s: %v\n", name, err)
	}
}

func demoSearchPatentsCtx(dctx *DemoContext) {
	printSubHeader("SearchPatents")

	query := "artificialIntelligence"
	result, err := dctx.Client.SearchPatents(dctx.Ctx, query, 0, 5)
	if err != nil {
		printError(err)
		return
	}

	dctx.savePatentExample("search_patents", map[string]string{"query": query, "offset": "0", "limit": "5"}, result)
	printPatentSearchResult(result)
}

func printPatentSearchResult(result *generated.PatentDataResponse) {
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

func demoGetPatentCtx(dctx *DemoContext, patentNumber string) {
	printSubHeader("GetPatent")

	result, err := dctx.Client.GetPatent(dctx.Ctx, patentNumber)
	if err != nil {
		printError(err)
		return
	}

	dctx.savePatentExample("get_patent", map[string]string{"patentNumber": patentNumber}, result)
	printPatentResult(result)
}

func printPatentResult(result *generated.PatentDataResponse) {
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

func demoGetPatentMetaDataCtx(dctx *DemoContext, appNumber string) {
	printSubHeader("GetPatentMetaData")

	result, err := dctx.Client.GetPatentMetaData(dctx.Ctx, appNumber)
	if err != nil {
		printError(err)
		return
	}

	dctx.savePatentExample("get_patent_meta_data", map[string]string{"applicationNumber": appNumber}, result)
	fmt.Printf("Metadata: %s\n", formatJSON(result))
}

func demoGetPatentAdjustmentCtx(dctx *DemoContext, appNumber string) {
	printSubHeader("GetPatentAdjustment")

	result, err := dctx.Client.GetPatentAdjustment(dctx.Ctx, appNumber)
	if err != nil {
		printError(err)
		return
	}

	dctx.savePatentExample("get_patent_adjustment", map[string]string{"applicationNumber": appNumber}, result)
	fmt.Printf("Adjustment data: %s\n", formatJSON(result))
}

func demoGetPatentContinuityCtx(dctx *DemoContext, appNumber string) {
	printSubHeader("GetPatentContinuity")

	result, err := dctx.Client.GetPatentContinuity(dctx.Ctx, appNumber)
	if err != nil {
		printError(err)
		return
	}

	dctx.savePatentExample("get_patent_continuity", map[string]string{"applicationNumber": appNumber}, result)
	fmt.Printf("Continuity data: %s\n", formatJSON(result))
}

func demoGetPatentDocumentsCtx(dctx *DemoContext, appNumber string) {
	printSubHeader("GetPatentDocuments")

	result, err := dctx.Client.GetPatentDocuments(dctx.Ctx, appNumber)
	if err != nil {
		printError(err)
		return
	}

	dctx.savePatentExample("get_patent_documents", map[string]string{"applicationNumber": appNumber}, result)
	printPatentDocumentsResult(result)
}

func printPatentDocumentsResult(result *generated.DocumentBag) {
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

func demoGetPatentAssignmentCtx(dctx *DemoContext, appNumber string) {
	printSubHeader("GetPatentAssignment")

	result, err := dctx.Client.GetPatentAssignment(dctx.Ctx, appNumber)
	if err != nil {
		printError(err)
		return
	}

	dctx.savePatentExample("get_patent_assignment", map[string]string{"applicationNumber": appNumber}, result)
	fmt.Printf("Assignment data: %s\n", formatJSON(result))
}

func demoGetPatentAssociatedDocumentsCtx(dctx *DemoContext, appNumber string) {
	printSubHeader("GetPatentAssociatedDocuments")

	result, err := dctx.Client.GetPatentAssociatedDocuments(dctx.Ctx, appNumber)
	if err != nil {
		printError(err)
		return
	}

	dctx.savePatentExample("get_patent_associated_documents", map[string]string{"applicationNumber": appNumber}, result)
	fmt.Printf("Associated documents: %s\n", formatJSON(result))
}

func demoGetPatentAttorneyCtx(dctx *DemoContext, appNumber string) {
	printSubHeader("GetPatentAttorney")

	result, err := dctx.Client.GetPatentAttorney(dctx.Ctx, appNumber)
	if err != nil {
		printError(err)
		return
	}

	dctx.savePatentExample("get_patent_attorney", map[string]string{"applicationNumber": appNumber}, result)
	fmt.Printf("Attorney data: %s\n", formatJSON(result))
}

func demoGetPatentForeignPriorityCtx(dctx *DemoContext, appNumber string) {
	printSubHeader("GetPatentForeignPriority")

	result, err := dctx.Client.GetPatentForeignPriority(dctx.Ctx, appNumber)
	if err != nil {
		printError(err)
		return
	}

	dctx.savePatentExample("get_patent_foreign_priority", map[string]string{"applicationNumber": appNumber}, result)
	fmt.Printf("Foreign priority data: %s\n", formatJSON(result))
}

func demoGetPatentTransactionsCtx(dctx *DemoContext, appNumber string) {
	printSubHeader("GetPatentTransactions")

	result, err := dctx.Client.GetPatentTransactions(dctx.Ctx, appNumber)
	if err != nil {
		printError(err)
		return
	}

	dctx.savePatentExample("get_patent_transactions", map[string]string{"applicationNumber": appNumber}, result)
	fmt.Printf("Transaction history: %s\n", formatJSON(result))
}

func demoSearchPatentsDownloadCtx(dctx *DemoContext) {
	printSubHeader("SearchPatentsDownload")

	req := generated.PatentDownloadRequest{
		Q: odp.StringPtr("artificialIntelligence"),
		Pagination: &generated.Pagination{
			Offset: ptrInt32(0),
			Limit:  ptrInt32(5),
		},
	}

	data, err := dctx.Client.SearchPatentsDownload(dctx.Ctx, req)
	if err != nil {
		printError(err)
		return
	}

	if dctx.Saver != nil {
		requestDesc := FormatRequestDescription("search_patents_download", map[string]string{"query": "artificialIntelligence", "offset": "0", "limit": "5"})
		format := DetectFormat(data)
		if err := dctx.Saver.SaveExample("search_patents_download", requestDesc, data, format); err != nil {
			fmt.Printf("Warning: failed to save example: %v\n", err)
		}
	}
	printDownloadResult(data)
}

func printDownloadResult(data []byte) {
	fmt.Printf("Downloaded data size: %d bytes\n", len(data))
	if len(data) > 500 {
		fmt.Printf("Preview:\n%s\n...\n", string(data[:500]))
	} else {
		fmt.Printf("Data:\n%s\n", string(data))
	}
}

func demoGetStatusCodesCtx(dctx *DemoContext) {
	printSubHeader("GetStatusCodes")

	result, err := dctx.Client.GetStatusCodes(dctx.Ctx)
	if err != nil {
		printError(err)
		return
	}

	dctx.savePatentExample("get_status_codes", map[string]string{}, result)
	printStatusCodesResult(result)
}

func printStatusCodesResult(result *generated.StatusCodeSearchResponse) {
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
