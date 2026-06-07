package odp

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// US grant numbers reached 8 digits at 10,000,000 (June 2018) and, as of 2026, run to
// roughly 12.7 million. An 8-digit value in this range is ambiguous because it can be
// either such a grant or an application in series 10-12. An 8-digit value at or beyond
// application series 13 (>= 13,000,000) cannot currently be a grant, so it is treated as
// an unambiguous application. Raise the upper bound as grant numbers grow.
const (
	firstEightDigitGrant = 10000000
	maxEightDigitGrant   = 12999999
)

// PatentNumberType indicates the type of patent number
type PatentNumberType int

// Patent number type values.
const (
	PatentNumberTypeUnknown PatentNumberType = iota
	PatentNumberTypeApplication
	PatentNumberTypeGrant
	PatentNumberTypePublication
	PatentNumberTypePCT
)

// PatentNumber represents a normalized patent number
type PatentNumber struct {
	Original      string           // Original input
	Normalized    string           // Normalized format (digits only, or PCT API form)
	ApplicationNo string           // Application number if derivable
	Type          PatentNumberType // Type of number
	Country       string           // Country code (usually "US")
	// KindCode is the suffix when supplied by the caller (e.g., "A1", "A2",
	// "B2"). For publication numbers it is threaded through resolution to
	// preserve republished kinds. For grant numbers it is captured for
	// display only -- USPTO's search API ignores grant kind codes when
	// resolving to an application number.
	KindCode string
	// Ambiguous is true when the input was a bare number that could be either a
	// grant or an application (an 8-digit value with no kind code, slash, or
	// comma to disambiguate). Resolution probes both interpretations rather than
	// silently assuming one. See Client.ResolvePatentNumber.
	Ambiguous bool
}

// Patent number patterns
var (
	// Application with slash: 17/248,024, 17/248024, US 17/248,024
	applicationWithSlashPattern = regexp.MustCompile(`^(?:US)?[\s]*(\d{2})/(\d{3})[,\s]*(\d{3})$`)

	// Grant with kind code, separated or compact: US 11,646,472 B2, 9,123,456 B1,
	// US11646472B2. The whitespace before the kind code is optional so the compact
	// form (no separators) parses the same as the formatted one.
	grantWithKindPattern = regexp.MustCompile(`^(?:US)?[\s]*(\d{1,2})[,\s]*(\d{3})[,\s]*(\d{3})[\s]*([A-Z]\d)$`)

	// Grant with comma formatting: 11,646,472, US 11,646,472
	grantWithCommaPattern = regexp.MustCompile(`^(?:US)?[\s]*(\d{1,2}),(\d{3}),(\d{3})$`)

	// Publication: 20250087686, US20250087686A1, US 2025/0087686 A1
	publicationPattern = regexp.MustCompile(`^(?:US)?[\s]*(\d{4})[/,\s]*(\d{7})(?:\s*([A-Z]\d))?$`)

	// PCT 15-char API form: PCTUS2025058371 (no slashes anywhere)
	pct15Pattern = regexp.MustCompile(`^(?i)PCTUS(\d{4})(\d{6})$`)

	// PCT 17-char display form: PCT/US2025/058371 (slashes after PCT and after year)
	pct17Pattern = regexp.MustCompile(`^(?i)PCT/US(\d{4})/(\d{6})$`)

	// PCT legacy 12-char form: PCTUS0719317 (preserve as-is, API accepts it)
	pct12Pattern = regexp.MustCompile(`^(?i)PCTUS(\d{7})$`)

	// Simple patterns for fallback
	digitsOnlyPattern = regexp.MustCompile(`^\d+$`)
)

