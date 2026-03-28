package main

import (
	"bytes"
	"fmt"
)

const (
	testTrademarkSN    = "97123456"
	testTrademarkDocID = "NOA20230322" // Notice of Abandonment from 2023-03-22
)

func demoTSDR(dctx *DemoContext) {
	printHeader("TSDR (Trademark Status & Document Retrieval) Demonstrations")

	demoGetTrademarkStatus(dctx)
	demoGetTrademarkDocuments(dctx)
	demoGetTrademarkDocumentInfo(dctx)
	demoDownloadTrademarkDocument(dctx)
	demoGetTrademarkLastUpdate(dctx)
	demoGetTrademarkMultiStatus(dctx)
}

func demoGetTrademarkStatus(dctx *DemoContext) {
	printSubHeader("GetTrademarkStatusJSON")

	result, err := dctx.Client.GetTrademarkStatusJSON(dctx.Ctx, testTrademarkSN)
	if err != nil {
		printError(err)
		return
	}

	if len(result.Trademarks) == 0 {
		fmt.Println("No trademark data found")
		return
	}

	fmt.Printf("Serial Number: %s\n", testTrademarkSN)
	fmt.Printf("Trademark entries: %d\n", len(result.Trademarks))

	tm := result.Trademarks[0]
	if status, ok := tm["status"].(map[string]any); ok {
		if markInfo, ok := status["markInfo"].(map[string]any); ok {
			if text, ok := markInfo["markText"]; ok {
				fmt.Printf("Mark Text: %v\n", text)
			}
		}
		if prosecution, ok := status["prosecution"].(map[string]any); ok {
			if statusText, ok := prosecution["statusText"]; ok {
				fmt.Printf("Status: %v\n", statusText)
			}
			if statusDate, ok := prosecution["statusDate"]; ok {
				fmt.Printf("Status Date: %v\n", statusDate)
			}
		}
	}

	dctx.saveExample("get_trademark_status", map[string]string{"serialNumber": testTrademarkSN}, result)
}

func demoGetTrademarkDocuments(dctx *DemoContext) {
	printSubHeader("GetTrademarkDocuments (XML)")

	result, err := dctx.Client.GetTrademarkDocumentsXML(dctx.Ctx, testTrademarkSN)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Serial Number: %s\n", testTrademarkSN)
	fmt.Printf("Document XML: %d bytes\n", len(result))

	dctx.saveRawExample("get_trademark_documents", map[string]string{"serialNumber": testTrademarkSN}, result, FormatXML)
}

func demoGetTrademarkDocumentInfo(dctx *DemoContext) {
	printSubHeader("GetTrademarkDocumentInfo")

	result, err := dctx.Client.GetTrademarkDocumentInfo(dctx.Ctx, testTrademarkSN, testTrademarkDocID)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Serial Number: %s  DocID: %s\n", testTrademarkSN, testTrademarkDocID)
	fmt.Printf("Document XML: %d bytes\n", len(result))

	dctx.saveRawExample("get_trademark_document_info", map[string]string{
		"serialNumber": testTrademarkSN,
		"docID":        testTrademarkDocID,
	}, result, FormatXML)
}

func demoDownloadTrademarkDocument(dctx *DemoContext) {
	printSubHeader("DownloadTrademarkDocument")

	var buf bytes.Buffer
	err := dctx.Client.DownloadTrademarkDocument(dctx.Ctx, testTrademarkSN, testTrademarkDocID, &buf)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Serial Number: %s  DocID: %s\n", testTrademarkSN, testTrademarkDocID)
	fmt.Printf("Downloaded: %d bytes (PDF)\n", buf.Len())
	// Don't save binary PDF as example - just verify download starts
}

func demoGetTrademarkMultiStatus(dctx *DemoContext) {
	printSubHeader("GetTrademarkMultiStatus")

	numbers := []string{"97123456", "90001000"}
	result, err := dctx.Client.GetTrademarkMultiStatus(dctx.Ctx, "sn", numbers)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Type: sn  Numbers: %v\n", numbers)
	if result.Size != nil {
		fmt.Printf("Size: %d\n", *result.Size)
	}
	if result.TransactionList != nil {
		fmt.Printf("Transactions: %d\n", len(*result.TransactionList))
	}

	dctx.saveExample("get_trademark_multi_status", map[string]string{
		"type":    "sn",
		"numbers": "97123456,90001000",
	}, result)
}

func demoGetTrademarkLastUpdate(dctx *DemoContext) {
	printSubHeader("GetTrademarkLastUpdate")

	result, err := dctx.Client.GetTrademarkLastUpdate(dctx.Ctx, testTrademarkSN)
	if err != nil {
		printError(err)
		return
	}

	fmt.Printf("Serial Number: %s\n", testTrademarkSN)
	if result != nil && result.CaseUpdateInfo != nil {
		for _, info := range *result.CaseUpdateInfo {
			if info.Name != nil && info.Value != nil {
				fmt.Printf("  %s: %s\n", *info.Name, *info.Value)
			}
		}
	}

	dctx.saveExample("get_trademark_last_update", map[string]string{"serialNumber": testTrademarkSN}, result)
}
