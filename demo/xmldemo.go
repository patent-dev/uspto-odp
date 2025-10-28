package main

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	odp "github.com/patent-dev/uspto-odp"
)

func demoXML(ctx context.Context, client *odp.Client, reader *bufio.Reader) {
	printHeader("Patent XML Full Text")

	fmt.Print("Enter patent number (e.g., US 11,646,472 B2, 17/248,024): ")
	patentNumber, _ := reader.ReadString('\n')
	patentNumber = strings.TrimSpace(patentNumber)

	if patentNumber == "" {
		fmt.Println("No patent number provided")
		return
	}

	fmt.Printf("\nFetching XML for: %s\n", patentNumber)

	doc, err := client.GetPatentXML(ctx, patentNumber)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	displayXMLDocument(doc)
}

func displayXMLDocument(doc *odp.XMLDocument) {
	docType := doc.GetDocumentType()
	fmt.Printf("Document type: ")
	switch docType {
	case odp.DocumentTypeGrant:
		fmt.Println("Grant")
	case odp.DocumentTypeApplication:
		fmt.Println("Application")
	default:
		fmt.Println("Unknown")
	}

	title := doc.GetTitle()
	if title != "" {
		fmt.Printf("\nTitle: %s\n", title)
	}

	abstract := doc.GetAbstract()
	if abstract != nil {
		displayAbstract(abstract)
	}

	claims := doc.GetClaims()
	if claims != nil {
		displayClaims(claims)
	}

	description := doc.GetDescription()
	if description != nil {
		descText := description.ExtractDescriptionText()
		fmt.Println("\n=== Description ===")
		fmt.Printf("Length: %d characters\n", len(descText))
		if len(descText) > 500 {
			fmt.Println(descText[:500] + "...")
		} else {
			fmt.Println(descText)
		}
	}

	displayStats(doc)
}

func displayAbstract(abstract *odp.Abstract) {
	abstractText := abstract.ExtractAbstractText()
	if abstractText != "" {
		fmt.Println("\n=== Abstract ===")
		fmt.Println(abstractText)
	}
}

func displayClaims(claims *odp.Claims) {
	if len(claims.ClaimList) == 0 {
		return
	}

	fmt.Printf("\n=== Claims (%d total) ===\n", len(claims.ClaimList))

	showCount := 3
	if len(claims.ClaimList) < showCount {
		showCount = len(claims.ClaimList)
	}

	for i := 0; i < showCount; i++ {
		claim := claims.ClaimList[i]
		claimText := claim.ExtractClaimText()
		fmt.Printf("\nClaim %d: %s\n", i+1, claimText)
	}

	if len(claims.ClaimList) > showCount {
		fmt.Printf("\n... and %d more claims\n", len(claims.ClaimList)-showCount)
	}
}

func displayStats(doc *odp.XMLDocument) {
	fmt.Println("\n=== Statistics ===")

	abstract := doc.GetAbstract()
	if abstract != nil {
		fmt.Printf("Abstract length: %d chars\n", len(abstract.ExtractAbstractText()))
	}

	claims := doc.GetClaims()
	if claims != nil {
		fmt.Printf("Number of claims: %d\n", len(claims.ClaimList))
	}

	description := doc.GetDescription()
	if description != nil {
		descText := description.ExtractDescriptionText()
		fmt.Printf("Description length: %d chars\n", len(descText))
	}
}
