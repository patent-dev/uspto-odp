package odp

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// XMLDocument represents either a patent grant or application XML document
type XMLDocument struct {
	// Grant fields (us-patent-grant)
	Grant *PatentGrant `xml:"us-patent-grant"`
	// Application fields (us-patent-application)
	Application *PatentApplication `xml:"us-patent-application"`
}

// PatentGrant represents a patent grant XML document (ICE DTD 4.7)
type PatentGrant struct {
	XMLName      xml.Name      `xml:"us-patent-grant"`
	Lang         string        `xml:"lang,attr"`
	DTDVersion   string        `xml:"dtd-version,attr"`
	File         string        `xml:"file,attr"`
	Status       string        `xml:"status,attr"`
	ID           string        `xml:"id,attr"`
	Country      string        `xml:"country,attr"`
	DateProduced string        `xml:"date-produced,attr"`
	DatePubl     string        `xml:"date-publ,attr"`
	Bibliography *Bibliography `xml:"us-bibliographic-data-grant"`
	Abstract     *Abstract     `xml:"abstract"`
	DrawingsInfo *DrawingsInfo `xml:"drawings"`
	Description  *Description  `xml:"description"`
	Claims       *Claims       `xml:"claims"`
}

// PatentApplication represents a patent application XML document (ICE DTD 4.6)
type PatentApplication struct {
	XMLName      xml.Name      `xml:"us-patent-application"`
	Lang         string        `xml:"lang,attr"`
	DTDVersion   string        `xml:"dtd-version,attr"`
	File         string        `xml:"file,attr"`
	Status       string        `xml:"status,attr"`
	ID           string        `xml:"id,attr"`
	Country      string        `xml:"country,attr"`
	DateProduced string        `xml:"date-produced,attr"`
	DatePubl     string        `xml:"date-publ,attr"`
	Bibliography *Bibliography `xml:"us-bibliographic-data-application"`
	Abstract     *Abstract     `xml:"abstract"`
	DrawingsInfo *DrawingsInfo `xml:"drawings"`
	Description  *Description  `xml:"description"`
	Claims       *Claims       `xml:"claims"`
}

// Bibliography contains bibliographic data
type Bibliography struct {
	PublicationReference *DocumentID `xml:"publication-reference>document-id"`
	ApplicationReference *DocumentID `xml:"application-reference>document-id"`
	InventionTitle       []Text      `xml:"invention-title"`
	// Add more fields as needed
}

// DocumentID represents document identification
type DocumentID struct {
	Country   string `xml:"country"`
	DocNumber string `xml:"doc-number"`
	Kind      string `xml:"kind"`
	Date      string `xml:"date"`
}

// Abstract represents the patent abstract
type Abstract struct {
	ID         string      `xml:"id,attr"`
	Lang       string      `xml:"lang,attr"`
	Paragraphs []Paragraph `xml:"p"`
}

// Description represents the detailed description section
type Description struct {
	ID         string      `xml:"id,attr"`
	Lang       string      `xml:"lang,attr"`
	Headings   []Heading   `xml:"heading"`
	Paragraphs []Paragraph `xml:"p"`
}

// Heading represents a section heading
type Heading struct {
	ID    string `xml:"id,attr"`
	Level string `xml:"level,attr"`
	Text  string `xml:",chardata"`
}

// Paragraph represents a text paragraph with possible nested elements
type Paragraph struct {
	ID   string `xml:"id,attr"`
	Num  string `xml:"num,attr"`
	Text string `xml:",chardata"`
	// Support for nested elements
	Sub []Sub    `xml:"sub"`
	Sup []Sup    `xml:"sup"`
	I   []Italic `xml:"i"`
	B   []Bold   `xml:"b"`
}

// Sub represents subscript text
type Sub struct {
	Text string `xml:",chardata"`
}

// Sup represents superscript text
type Sup struct {
	Text string `xml:",chardata"`
}

// Italic represents italic text
type Italic struct {
	Text string `xml:",chardata"`
}

// Bold represents bold text
type Bold struct {
	Text string `xml:",chardata"`
}

// DrawingsInfo contains drawing/figure information
type DrawingsInfo struct {
	ID      string   `xml:"id,attr"`
	Figures []Figure `xml:"figure"`
}

