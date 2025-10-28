package odp

import (
	"fmt"
	"regexp"
	"strings"
)

// PatentNumberType indicates the type of patent number
type PatentNumberType int

const (
	PatentNumberTypeUnknown PatentNumberType = iota
	PatentNumberTypeApplication
	PatentNumberTypeGrant
	PatentNumberTypePublication
)

// PatentNumber represents a normalized patent number
type PatentNumber struct {
	Original      string           // Original input
	Normalized    string           // Normalized format (digits only)
	ApplicationNo string           // Application number if derivable
	Type          PatentNumberType // Type of number
	Country       string           // Country code (usually "US")
}

// Patent number patterns
var (
	// Application with slash: 17/248,024, 17/248024, US 17/248,024
	applicationWithSlashPattern = regexp.MustCompile(`^(?:US)?[\s]*(\d{2})/(\d{3})[,\s]*(\d{3})$`)

	// Grant with kind code: US 11,646,472 B2, 9,123,456 B1
	grantWithKindPattern = regexp.MustCompile(`^(?:US)?[\s]*(\d{1,2})[,\s]*(\d{3})[,\s]*(\d{3})[\s]+[A-Z]\d$`)

	// Grant with comma formatting: 11,646,472, US 11,646,472
	grantWithCommaPattern = regexp.MustCompile(`^(?:US)?[\s]*(\d{1,2}),(\d{3}),(\d{3})$`)

	// Publication: 20250087686, US20250087686A1, US 2025/0087686 A1
	publicationPattern = regexp.MustCompile(`^(?:US)?[\s]*(\d{4})[/,\s]*(\d{7})(?:\s*[A-Z]\d)?$`)

	// Simple patterns for fallback
	digitsOnlyPattern = regexp.MustCompile(`^\d+$`)
)

// NormalizePatentNumber normalizes various patent number formats to application numbers
// Accepts formats like:
//   - Application: "17248024", "17/248,024", "17/248024"
//   - Grant: "11646472", "11,646,472", "US 11,646,472 B2"
//   - Publication: "20250087686", "US20250087686A1", "US 2025/0087686 A1"
func NormalizePatentNumber(input string) (*PatentNumber, error) {
	if input == "" {
		return nil, fmt.Errorf("patent number cannot be empty")
	}

	// Clean up input
	cleaned := strings.TrimSpace(input)

	result := &PatentNumber{
		Original: input,
		Country:  "US",
	}

	// Try grant with kind code first (most specific, e.g., US 11,646,472 B2)
	if matches := grantWithKindPattern.FindStringSubmatch(cleaned); matches != nil {
		series := matches[1]
		number := matches[2] + matches[3]
		result.Normalized = series + number
		result.Type = PatentNumberTypeGrant
		result.ApplicationNo = "" // Will need to be looked up
		return result, nil
	}

	// Try application with slash (e.g., 17/248,024)
	if matches := applicationWithSlashPattern.FindStringSubmatch(cleaned); matches != nil {
		series := matches[1]
		number := matches[2] + matches[3]
		result.Normalized = series + number
		result.ApplicationNo = series + number
		result.Type = PatentNumberTypeApplication
		return result, nil
	}

	// Try grant with comma formatting (e.g., 11,646,472)
	if matches := grantWithCommaPattern.FindStringSubmatch(cleaned); matches != nil {
		series := matches[1]
		number := matches[2] + matches[3]
		result.Normalized = series + number
		result.Type = PatentNumberTypeGrant
		result.ApplicationNo = "" // Will need to be looked up
		return result, nil
	}

	// Try publication pattern (e.g., US20250087686A1)
	if matches := publicationPattern.FindStringSubmatch(cleaned); matches != nil {
		year := matches[1]
		number := matches[2]
		result.Normalized = year + number
		result.Type = PatentNumberTypePublication
		result.ApplicationNo = "" // Will need to be looked up
		return result, nil
	}

	// Fallback: digits only
	if digitsOnlyPattern.MatchString(cleaned) {
		result.Normalized = cleaned

		// Heuristics based on length
		length := len(cleaned)
		switch {
		case length == 7:
			// 7-digit grant number (e.g., 9123456)
			result.Type = PatentNumberTypeGrant
		case length == 8 || length == 9:
			// 8-9 digit application number (e.g., 17248024)
			// Note: 8 digits is ambiguous (could be recent grant or application).
			// Without formatting clues (commas, slashes, kind codes), we default to
			// application as both work with the API and applications are more common.
			result.Type = PatentNumberTypeApplication
			result.ApplicationNo = cleaned
		case length == 11:
			// 11-digit publication number (e.g., 20250087686)
			result.Type = PatentNumberTypePublication
		default:
			// Invalid length
			return nil, fmt.Errorf("unrecognized patent number format (invalid length %d): %s", length, input)
		}

		return result, nil
	}

	return nil, fmt.Errorf("unrecognized patent number format: %s", input)
}

// ToApplicationNumber converts a patent number to application number format
// For application numbers, returns as-is
// For grant/publication numbers, returns the normalized number which can be used with the API
func (pn *PatentNumber) ToApplicationNumber() string {
	if pn.ApplicationNo != "" {
		return pn.ApplicationNo
	}
	// For grants and publications, the normalized number works with the API
	return pn.Normalized
}

// String returns a human-readable representation
func (pn *PatentNumber) String() string {
	typeStr := "unknown"
	switch pn.Type {
	case PatentNumberTypeApplication:
		typeStr = "application"
	case PatentNumberTypeGrant:
		typeStr = "grant"
	case PatentNumberTypePublication:
		typeStr = "publication"
	}

	return fmt.Sprintf("%s (%s: %s)", pn.Original, typeStr, pn.Normalized)
}

// FormatAsApplication formats number as application (e.g., 17/248,024)
func (pn *PatentNumber) FormatAsApplication() string {
	if pn.Type != PatentNumberTypeApplication {
		return pn.Normalized
	}

	if len(pn.Normalized) >= 8 {
		series := pn.Normalized[:2]
		first := pn.Normalized[2:5]
		second := pn.Normalized[5:]
		return fmt.Sprintf("%s/%s,%s", series, first, second)
	}

	return pn.Normalized
}

// FormatAsGrant formats number as grant (e.g., 11,646,472)
func (pn *PatentNumber) FormatAsGrant() string {
	if pn.Type != PatentNumberTypeGrant {
		return pn.Normalized
	}

	if len(pn.Normalized) >= 7 {
		// Handle both 7 and 8 digit grant numbers
		if len(pn.Normalized) == 7 {
			first := pn.Normalized[:1]
			second := pn.Normalized[1:4]
			third := pn.Normalized[4:]
			return fmt.Sprintf("%s,%s,%s", first, second, third)
		} else if len(pn.Normalized) == 8 {
			first := pn.Normalized[:2]
			second := pn.Normalized[2:5]
			third := pn.Normalized[5:]
			return fmt.Sprintf("%s,%s,%s", first, second, third)
		}
	}

	return pn.Normalized
}

// FormatAsPublication formats number as publication (e.g., 2025/0087686)
func (pn *PatentNumber) FormatAsPublication() string {
	if pn.Type != PatentNumberTypePublication {
		return pn.Normalized
	}

	if len(pn.Normalized) == 11 {
		year := pn.Normalized[:4]
		number := pn.Normalized[4:]
		return fmt.Sprintf("%s/%s", year, number)
	}

	return pn.Normalized
}
