package odp

import (
	"strings"
	"testing"
)

// Sample patent grant XML (simplified but representative of ICE DTD 4.7 structure)
const sampleGrantXML = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE us-patent-grant SYSTEM "us-patent-grant-v47-2022-02-17.dtd">
<us-patent-grant lang="EN" dtd-version="v4.7 2022-02-17" file="US11234567-20220101.XML" status="PRODUCTION" id="us-patent-grant" country="US" date-produced="20220101" date-publ="20220101">
  <us-bibliographic-data-grant>
    <publication-reference>
      <document-id>
        <country>US</country>
        <doc-number>11234567</doc-number>
        <kind>B2</kind>
        <date>20220101</date>
      </document-id>
    </publication-reference>
    <application-reference appl-type="utility">
      <document-id>
        <country>US</country>
        <doc-number>17248024</doc-number>
        <date>20210101</date>
      </document-id>
    </application-reference>
    <invention-title id="d2e43">SYSTEM AND METHOD FOR ARTIFICIAL INTELLIGENCE</invention-title>
  </us-bibliographic-data-grant>
  <abstract id="abstract">
    <p id="p-0001" num="0001">A system for artificial intelligence processing includes a neural network architecture designed to optimize computational efficiency. The system comprises multiple layers of interconnected nodes.</p>
    <p id="p-0002" num="0002">The invention further provides methods for training the neural network using novel algorithms.</p>
  </abstract>
  <description id="description">
    <heading id="h-0001" level="1">TECHNICAL FIELD</heading>
    <p id="p-0003" num="0003">This invention relates to artificial intelligence systems.</p>
    <heading id="h-0002" level="1">BACKGROUND</heading>
    <p id="p-0004" num="0004">Traditional neural networks face computational challenges.</p>
  </description>
  <claims id="claims">
    <claim id="CLM-00001" num="1">
      <claim-text>1. A system comprising:
        <claim-text>a processor; and</claim-text>
        <claim-text>memory storing instructions that, when executed, cause the processor to perform operations.</claim-text>
      </claim-text>
    </claim>
    <claim id="CLM-00002" num="2">
      <claim-text>2. The system of claim 1, wherein the operations include:
        <claim-text>receiving input data;</claim-text>
        <claim-text>processing the input data using a neural network; and</claim-text>
        <claim-text>generating output predictions.</claim-text>
      </claim-text>
    </claim>
    <claim id="CLM-00003" num="3">
      <claim-text>3. The system of claim 2, wherein the neural network comprises multiple layers of interconnected nodes.</claim-text>
    </claim>
  </claims>
</us-patent-grant>`

// Sample patent application XML (simplified but representative of ICE DTD 4.6 structure)
const sampleApplicationXML = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE us-patent-application SYSTEM "us-patent-application-v46-2022-02-17.dtd">
<us-patent-application lang="EN" dtd-version="v4.6 2022-02-17" file="US17248024-20210101.XML" status="PRODUCTION" id="us-patent-application" country="US" date-produced="20210101" date-publ="20210701">
  <us-bibliographic-data-application>
    <application-reference appl-type="utility">
      <document-id>
        <country>US</country>
        <doc-number>17248024</doc-number>
        <date>20210101</date>
      </document-id>
    </application-reference>
    <invention-title id="d2e43">MACHINE LEARNING OPTIMIZATION METHOD</invention-title>
  </us-bibliographic-data-application>
  <abstract id="abstract">
    <p id="p-0001" num="0001">A method for optimizing machine learning models through automated hyperparameter tuning.</p>
  </abstract>
  <description id="description">
    <heading id="h-0001" level="1">FIELD OF THE INVENTION</heading>
    <p id="p-0002" num="0002">This application relates to machine learning systems.</p>
  </description>
  <claims id="claims">
    <claim id="CLM-00001" num="1">
      <claim-text>1. A method for optimizing a machine learning model, the method comprising:
        <claim-text>selecting hyperparameters;</claim-text>
        <claim-text>training the model; and</claim-text>
        <claim-text>evaluating performance.</claim-text>
      </claim-text>
    </claim>
  </claims>
</us-patent-application>`

// Invalid XML for error testing
const invalidXML = `<?xml version="1.0" encoding="UTF-8"?>
<invalid-root>
  <some-data>Invalid document structure</some-data>
</invalid-root>`

// Malformed XML
const malformedXML = `<?xml version="1.0" encoding="UTF-8"?>
<us-patent-grant>
  <unclosed-tag>
</us-patent-grant>`