// Figure represents a drawing figure
type Figure struct {
	ID  string `xml:"id,attr"`
	Img Image  `xml:"img"`
}

// Image represents an image reference
type Image struct {
	ID         string `xml:"id,attr"`
	He         string `xml:"he,attr"`
	Wi         string `xml:"wi,attr"`
	File       string `xml:"file,attr"`
	Alt        string `xml:"alt,attr"`
	ImgContent string `xml:"img-content,attr"`
	ImgFormat  string `xml:"img-format,attr"`
}

// Claims represents the claims section
type Claims struct {
	ID        string  `xml:"id,attr"`
	Lang      string  `xml:"lang,attr"`
	ClaimList []Claim `xml:"claim"`
}

// Claim represents a single claim with possibly nested claim-text
type Claim struct {
	ID        string      `xml:"id,attr"`
	Num       string      `xml:"num,attr"`
	ClaimText []ClaimText `xml:"claim-text"`
}

// ClaimText represents claim text with support for nested claim-text elements
// This recursive structure handles the hierarchical nature of claim dependencies
type ClaimText struct {
	ID   string `xml:"id,attr"`
	Text string `xml:",chardata"`
	// Nested claim-text elements for dependent claims
	NestedClaims []ClaimText `xml:"claim-text"`
	// Support for formatting elements
	Sub []Sub    `xml:"sub"`
	Sup []Sup    `xml:"sup"`
	I   []Italic `xml:"i"`
	B   []Bold   `xml:"b"`
}

// Text represents simple text with language attribute
type Text struct {
	ID   string `xml:"id,attr"`
	Lang string `xml:"lang,attr"`
	Text string `xml:",chardata"`
}

// DocumentType identifies the type of XML document
type DocumentType int

const (
	DocumentTypeUnknown DocumentType = iota
	DocumentTypeGrant
	DocumentTypeApplication
)

// GetDocumentType returns the type of document (grant or application)
func (d *XMLDocument) GetDocumentType() DocumentType {
	if d.Grant != nil {
		return DocumentTypeGrant
	}
	if d.Application != nil {
		return DocumentTypeApplication
	}
	return DocumentTypeUnknown
}

// GetTitle returns the invention title from either grant or application
func (d *XMLDocument) GetTitle() string {
	var bib *Bibliography
	switch d.GetDocumentType() {
	case DocumentTypeGrant:
		bib = d.Grant.Bibliography
	case DocumentTypeApplication:
		bib = d.Application.Bibliography
	default:
		return ""
	}

	if bib == nil || len(bib.InventionTitle) == 0 {
		return ""
	}
	return strings.TrimSpace(bib.InventionTitle[0].Text)
}

// GetAbstract returns the abstract section
func (d *XMLDocument) GetAbstract() *Abstract {
	switch d.GetDocumentType() {
	case DocumentTypeGrant:
		return d.Grant.Abstract
	case DocumentTypeApplication:
		return d.Application.Abstract
	default:
		return nil
	}
}

// GetDescription returns the description section
func (d *XMLDocument) GetDescription() *Description {
	switch d.GetDocumentType() {
	case DocumentTypeGrant:
		return d.Grant.Description
	case DocumentTypeApplication:
		return d.Application.Description
	default:
		return nil
	}
}

// GetClaims returns the claims section
func (d *XMLDocument) GetClaims() *Claims {
	switch d.GetDocumentType() {
	case DocumentTypeGrant:
		return d.Grant.Claims
	case DocumentTypeApplication:
		return d.Application.Claims
	default:
		return nil
	}
}

// ExtractAbstractText extracts full text from the abstract
func (a *Abstract) ExtractAbstractText() string {
	if a == nil {
		return ""
	}

	var builder strings.Builder
	for i, p := range a.Paragraphs {
		if i > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(extractParagraphText(&p))
	}

	return strings.TrimSpace(builder.String())
}