// NormalizePatentNumber normalizes various patent number formats to application numbers
// Accepts formats like:
//   - Application: "17248024", "17/248,024", "17/248024"
//   - Grant: "11646472", "11,646,472", "US 11,646,472 B2"
//   - Publication: "20250087686", "US20250087686A1", "US 2025/0087686 A1"
//   - PCT: "PCTUS2025058371" (15-char API form), "PCT/US2025/058371" (17-char display),
//     "PCTUS0719317" (12-char legacy)
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

	// Try PCT 15-char API form (no slashes)
	if matches := pct15Pattern.FindStringSubmatch(cleaned); matches != nil {
		result.Normalized = "PCTUS" + matches[1] + matches[2]
		result.Type = PatentNumberTypePCT
		result.ApplicationNo = result.Normalized
		return result, nil
	}

	// Try PCT 17-char display form (PCT/US####/######)
	if matches := pct17Pattern.FindStringSubmatch(cleaned); matches != nil {
		result.Normalized = "PCTUS" + matches[1] + matches[2]
		result.Type = PatentNumberTypePCT
		result.ApplicationNo = result.Normalized
		return result, nil
	}

	// Try PCT 12-char legacy form
	if matches := pct12Pattern.FindStringSubmatch(cleaned); matches != nil {
		result.Normalized = "PCTUS" + matches[1]
		result.Type = PatentNumberTypePCT
		result.ApplicationNo = result.Normalized
		return result, nil
	}

	// Try grant with kind code first (most specific, e.g., US 11,646,472 B2)
	if matches := grantWithKindPattern.FindStringSubmatch(cleaned); matches != nil {
		series := matches[1]
		number := matches[2] + matches[3]
		result.Normalized = series + number
		result.Type = PatentNumberTypeGrant
		result.KindCode = matches[4]
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
		// matches[3] is the optional kind-code group; empty when absent.
		result.KindCode = matches[3]
		result.ApplicationNo = "" // Will need to be looked up
		return result, nil
	}

	// Fallback: digits only, optionally with a US prefix (e.g. "US11646472").
	// The grant/application/publication patterns above already accept an optional
	// "US"; mirror that here so a bare prefixed number resolves the same way.
	bare := cleaned
	if t := strings.TrimPrefix(strings.TrimPrefix(cleaned, "US"), "us"); digitsOnlyPattern.MatchString(t) {
		bare = t
	}
	if digitsOnlyPattern.MatchString(bare) {
		result.Normalized = bare

		// Heuristics based on length
		length := len(bare)
		switch length {
		case 7:
			// 7-digit grant number (e.g., 9123456)
			result.Type = PatentNumberTypeGrant
		case 8, 9:
			// 8-9 digit application number (e.g., 17248024). An 8-digit value in
			// the 8-digit grant range is ambiguous (grant vs application series
			// 10-12); resolution probes both. Everything else here is treated as
			// an unambiguous application, preserving the direct lookup.
			result.Type = PatentNumberTypeApplication
			result.ApplicationNo = bare
			if length == 8 {
				if n, convErr := strconv.Atoi(bare); convErr == nil &&
					n >= firstEightDigitGrant && n <= maxEightDigitGrant {
					result.Ambiguous = true
				}
			}
		case 11:
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
// For application and PCT numbers, returns as-is.
// For grant/publication numbers, returns the normalized number which can be used with the API.
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
	case PatentNumberTypePCT:
		typeStr = "pct"
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

// FormatAsPCT formats a PCT number for display (e.g., PCT/US2025/058371).
// 15-char API form -> 17-char display form.
// 12-char legacy form is returned unchanged (no canonical display form).
func (pn *PatentNumber) FormatAsPCT() string {
	if pn.Type != PatentNumberTypePCT {
		return pn.Normalized
	}
	// PCTUS + 4 year + 6 sequence = 15 chars
	if len(pn.Normalized) == 15 && strings.HasPrefix(pn.Normalized, "PCTUS") {
		year := pn.Normalized[5:9]
		seq := pn.Normalized[9:]
		return fmt.Sprintf("PCT/US%s/%s", year, seq)
	}
	return pn.Normalized
}