func TestParseXML_Grant(t *testing.T) {
	doc, err := ParseXML([]byte(sampleGrantXML))
	if err != nil {
		t.Fatalf("Failed to parse grant XML: %v", err)
	}

	if doc == nil {
		t.Fatal("Parsed document is nil")
	}

	// Check document type
	if doc.GetDocumentType() != DocumentTypeGrant {
		t.Errorf("Expected DocumentTypeGrant, got %v", doc.GetDocumentType())
	}

	// Check grant is not nil
	if doc.Grant == nil {
		t.Fatal("Grant field is nil")
	}

	// Check application is nil
	if doc.Application != nil {
		t.Error("Application field should be nil for grant document")
	}

	// Check attributes
	if doc.Grant.Lang != "EN" {
		t.Errorf("Expected Lang 'EN', got '%s'", doc.Grant.Lang)
	}

	if doc.Grant.Country != "US" {
		t.Errorf("Expected Country 'US', got '%s'", doc.Grant.Country)
	}
}

func TestParseGrantXML(t *testing.T) {
	doc, err := ParseGrantXML([]byte(sampleGrantXML))
	if err != nil {
		t.Fatalf("Failed to parse grant XML: %v", err)
	}

	if doc == nil {
		t.Fatal("Parsed document is nil")
	}

	if doc.GetDocumentType() != DocumentTypeGrant {
		t.Errorf("Expected DocumentTypeGrant, got %v", doc.GetDocumentType())
	}

	if doc.Grant == nil {
		t.Fatal("Grant field is nil")
	}
}

func TestParseGrantXML_WrongType(t *testing.T) {
	// Try to parse application XML with ParseGrantXML
	_, err := ParseGrantXML([]byte(sampleApplicationXML))
	if err == nil {
		t.Error("Expected error when parsing application XML as grant")
	}
}

func TestParseApplicationXML(t *testing.T) {
	doc, err := ParseApplicationXML([]byte(sampleApplicationXML))
	if err != nil {
		t.Fatalf("Failed to parse application XML: %v", err)
	}

	if doc == nil {
		t.Fatal("Parsed document is nil")
	}

	if doc.GetDocumentType() != DocumentTypeApplication {
		t.Errorf("Expected DocumentTypeApplication, got %v", doc.GetDocumentType())
	}

	if doc.Application == nil {
		t.Fatal("Application field is nil")
	}
}

func TestParseApplicationXML_WrongType(t *testing.T) {
	// Try to parse grant XML with ParseApplicationXML
	_, err := ParseApplicationXML([]byte(sampleGrantXML))
	if err == nil {
		t.Error("Expected error when parsing grant XML as application")
	}
}

func TestParseXML_Application(t *testing.T) {
	doc, err := ParseXML([]byte(sampleApplicationXML))
	if err != nil {
		t.Fatalf("Failed to parse application XML: %v", err)
	}

	if doc == nil {
		t.Fatal("Parsed document is nil")
	}

	// Check document type
	if doc.GetDocumentType() != DocumentTypeApplication {
		t.Errorf("Expected DocumentTypeApplication, got %v", doc.GetDocumentType())
	}

	// Check application is not nil
	if doc.Application == nil {
		t.Fatal("Application field is nil")
	}

	// Check grant is nil
	if doc.Grant != nil {
		t.Error("Grant field should be nil for application document")
	}
}

func TestParseXML_InvalidDocument(t *testing.T) {
	doc, err := ParseXML([]byte(invalidXML))
	if err == nil {
		t.Error("Expected error for invalid document type")
	}

	if doc != nil {
		t.Error("Document should be nil for invalid XML")
	}

	if err != nil && !strings.Contains(err.Error(), "unrecognized XML document type") {
		t.Errorf("Expected unrecognized document error, got: %v", err)
	}
}

func TestParseXML_Malformed(t *testing.T) {
	_, err := ParseXML([]byte(malformedXML))
	if err == nil {
		t.Error("Expected error for malformed XML")
	}
}

func TestParseXML_Empty(t *testing.T) {
	_, err := ParseXML([]byte(""))
	if err == nil {
		t.Error("Expected error for empty XML")
	}
}

func TestGetTitle_Grant(t *testing.T) {
	doc, err := ParseXML([]byte(sampleGrantXML))
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	title := doc.GetTitle()
	expected := "SYSTEM AND METHOD FOR ARTIFICIAL INTELLIGENCE"
	if title != expected {
		t.Errorf("Expected title '%s', got '%s'", expected, title)
	}
}

func TestGetTitle_Application(t *testing.T) {
	doc, err := ParseXML([]byte(sampleApplicationXML))
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	title := doc.GetTitle()
	expected := "MACHINE LEARNING OPTIMIZATION METHOD"
	if title != expected {
		t.Errorf("Expected title '%s', got '%s'", expected, title)
	}
}

func TestGetTitle_NilDocument(t *testing.T) {
	doc := &XMLDocument{}
	title := doc.GetTitle()
	if title != "" {
		t.Errorf("Expected empty title for nil document, got '%s'", title)
	}
}

