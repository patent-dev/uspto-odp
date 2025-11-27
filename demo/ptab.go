package main

import (
	"context"
	"fmt"

	odp "github.com/patent-dev/uspto-odp"
	"github.com/patent-dev/uspto-odp/generated"
)

// demoPTAB runs basic PTAB demos (without example saving)
func demoPTAB(ctx context.Context, client *odp.Client) {
	dctx := &DemoContext{
		Client:   client,
		Ctx:      ctx,
		SkipSave: true,
	}
	demoPTABWithContext(dctx)
}

// demoPTABWithContext runs all PTAB demos with optional example saving
func demoPTABWithContext(dctx *DemoContext) {
	printHeader("PTAB (Patent Trial and Appeal Board) API - 19 Endpoints")

	// Trial Proceedings (3 endpoints)
	demoSearchTrialProceedings(dctx)
	demoGetTrialProceeding(dctx)
	demoSearchTrialProceedingsDownload(dctx)

	// Trial Decisions (4 endpoints)
	demoSearchTrialDecisions(dctx)
	demoGetTrialDecision(dctx)
	demoGetTrialDecisionsByTrialNumber(dctx)
	demoSearchTrialDecisionsDownload(dctx)

	// Trial Documents (4 endpoints)
	demoSearchTrialDocuments(dctx)
	demoGetTrialDocument(dctx)
	demoGetTrialDocumentsByTrialNumber(dctx)
	demoSearchTrialDocumentsDownload(dctx)

	// Appeal Decisions (4 endpoints)
	demoSearchAppealDecisions(dctx)
	demoGetAppealDecision(dctx)
	demoGetAppealDecisionsByAppealNumber(dctx)
	demoSearchAppealDecisionsDownload(dctx)

	// Interference Decisions (4 endpoints)
	demoSearchInterferenceDecisions(dctx)
	demoGetInterferenceDecision(dctx)
	demoGetInterferenceDecisionsByNumber(dctx)
	demoSearchInterferenceDecisionsDownload(dctx)
}

// Helper to save example if saver is configured
func (dctx *DemoContext) saveExample(name string, params map[string]string, response interface{}) {
	if dctx.SkipSave || dctx.Saver == nil {
		return
	}
	requestDesc := FormatRequestDescription(name, params)
	if err := dctx.Saver.SaveJSONExample(name, requestDesc, response); err != nil {
		fmt.Printf("  (Failed to save example: %v)\n", err)
	} else {
		fmt.Printf("  (Saved to examples/%s/)\n", name)
	}
}

func (dctx *DemoContext) saveRawExample(name string, params map[string]string, data []byte, format FileFormat) {
	if dctx.SkipSave || dctx.Saver == nil {
		return
	}
	requestDesc := FormatRequestDescription(name, params)
	if err := dctx.Saver.SaveExample(name, requestDesc, data, format); err != nil {
		fmt.Printf("  (Failed to save example: %v)\n", err)
	} else {
		fmt.Printf("  (Saved to examples/%s/)\n", name)
	}
}

// Trial Proceedings

func demoSearchTrialProceedings(dctx *DemoContext) {
	printSubHeader("SearchTrialProceedings")

	query := "trialMetaData.trialTypeCode:IPR"
	result, err := dctx.Client.SearchTrialProceedings(dctx.Ctx, query, 0, 5)
	if err != nil {
		printError(err)
		return
	}

	if result.Count != nil {
		fmt.Printf("Total IPR proceedings: %d\n", *result.Count)
	}

	if result.PatentTrialProceedingDataBag != nil {
		fmt.Printf("Returned: %d proceedings\n", len(*result.PatentTrialProceedingDataBag))
		for i, proc := range *result.PatentTrialProceedingDataBag {
			if i >= 3 {
				break
			}
			fmt.Printf("\n%d. ", i+1)
			if proc.TrialNumber != nil {
				fmt.Printf("Trial: %s", *proc.TrialNumber)
			}
			if proc.TrialMetaData != nil {
				if proc.TrialMetaData.TrialStatusCategory != nil {
					fmt.Printf("\n   Status: %s", *proc.TrialMetaData.TrialStatusCategory)
				}
				if proc.TrialMetaData.PetitionFilingDate != nil {
					fmt.Printf("\n   Filing Date: %s", *proc.TrialMetaData.PetitionFilingDate)
				}
			}
			if proc.PatentOwnerData != nil && proc.PatentOwnerData.PatentNumber != nil {
				fmt.Printf("\n   Patent: %s", *proc.PatentOwnerData.PatentNumber)
			}
			fmt.Println()
		}
	}

	dctx.saveExample("search_trial_proceedings", map[string]string{"query": query}, result)
}