// ExtractDescriptionText extracts full text from the description
func (d *Description) ExtractDescriptionText() string {
	if d == nil {
		return ""
	}

	var builder strings.Builder
	lastWasHeading := false

	for _, h := range d.Headings {
		if builder.Len() > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(strings.TrimSpace(h.Text))
		lastWasHeading = true
	}

	for _, p := range d.Paragraphs {
		if builder.Len() > 0 {
			if lastWasHeading {
				builder.WriteString("\n\n")
			} else {
				builder.WriteString("\n\n")
			}
		}
		builder.WriteString(extractParagraphText(&p))
		lastWasHeading = false
	}

	return strings.TrimSpace(builder.String())
}

// extractParagraphText extracts text from a paragraph, handling nested elements
func extractParagraphText(p *Paragraph) string {
	if p == nil {
		return ""
	}

	return strings.TrimSpace(p.Text)
}

// ExtractClaimText recursively extracts full text from a claim
func (c *Claim) ExtractClaimText() string {
	if c == nil {
		return ""
	}

	return extractClaimTextRecursive(c.ClaimText, 0)
}

// extractClaimTextRecursive recursively extracts text from nested claim-text elements
func extractClaimTextRecursive(claimTexts []ClaimText, depth int) string {
	var builder strings.Builder

	for _, ct := range claimTexts {
		// Add the text at this level
		text := strings.TrimSpace(ct.Text)
		if text != "" {
			if builder.Len() > 0 {
				builder.WriteString(" ")
			}
			builder.WriteString(text)
		}

		// Recursively process nested claim-text elements
		if len(ct.NestedClaims) > 0 {
			nested := extractClaimTextRecursive(ct.NestedClaims, depth+1)
			if nested != "" {
				if builder.Len() > 0 {
					builder.WriteString(" ")
				}
				builder.WriteString(nested)
			}
		}
	}

	return strings.TrimSpace(builder.String())
}

// ExtractAllClaimsText extracts text from all claims
func (c *Claims) ExtractAllClaimsText() []string {
	if c == nil || len(c.ClaimList) == 0 {
		return nil
	}

	result := make([]string, 0, len(c.ClaimList))
	for _, claim := range c.ClaimList {
		text := claim.ExtractClaimText()
		if text != "" {
			result = append(result, text)
		}
	}

	return result
}

// ExtractAllClaimsTextFormatted returns formatted claim text with claim numbers
func (c *Claims) ExtractAllClaimsTextFormatted() string {
	if c == nil || len(c.ClaimList) == 0 {
		return ""
	}

	var builder strings.Builder
	for i, claim := range c.ClaimList {
		claimNum := i + 1
		if claim.Num != "" {
			claimNum = 0 // Use the claim number from XML if available
			// Try to parse it
			if _, err := fmt.Sscanf(claim.Num, "%d", &claimNum); err != nil {
				claimNum = i + 1
			}
		}

		if i > 0 {
			builder.WriteString("\n\n")
		}

		fmt.Fprintf(&builder, "CLAIM %d:\n%s", claimNum, claim.ExtractClaimText())
	}

	return builder.String()
}

// DownloadXML downloads and parses an XML document from a given URL
// If you know the document type, use DownloadXMLWithType for better performance
func (c *Client) DownloadXML(ctx context.Context, url string) (*XMLDocument, error) {
	return c.DownloadXMLWithType(ctx, url, DocumentTypeUnknown)
}

// DownloadXMLWithType downloads and parses an XML document with a known type hint
func (c *Client) DownloadXMLWithType(ctx context.Context, url string, expectedType DocumentType) (*XMLDocument, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.config.UserAgent)
	if c.config.APIKey != "" {
		req.Header.Set("X-API-Key", c.config.APIKey)
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading XML: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	xmlData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading XML data: %w", err)
	}

	return ParseXMLWithType(xmlData, expectedType)
}

// ParseXML parses XML data and auto-detects the document type
func ParseXML(data []byte) (*XMLDocument, error) {
	return ParseXMLWithType(data, DocumentTypeUnknown)
}

// ParseGrantXML parses patent grant XML data (us-patent-grant)
func ParseGrantXML(data []byte) (*XMLDocument, error) {
	return ParseXMLWithType(data, DocumentTypeGrant)
}

// ParseApplicationXML parses patent application XML data (us-patent-application)
func ParseApplicationXML(data []byte) (*XMLDocument, error) {
	return ParseXMLWithType(data, DocumentTypeApplication)
}