func TestGetAbstract(t *testing.T) {
	doc, err := ParseXML([]byte(sampleGrantXML))
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	abstract := doc.GetAbstract()
	if abstract == nil {
		t.Fatal("Abstract is nil")
	}

	if len(abstract.Paragraphs) != 2 {
		t.Errorf("Expected 2 paragraphs in abstract, got %d", len(abstract.Paragraphs))
	}
}

func TestExtractAbstractText(t *testing.T) {
	doc, err := ParseXML([]byte(sampleGrantXML))
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	abstract := doc.GetAbstract()
	if abstract == nil {
		t.Fatal("Abstract is nil")
	}

	text := abstract.ExtractAbstractText()
	if text == "" {
		t.Error("Abstract text is empty")
	}

	// Check that it contains expected content
	if !strings.Contains(text, "artificial intelligence") {
		t.Error("Abstract text does not contain expected content")
	}

	if !strings.Contains(text, "neural network") {
		t.Error("Abstract text does not contain 'neural network'")
	}
}

func TestExtractAbstractText_Nil(t *testing.T) {
	var abstract *Abstract
	text := abstract.ExtractAbstractText()
	if text != "" {
		t.Errorf("Expected empty text for nil abstract, got '%s'", text)
	}
}

func TestGetDescription(t *testing.T) {
	doc, err := ParseXML([]byte(sampleGrantXML))
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	desc := doc.GetDescription()
	if desc == nil {
		t.Fatal("Description is nil")
	}

	if len(desc.Headings) != 2 {
		t.Errorf("Expected 2 headings, got %d", len(desc.Headings))
	}

	if len(desc.Paragraphs) != 2 {
		t.Errorf("Expected 2 paragraphs, got %d", len(desc.Paragraphs))
	}
}

func TestExtractDescriptionText(t *testing.T) {
	doc, err := ParseXML([]byte(sampleGrantXML))
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	desc := doc.GetDescription()
	if desc == nil {
		t.Fatal("Description is nil")
	}

	text := desc.ExtractDescriptionText()
	if text == "" {
		t.Error("Description text is empty")
	}

	// Check headings are included
	if !strings.Contains(text, "TECHNICAL FIELD") {
		t.Error("Description text does not contain 'TECHNICAL FIELD' heading")
	}

	if !strings.Contains(text, "BACKGROUND") {
		t.Error("Description text does not contain 'BACKGROUND' heading")
	}

	// Check content is included
	if !strings.Contains(text, "artificial intelligence systems") {
		t.Error("Description text does not contain expected content")
	}
}

func TestExtractDescriptionText_Nil(t *testing.T) {
	var desc *Description
	text := desc.ExtractDescriptionText()
	if text != "" {
		t.Errorf("Expected empty text for nil description, got '%s'", text)
	}
}

func TestGetClaims(t *testing.T) {
	doc, err := ParseXML([]byte(sampleGrantXML))
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	claims := doc.GetClaims()
	if claims == nil {
		t.Fatal("Claims is nil")
	}

	if len(claims.ClaimList) != 3 {
		t.Errorf("Expected 3 claims, got %d", len(claims.ClaimList))
	}
}

func TestExtractClaimText_Simple(t *testing.T) {
	doc, err := ParseXML([]byte(sampleApplicationXML))
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	claims := doc.GetClaims()
	if claims == nil || len(claims.ClaimList) == 0 {
		t.Fatal("No claims found")
	}

	claim := claims.ClaimList[0]
	text := claim.ExtractClaimText()

	if text == "" {
		t.Error("Claim text is empty")
	}

	// Check that nested elements are concatenated
	if !strings.Contains(text, "method for optimizing") {
		t.Error("Claim text does not contain expected content")
	}

	if !strings.Contains(text, "selecting hyperparameters") {
		t.Error("Claim text does not contain nested element text")
	}
}

func TestExtractClaimText_Nested(t *testing.T) {
	doc, err := ParseXML([]byte(sampleGrantXML))
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	claims := doc.GetClaims()
	if claims == nil || len(claims.ClaimList) == 0 {
		t.Fatal("No claims found")
	}

	// Test first claim with nested structure
	claim := claims.ClaimList[0]
	text := claim.ExtractClaimText()

	if text == "" {
		t.Error("Claim text is empty")
	}

	// Check that all parts are included
	if !strings.Contains(text, "processor") {
		t.Error("Claim text does not contain 'processor'")
	}

	if !strings.Contains(text, "memory storing instructions") {
		t.Error("Claim text does not contain nested element")
	}

	// Test second claim with more nesting
	if len(claims.ClaimList) > 1 {
		claim2 := claims.ClaimList[1]
		text2 := claim2.ExtractClaimText()

		if !strings.Contains(text2, "receiving input data") {
			t.Error("Claim 2 text does not contain expected nested content")
		}

		if !strings.Contains(text2, "generating output predictions") {
			t.Error("Claim 2 text does not contain deeply nested content")
		}
	}
}