func demoGetTrialProceeding(dctx *DemoContext) {
	printSubHeader("GetTrialProceeding")

	// First search to get a valid trial number
	result, err := dctx.Client.SearchTrialProceedings(dctx.Ctx, "", 0, 1)
	if err != nil || result.PatentTrialProceedingDataBag == nil || len(*result.PatentTrialProceedingDataBag) == 0 {
		fmt.Println("No trial proceedings found to demo")
		return
	}

	trialNumber := ""
	if (*result.PatentTrialProceedingDataBag)[0].TrialNumber != nil {
		trialNumber = *(*result.PatentTrialProceedingDataBag)[0].TrialNumber
	}
	if trialNumber == "" {
		fmt.Println("No trial number found")
		return
	}

	procResult, err := dctx.Client.GetTrialProceeding(dctx.Ctx, trialNumber)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Retrieved trial: %s\n", trialNumber)
	if procResult.PatentTrialProceedingDataBag != nil && len(*procResult.PatentTrialProceedingDataBag) > 0 {
		proc := (*procResult.PatentTrialProceedingDataBag)[0]
		if proc.TrialMetaData != nil && proc.TrialMetaData.TrialStatusCategory != nil {
			fmt.Printf("Status: %s\n", *proc.TrialMetaData.TrialStatusCategory)
		}
	}

	dctx.saveExample("get_trial_proceeding", map[string]string{"trialNumber": trialNumber}, procResult)
}

func demoSearchTrialProceedingsDownload(dctx *DemoContext) {
	printSubHeader("SearchTrialProceedingsDownload")

	req := generated.DownloadRequest{
		Q: odp.StringPtr("trialMetaData.trialTypeCode:IPR"),
	}
	data, err := dctx.Client.SearchTrialProceedingsDownload(dctx.Ctx, req)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Downloaded %d bytes\n", len(data))
	if len(data) > 0 {
		preview := string(data)
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		fmt.Printf("Preview:\n%s\n", preview)
	}

	dctx.saveRawExample("search_trial_proceedings_download", map[string]string{"query": "trialMetaData.trialTypeCode:IPR"}, data, DetectFormat(data))
}

// Trial Decisions

func demoSearchTrialDecisions(dctx *DemoContext) {
	printSubHeader("SearchTrialDecisions")

	result, err := dctx.Client.SearchTrialDecisions(dctx.Ctx, "", 0, 5)
	if err != nil {
		printError(err)
		return
	}

	if result.Count != nil {
		fmt.Printf("Total trial decisions: %d\n", *result.Count)
	}

	if result.PatentTrialDocumentDataBag != nil {
		fmt.Printf("Returned: %d decisions\n", len(*result.PatentTrialDocumentDataBag))
		for i, doc := range *result.PatentTrialDocumentDataBag {
			if i >= 3 {
				break
			}
			fmt.Printf("\n%d. ", i+1)
			if doc.TrialNumber != nil {
				fmt.Printf("Trial: %s", *doc.TrialNumber)
			}
			if doc.DocumentData != nil {
				if doc.DocumentData.DocumentTitleText != nil {
					fmt.Printf("\n   Title: %s", *doc.DocumentData.DocumentTitleText)
				}
				if doc.DocumentData.DocumentFilingDate != nil {
					fmt.Printf("\n   Filing Date: %s", *doc.DocumentData.DocumentFilingDate)
				}
			}
			fmt.Println()
		}
	}

	dctx.saveExample("search_trial_decisions", map[string]string{"query": ""}, result)
}

func demoGetTrialDecision(dctx *DemoContext) {
	printSubHeader("GetTrialDecision")

	// First search to get a valid document identifier
	result, err := dctx.Client.SearchTrialDecisions(dctx.Ctx, "", 0, 1)
	if err != nil || result.PatentTrialDocumentDataBag == nil || len(*result.PatentTrialDocumentDataBag) == 0 {
		fmt.Println("No trial decisions found to demo")
		return
	}

	docID := ""
	if (*result.PatentTrialDocumentDataBag)[0].DocumentData != nil &&
		(*result.PatentTrialDocumentDataBag)[0].DocumentData.DocumentIdentifier != nil {
		docID = *(*result.PatentTrialDocumentDataBag)[0].DocumentData.DocumentIdentifier
	}
	if docID == "" {
		fmt.Println("No document identifier found")
		return
	}

	decResult, err := dctx.Client.GetTrialDecision(dctx.Ctx, docID)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Retrieved decision: %s\n", docID)
	dctx.saveExample("get_trial_decision", map[string]string{"documentIdentifier": docID}, decResult)
}

