package odp

import (
	"context"
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

// ClaimText represents the text of a claim, including any nested <claim-text>.
//
// A <claim-text> element mixes character data with inline child elements - most
// importantly <claim-ref>, which carries a dependent claim's reference to its parent
// (e.g. "claim 1" in "The system according to <claim-ref>claim 1</claim-ref>, ...").
// Go's struct-tag ",chardata" mapping concatenates only the character data directly
// under the element and silently drops the text of child elements, which would render
// that claim as "The system according to , ...". ClaimText is therefore parsed with a
// custom UnmarshalXML (see below) that walks tokens in document order.
//
// Text holds the fully flattened text of this element and all of its descendants
// (inline references and nested claim-text included). NestedClaims preserves the
// structural tree for callers that need the hierarchy.
type ClaimText struct {
	ID           string
	Text         string
	NestedClaims []ClaimText
}

// UnmarshalXML flattens a <claim-text> element into its complete, in-document-order
// text. It is the encoding/xml equivalent of itertext(): character data, the inner
// text of inline elements such as <claim-ref> and <figref>, and the text of nested
// <claim-text> are concatenated in the order they appear. Whitespace is preserved
// here as-is; callers collapse it at extraction time (see Claim.ExtractClaimText).
func (ct *ClaimText) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, attr := range start.Attr {
		if attr.Name.Local == "id" {
			ct.ID = attr.Value
		}
	}

	var buf strings.Builder
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.CharData:
			buf.Write(t)
		case xml.StartElement:
			if t.Name.Local == "claim-text" {
				// A nested claim-text: recurse so its own descendants flatten too,
				// keep the structural node, and inline its text in place.
				var nested ClaimText
				if err := nested.UnmarshalXML(d, t); err != nil {
					return err
				}
				ct.NestedClaims = append(ct.NestedClaims, nested)
				buf.WriteString(nested.Text)
			} else {
				// Any other inline element (claim-ref, figref, b, i, sub, sup, ...):
				// keep its inner text in place so the sentence stays intact.
				inner, err := flattenElementText(d)
				if err != nil {
					return err
				}
				buf.WriteString(inner)
			}
		case xml.EndElement:
			ct.Text = buf.String()
			return nil
		}
	}
}

// flattenElementText consumes tokens until the currently open element closes and
// returns the concatenated character data of all its descendants. The opening token
// has already been read by the caller.
func flattenElementText(d *xml.Decoder) (string, error) {
	var buf strings.Builder
	depth := 1
	for {
		tok, err := d.Token()
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.CharData:
			buf.Write(t)
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
			if depth == 0 {
				return buf.String(), nil
			}
		}
	}
}

// Text represents simple text with language attribute
type Text struct {
	ID   string `xml:"id,attr"`
	Lang string `xml:"lang,attr"`
	Text string `xml:",chardata"`
}

// DocumentType identifies the type of XML document
type DocumentType int

// Document type values.
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

// ExtractClaimText returns the full text of a claim with internal runs of whitespace
// collapsed to single spaces. Inline references (<claim-ref>) and nested claim-text are
// already flattened into each top-level ClaimText.Text during parsing, so no recursion
// is needed here.
func (c *Claim) ExtractClaimText() string {
	if c == nil {
		return ""
	}

	parts := make([]string, 0, len(c.ClaimText))
	for _, ct := range c.ClaimText {
		if s := normalizeSpace(ct.Text); s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, " ")
}

// normalizeSpace collapses all runs of whitespace (including the newlines and
// indentation between nested claim-text elements) into single spaces and trims.
func normalizeSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading XML: %w", err)
	}
	defer drainClose(resp.Body)

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

// GetXMLURLForApplication retrieves the XML URL and document type for a patent.
// Checks grant metadata first, then pre-grant publication (pgpub) metadata.
func (c *Client) GetXMLURLForApplication(ctx context.Context, patentNumber string) (string, DocumentType, error) {
	resp, err := c.GetPatent(ctx, patentNumber)
	if err != nil {
		return "", DocumentTypeUnknown, fmt.Errorf("failed to get patent data: %w", err)
	}

	if resp.PatentFileWrapperDataBag == nil || len(*resp.PatentFileWrapperDataBag) == 0 {
		return "", DocumentTypeUnknown, fmt.Errorf("no patent data found")
	}

	patentData := (*resp.PatentFileWrapperDataBag)[0]

	// Try grant XML first (granted patents)
	if patentData.GrantDocumentMetaData != nil && patentData.GrantDocumentMetaData.FileLocationURI != nil {
		if uri := *patentData.GrantDocumentMetaData.FileLocationURI; uri != "" {
			return uri, DocumentTypeGrant, nil
		}
	}

	// Try pre-grant publication XML (published applications)
	if patentData.PgpubDocumentMetaData != nil && patentData.PgpubDocumentMetaData.FileLocationURI != nil {
		if uri := *patentData.PgpubDocumentMetaData.FileLocationURI; uri != "" {
			return uri, DocumentTypeApplication, nil
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