// ParseXMLWithType parses XML data with a known document type hint
func ParseXMLWithType(data []byte, expectedType DocumentType) (*XMLDocument, error) {
	var doc XMLDocument

	// If we know the type, parse directly
	switch expectedType {
	case DocumentTypeGrant:
		var grant PatentGrant
		if err := xml.Unmarshal(data, &grant); err != nil {
			return nil, fmt.Errorf("failed to parse as patent grant: %w", err)
		}
		if grant.XMLName.Local != "us-patent-grant" {
			return nil, fmt.Errorf("expected us-patent-grant root element, got %s", grant.XMLName.Local)
		}
		doc.Grant = &grant
		return &doc, nil

	case DocumentTypeApplication:
		var app PatentApplication
		if err := xml.Unmarshal(data, &app); err != nil {
			return nil, fmt.Errorf("failed to parse as patent application: %w", err)
		}
		if app.XMLName.Local != "us-patent-application" {
			return nil, fmt.Errorf("expected us-patent-application root element, got %s", app.XMLName.Local)
		}
		doc.Application = &app
		return &doc, nil

	case DocumentTypeUnknown:
		// Try parsing as patent grant first (more common)
		var grant PatentGrant
		if err := xml.Unmarshal(data, &grant); err == nil && grant.XMLName.Local == "us-patent-grant" {
			doc.Grant = &grant
			return &doc, nil
		}

		// Try parsing as patent application
		var app PatentApplication
		if err := xml.Unmarshal(data, &app); err == nil && app.XMLName.Local == "us-patent-application" {
			doc.Application = &app
			return &doc, nil
		}

		// If neither worked, return error
		return nil, fmt.Errorf("unrecognized XML document type (expected us-patent-grant or us-patent-application)")
	}

	return nil, fmt.Errorf("invalid document type: %v", expectedType)
}

// GetXMLURLForApplication retrieves the XML URL and document type for a patent
func (c *Client) GetXMLURLForApplication(ctx context.Context, patentNumber string) (string, DocumentType, error) {
	resp, err := c.GetPatent(ctx, patentNumber)
	if err != nil {
		return "", DocumentTypeUnknown, fmt.Errorf("failed to get patent data: %w", err)
	}

	if resp.PatentFileWrapperDataBag == nil || len(*resp.PatentFileWrapperDataBag) == 0 {
		return "", DocumentTypeUnknown, fmt.Errorf("no patent data found")
	}

	// Get the first patent record
	patentData := (*resp.PatentFileWrapperDataBag)[0]

	// Try to get grant XML first (if patent has been granted)
	if patentData.GrantDocumentMetaData != nil {
		// Convert to JSON and back to extract the fileLocationURI
		// The generated type uses interface{} for this field, so we use JSON
		// marshaling as a type-safe way to access nested string fields
		jsonData, err := json.Marshal(patentData.GrantDocumentMetaData)
		if err == nil {
			var meta map[string]interface{}
			if err := json.Unmarshal(jsonData, &meta); err == nil {
				if uri, ok := meta["fileLocationURI"].(string); ok && uri != "" {
					return uri, DocumentTypeGrant, nil
				}
			}
		}
	}

	// Try to get application XML
	if patentData.ApplicationMetaData != nil {
		// Convert to JSON and back to extract the fileLocationURI
		// The generated type uses interface{} for this field, so we use JSON
		// marshaling as a type-safe way to access nested string fields
		jsonData, err := json.Marshal(patentData.ApplicationMetaData)
		if err == nil {
			var meta map[string]interface{}
			if err := json.Unmarshal(jsonData, &meta); err == nil {
				if uri, ok := meta["fileLocationURI"].(string); ok && uri != "" {
					return uri, DocumentTypeApplication, nil
				}
			}
		}
	}

	return "", DocumentTypeUnknown, fmt.Errorf("no XML URL found in patent data")
}

// GetPatentXML retrieves and parses the XML document for a patent
// Accepts application numbers, grant numbers, or publication numbers
func (c *Client) GetPatentXML(ctx context.Context, patentNumber string) (*XMLDocument, error) {
	xmlURL, docType, err := c.GetXMLURLForApplication(ctx, patentNumber)
	if err != nil {
		return nil, err
	}

	return c.DownloadXMLWithType(ctx, xmlURL, docType)
}