func demoGetTrialDecisionsByTrialNumber(dctx *DemoContext) {
	printSubHeader("GetTrialDecisionsByTrialNumber")

	// Get trial number from SearchTrialDecisions (not Proceedings) to ensure it has decisions
	result, err := dctx.Client.SearchTrialDecisions(dctx.Ctx, "", 0, 1)
	if err != nil || result.PatentTrialDocumentDataBag == nil || len(*result.PatentTrialDocumentDataBag) == 0 {
		fmt.Println("No trial decisions found to get trial number")
		return
	}

	trialNumber := ""
	if (*result.PatentTrialDocumentDataBag)[0].TrialNumber != nil {
		trialNumber = *(*result.PatentTrialDocumentDataBag)[0].TrialNumber
	}
	if trialNumber == "" {
		fmt.Println("No trial number found in decisions")
		return
	}

	decResult, err := dctx.Client.GetTrialDecisionsByTrialNumber(dctx.Ctx, trialNumber)
	if err != nil {
		printError(err)
		return
	}

	if decResult.Count != nil {
		fmt.Printf("Decisions for %s: %d\n", trialNumber, *decResult.Count)
	}
	dctx.saveExample("get_trial_decisions_by_trial_number", map[string]string{"trialNumber": trialNumber}, decResult)
}

func demoSearchTrialDecisionsDownload(dctx *DemoContext) {
	printSubHeader("SearchTrialDecisionsDownload")

	req := generated.DownloadRequest{
		Q: odp.StringPtr(""),
	}
	data, err := dctx.Client.SearchTrialDecisionsDownload(dctx.Ctx, req)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Downloaded %d bytes\n", len(data))
	dctx.saveRawExample("search_trial_decisions_download", map[string]string{}, data, DetectFormat(data))
}

// Trial Documents

func demoSearchTrialDocuments(dctx *DemoContext) {
	printSubHeader("SearchTrialDocuments")

	result, err := dctx.Client.SearchTrialDocuments(dctx.Ctx, "", 0, 5)
	if err != nil {
		printError(err)
		return
	}

	if result.Count != nil {
		fmt.Printf("Total trial documents: %d\n", *result.Count)
	}

	if result.PatentTrialDocumentDataBag != nil {
		fmt.Printf("Returned: %d documents\n", len(*result.PatentTrialDocumentDataBag))
		for i, doc := range *result.PatentTrialDocumentDataBag {
			if i >= 3 {
				break
			}
			fmt.Printf("\n%d. ", i+1)
			if doc.TrialNumber != nil {
				fmt.Printf("Trial: %s", *doc.TrialNumber)
			}
			if doc.DocumentData != nil {
				if doc.DocumentData.DocumentName != nil {
					fmt.Printf("\n   Name: %s", *doc.DocumentData.DocumentName)
				}
				if doc.DocumentData.DocumentCategory != nil {
					fmt.Printf("\n   Category: %s", *doc.DocumentData.DocumentCategory)
				}
			}
			fmt.Println()
		}
	}

	dctx.saveExample("search_trial_documents", map[string]string{}, result)
}

func demoGetTrialDocument(dctx *DemoContext) {
	printSubHeader("GetTrialDocument")

	// First search to get a valid document identifier
	result, err := dctx.Client.SearchTrialDocuments(dctx.Ctx, "", 0, 1)
	if err != nil || result.PatentTrialDocumentDataBag == nil || len(*result.PatentTrialDocumentDataBag) == 0 {
		fmt.Println("No trial documents found to demo")
		return
	}

	docID := ""
	if (*result.PatentTrialDocumentDataBag)[0].DocumentData != nil &&
		(*result.PatentTrialDocumentDataBag)[0].DocumentData.DocumentIdentifier != nil {
		docID = *(*result.PatentTrialDocumentDataBag)[0].DocumentData.DocumentIdentifier
	}
	if docID == "" {
		fmt.Println("No document identifier found")
		return
	}

	docResult, err := dctx.Client.GetTrialDocument(dctx.Ctx, docID)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Retrieved document: %s\n", docID)
	dctx.saveExample("get_trial_document", map[string]string{"documentIdentifier": docID}, docResult)
}