func TestExtractClaimText_Nil(t *testing.T) {
	var claim *Claim
	text := claim.ExtractClaimText()
	if text != "" {
		t.Errorf("Expected empty text for nil claim, got '%s'", text)
	}
}

func TestExtractAllClaimsText(t *testing.T) {
	doc, err := ParseXML([]byte(sampleGrantXML))
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	claims := doc.GetClaims()
	if claims == nil {
		t.Fatal("Claims is nil")
	}

	allTexts := claims.ExtractAllClaimsText()
	if len(allTexts) != 3 {
		t.Errorf("Expected 3 claim texts, got %d", len(allTexts))
	}

	for i, text := range allTexts {
		if text == "" {
			t.Errorf("Claim %d text is empty", i+1)
		}
	}
}

func TestExtractAllClaimsText_Nil(t *testing.T) {
	var claims *Claims
	texts := claims.ExtractAllClaimsText()
	if texts != nil {
		t.Errorf("Expected nil for nil claims, got %v", texts)
	}
}

func TestExtractAllClaimsTextFormatted(t *testing.T) {
	doc, err := ParseXML([]byte(sampleGrantXML))
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	claims := doc.GetClaims()
	if claims == nil {
		t.Fatal("Claims is nil")
	}

	formatted := claims.ExtractAllClaimsTextFormatted()
	if formatted == "" {
		t.Error("Formatted claims text is empty")
	}

	// Check that claim numbers are present
	if !strings.Contains(formatted, "CLAIM 1:") {
		t.Error("Formatted text does not contain 'CLAIM 1:'")
	}

	if !strings.Contains(formatted, "CLAIM 2:") {
		t.Error("Formatted text does not contain 'CLAIM 2:'")
	}

	if !strings.Contains(formatted, "CLAIM 3:") {
		t.Error("Formatted text does not contain 'CLAIM 3:'")
	}

	// Check that claims are separated
	claimSections := strings.Split(formatted, "CLAIM ")
	if len(claimSections) != 4 { // Empty string before first CLAIM, then 3 claims
		t.Errorf("Expected 4 sections (including empty), got %d", len(claimSections))
	}
}

func TestExtractAllClaimsTextFormatted_Nil(t *testing.T) {
	var claims *Claims
	formatted := claims.ExtractAllClaimsTextFormatted()
	if formatted != "" {
		t.Errorf("Expected empty string for nil claims, got '%s'", formatted)
	}
}

// Test comprehensive text extraction
func TestComprehensiveTextExtraction(t *testing.T) {
	doc, err := ParseXML([]byte(sampleGrantXML))
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	// Extract all sections
	title := doc.GetTitle()
	abstract := doc.GetAbstract()
	description := doc.GetDescription()
	claims := doc.GetClaims()

	// Verify all sections exist
	if title == "" {
		t.Error("Title is empty")
	}

	if abstract == nil {
		t.Error("Abstract is nil")
	}

	if description == nil {
		t.Error("Description is nil")
	}

	if claims == nil {
		t.Error("Claims is nil")
	}

	// Extract full text from each section
	abstractText := abstract.ExtractAbstractText()
	descriptionText := description.ExtractDescriptionText()
	claimsText := claims.ExtractAllClaimsTextFormatted()

	// Verify all text is extracted
	if abstractText == "" {
		t.Error("Abstract text is empty")
	}

	if descriptionText == "" {
		t.Error("Description text is empty")
	}

	if claimsText == "" {
		t.Error("Claims text is empty")
	}

	// Verify content quality
	totalLength := len(title) + len(abstractText) + len(descriptionText) + len(claimsText)
	if totalLength < 500 {
		t.Errorf("Total extracted text seems too short (%d bytes), may indicate parsing issues", totalLength)
	}
}

// Benchmark tests
func BenchmarkParseXML_Grant(b *testing.B) {
	data := []byte(sampleGrantXML)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ParseXML(data)
		if err != nil {
			b.Fatalf("Failed to parse: %v", err)
		}
	}
}

func BenchmarkExtractClaimText(b *testing.B) {
	doc, err := ParseXML([]byte(sampleGrantXML))
	if err != nil {
		b.Fatalf("Failed to parse: %v", err)
	}

	claims := doc.GetClaims()
	if claims == nil || len(claims.ClaimList) == 0 {
		b.Fatal("No claims found")
	}

	claim := claims.ClaimList[0]
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = claim.ExtractClaimText()
	}
}

func BenchmarkExtractAllClaimsText(b *testing.B) {
	doc, err := ParseXML([]byte(sampleGrantXML))
	if err != nil {
		b.Fatalf("Failed to parse: %v", err)
	}

	claims := doc.GetClaims()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = claims.ExtractAllClaimsText()
	}
}