func demoGetTrialDocumentsByTrialNumber(dctx *DemoContext) {
	printSubHeader("GetTrialDocumentsByTrialNumber")

	// First search to get a valid trial number
	result, err := dctx.Client.SearchTrialProceedings(dctx.Ctx, "", 0, 1)
	if err != nil || result.PatentTrialProceedingDataBag == nil || len(*result.PatentTrialProceedingDataBag) == 0 {
		fmt.Println("No trials found to demo")
		return
	}

	trialNumber := ""
	if (*result.PatentTrialProceedingDataBag)[0].TrialNumber != nil {
		trialNumber = *(*result.PatentTrialProceedingDataBag)[0].TrialNumber
	}
	if trialNumber == "" {
		fmt.Println("No trial number found")
		return
	}

	docResult, err := dctx.Client.GetTrialDocumentsByTrialNumber(dctx.Ctx, trialNumber)
	if err != nil {
		printError(err)
		return
	}

	if docResult.Count != nil {
		fmt.Printf("Documents for %s: %d\n", trialNumber, *docResult.Count)
	}
	dctx.saveExample("get_trial_documents_by_trial_number", map[string]string{"trialNumber": trialNumber}, docResult)
}

func demoSearchTrialDocumentsDownload(dctx *DemoContext) {
	printSubHeader("SearchTrialDocumentsDownload")

	req := generated.DownloadRequest{
		Q: odp.StringPtr(""),
	}
	data, err := dctx.Client.SearchTrialDocumentsDownload(dctx.Ctx, req)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Downloaded %d bytes\n", len(data))
	dctx.saveRawExample("search_trial_documents_download", map[string]string{}, data, DetectFormat(data))
}

// Appeal Decisions

func demoSearchAppealDecisions(dctx *DemoContext) {
	printSubHeader("SearchAppealDecisions")

	result, err := dctx.Client.SearchAppealDecisions(dctx.Ctx, "", 0, 5)
	if err != nil {
		printError(err)
		return
	}

	if result.Count != nil {
		fmt.Printf("Total appeal decisions: %d\n", *result.Count)
	}

	if result.PatentAppealDataBag != nil {
		fmt.Printf("Returned: %d decisions\n", len(*result.PatentAppealDataBag))
		for i, appeal := range *result.PatentAppealDataBag {
			if i >= 3 {
				break
			}
			fmt.Printf("\n%d. ", i+1)
			if appeal.AppealNumber != nil {
				fmt.Printf("Appeal: %s", *appeal.AppealNumber)
			}
			if appeal.DocumentData != nil {
				if appeal.DocumentData.DocumentName != nil {
					fmt.Printf("\n   Document: %s", *appeal.DocumentData.DocumentName)
				}
				if appeal.DocumentData.DocumentFilingDate != nil {
					fmt.Printf("\n   Filing Date: %s", *appeal.DocumentData.DocumentFilingDate)
				}
			}
			fmt.Println()
		}
	}

	dctx.saveExample("search_appeal_decisions", map[string]string{}, result)
}

func demoGetAppealDecision(dctx *DemoContext) {
	printSubHeader("GetAppealDecision")

	// First search to get a valid document identifier
	result, err := dctx.Client.SearchAppealDecisions(dctx.Ctx, "", 0, 1)
	if err != nil || result.PatentAppealDataBag == nil || len(*result.PatentAppealDataBag) == 0 {
		fmt.Println("No appeal decisions found to demo")
		return
	}

	docID := ""
	if (*result.PatentAppealDataBag)[0].DocumentData != nil &&
		(*result.PatentAppealDataBag)[0].DocumentData.DocumentIdentifier != nil {
		docID = *(*result.PatentAppealDataBag)[0].DocumentData.DocumentIdentifier
	}
	if docID == "" {
		fmt.Println("No document identifier found")
		return
	}

	decResult, err := dctx.Client.GetAppealDecision(dctx.Ctx, docID)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Retrieved appeal decision: %s\n", docID)
	dctx.saveExample("get_appeal_decision", map[string]string{"documentIdentifier": docID}, decResult)
}

func demoGetAppealDecisionsByAppealNumber(dctx *DemoContext) {
	printSubHeader("GetAppealDecisionsByAppealNumber")

	// First search to get a valid appeal number
	result, err := dctx.Client.SearchAppealDecisions(dctx.Ctx, "", 0, 1)
	if err != nil || result.PatentAppealDataBag == nil || len(*result.PatentAppealDataBag) == 0 {
		fmt.Println("No appeals found to demo")
		return
	}

	appealNumber := ""
	if (*result.PatentAppealDataBag)[0].AppealNumber != nil {
		appealNumber = *(*result.PatentAppealDataBag)[0].AppealNumber
	}
	if appealNumber == "" {
		fmt.Println("No appeal number found")
		return
	}

	decResult, err := dctx.Client.GetAppealDecisionsByAppealNumber(dctx.Ctx, appealNumber)
	if err != nil {
		printError(err)
		return
	}

	if decResult.Count != nil {
		fmt.Printf("Decisions for appeal %s: %d\n", appealNumber, *decResult.Count)
	}
	dctx.saveExample("get_appeal_decisions_by_appeal_number", map[string]string{"appealNumber": appealNumber}, decResult)
}

func demoSearchAppealDecisionsDownload(dctx *DemoContext) {
	printSubHeader("SearchAppealDecisionsDownload")

	req := generated.DownloadRequest{
		Q: odp.StringPtr(""),
	}
	data, err := dctx.Client.SearchAppealDecisionsDownload(dctx.Ctx, req)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Downloaded %d bytes\n", len(data))
	dctx.saveRawExample("search_appeal_decisions_download", map[string]string{}, data, DetectFormat(data))
}

// Interference Decisions

func demoSearchInterferenceDecisions(dctx *DemoContext) {
	printSubHeader("SearchInterferenceDecisions")

	result, err := dctx.Client.SearchInterferenceDecisions(dctx.Ctx, "", 0, 5)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Total interference decisions: %d\n", result.Count)
	fmt.Printf("Returned: %d decisions\n", len(result.PatentInterferenceDataBag))

	for i, interference := range result.PatentInterferenceDataBag {
		if i >= 3 {
			break
		}
		fmt.Printf("\n%d. Interference: %s", i+1, interference.InterferenceNumber)
		if interference.DocumentData != nil && interference.DocumentData.DocumentTitleText != nil {
			fmt.Printf("\n   Title: %s", *interference.DocumentData.DocumentTitleText)
		}
		if interference.DocumentData != nil && interference.DocumentData.DecisionIssueDate != nil {
			fmt.Printf("\n   Decision Date: %s", *interference.DocumentData.DecisionIssueDate)
		}
		fmt.Println()
	}

	dctx.saveExample("search_interference_decisions", map[string]string{}, result)
}

func demoGetInterferenceDecision(dctx *DemoContext) {
	printSubHeader("GetInterferenceDecision")

	// Search for interference decisions and find one with a document identifier
	result, err := dctx.Client.SearchInterferenceDecisions(dctx.Ctx, "", 0, 20)
	if err != nil || len(result.PatentInterferenceDataBag) == 0 {
		fmt.Println("No interference decisions found to demo")
		return
	}

	// Find first result with a document identifier
	docID := ""
	for _, interference := range result.PatentInterferenceDataBag {
		if interference.DocumentData != nil && interference.DocumentData.DocumentIdentifier != nil && *interference.DocumentData.DocumentIdentifier != "" {
			docID = *interference.DocumentData.DocumentIdentifier
			break
		}
	}
	if docID == "" {
		fmt.Println("No document identifier found in search results")
		return
	}

	decResult, err := dctx.Client.GetInterferenceDecision(dctx.Ctx, docID)
	if err != nil {
		printError(err)
		return
	}

	interferenceNum := ""
	if len(decResult.PatentInterferenceDataBag) > 0 {
		interferenceNum = decResult.PatentInterferenceDataBag[0].InterferenceNumber
	}
	fmt.Printf("Retrieved interference decision for %s: %s\n", interferenceNum, docID[:16]+"...")
	dctx.saveExample("get_interference_decision", map[string]string{"documentIdentifier": docID}, decResult)
}

func demoGetInterferenceDecisionsByNumber(dctx *DemoContext) {
	printSubHeader("GetInterferenceDecisionsByNumber")

	// First search to get a valid interference number
	result, err := dctx.Client.SearchInterferenceDecisions(dctx.Ctx, "", 0, 1)
	if err != nil || len(result.PatentInterferenceDataBag) == 0 {
		fmt.Println("No interferences found to demo")
		return
	}

	interferenceNumber := result.PatentInterferenceDataBag[0].InterferenceNumber

	decResult, err := dctx.Client.GetInterferenceDecisionsByNumber(dctx.Ctx, interferenceNumber)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Decisions for interference %s: %d\n", interferenceNumber, decResult.Count)
	dctx.saveExample("get_interference_decisions_by_number", map[string]string{"interferenceNumber": interferenceNumber}, decResult)
}

func demoSearchInterferenceDecisionsDownload(dctx *DemoContext) {
	printSubHeader("SearchInterferenceDecisionsDownload")

	req := generated.PatentDownloadRequest{
		Q: odp.StringPtr(""),
	}
	data, err := dctx.Client.SearchInterferenceDecisionsDownload(dctx.Ctx, req)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Downloaded %d bytes\n", len(data))
	dctx.saveRawExample("search_interference_decisions_download", map[string]string{}, data, DetectFormat(data))
}
