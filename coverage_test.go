package odp

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/patent-dev/uspto-odp/generated"
	oa "github.com/patent-dev/uspto-odp/generated/oa"
	tsdrgen "github.com/patent-dev/uspto-odp/generated/tsdr"
)

// unwrappedPatterns matches generated method names we deliberately do not
// expose. Each entry's `match` is interpreted as a suffix on the generated
// method name. The matched portion identifies the family of endpoints
// (variant, status-code group, etc.) that the wrappers in this package
// route around. Keep entries narrow so the test catches genuinely new
// endpoints rather than absorbing them into broad allow-lists.
var unwrappedPatterns = []struct {
	match  string
	reason string
}{
	// Body+response combo wrappers -- we always pass typed JSON (the non-WithBody form).
	{"WithBodyWithResponse", "we use the typed JSON variant instead"},
	// GET form of search endpoints (POST variants already wrapped).
	{"SearchWithResponse", "we use the POST search variant"},
	{"SearchDownloadWithResponse", "we use the POST search-download variant"},
	// OA DSAPI raw form helpers -- we use *WithFormdataBody, not the WithFormdataBodyWithResponse one.
	{"WithFormdataBodyWithResponse", "DSAPI uses formdata body via raw helper"},
	// OA *List fields *WithResponse -- we use the bare ListFields() helper that returns *http.Response.
	{"ListFieldsWithResponse", "we use the bare ListFields() helper"},
	// TSDR PDF/HTML/XML/ZIP rendering endpoints we don't surface today.
	{"GetBundleInfoPdfWithResponse", "not exposed by tsdr.go"},
	{"GetBundleInfoXmlWithResponse", "not exposed by tsdr.go"},
	{"GetBundleInfoZipWithResponse", "not exposed by tsdr.go"},
	{"GetCaseBundleContentXmlWithResponse", "not exposed by tsdr.go"},
	{"GetCaseDocsPdfWithResponse", "not exposed by tsdr.go"},
	{"GetCaseDocsPdfDownloadWithResponse", "not exposed by tsdr.go"},
	{"GetCaseDocsZipWithResponse", "not exposed by tsdr.go"},
	{"GetCaseDocsZipDownloadWithResponse", "not exposed by tsdr.go"},
	{"GetCaseStatusInfoContentHtmlWithResponse", "not exposed by tsdr.go"},
	{"GetCaseStatusInfoContentPdfWithResponse", "not exposed by tsdr.go"},
	{"GetCaseStatusInfoContentZipWithResponse", "not exposed by tsdr.go"},
	{"GetCaseStatusInfoPdfDownloadWithResponse", "not exposed by tsdr.go"},
	{"GetCaseStatusInfoZipDownloadWithResponse", "not exposed by tsdr.go"},
	{"GetDocumentContentPdfWithResponse", "not exposed by tsdr.go"},
	{"GetDocumentContentPdfDownloadWithResponse", "not exposed by tsdr.go"},
	{"GetDocumentContentZipWithResponse", "not exposed by tsdr.go"},
	{"GetDocumentContentZipDownloadWithResponse", "not exposed by tsdr.go"},
	{"GetPageContentMediaWithResponse", "not exposed by tsdr.go"},
	{"GetcaseUpdateInfoXmlWithResponse", "we use the JSON variant via GetTrademarkLastUpdate"},
	{"GetCaseDocsInfoXmlWithResponse", "API returns 406 for JSON; we use raw GetCaseDocsInfoXml"},
	{"GetDocumentInfoXmlWithResponse", "API returns 406 for JSON; we use raw GetDocumentInfoXml"},
	{"LoadOldXMLWithResponse", "legacy fallback; not exposed"},
	// Bulk file streaming wrapper -- we use DownloadBulkFile* by URI instead of the byPath helper.
	{"GetApiV1DatasetsProductsFilesProductIdentifierFileNameWithResponse", "we expose DownloadBulkFile* by URI"},
	// Status code POST: we expose the GET variant via GetStatusCodes only.
	{"PostApiV1PatentStatusCodesWithResponse", "GET StatusCodes wrapper covers this"},
}

// extraUsedRawMethods names raw (no WithResponse) generated helpers we call
// directly. Anything not in this set, not WithResponse-suffixed, and not in
// unwrappedPatterns is ignored because it isn't a high-level endpoint
// anyway. WithResponse-suffixed methods used in the wrappers are picked up
// by the AST scan and should NOT need manual entries here.
var extraUsedRawMethods = map[string]bool{
	// OA: we call the formdata variants and the bare ListFields directly.
	"OaActionsSearchWithFormdataBody":         true,
	"OaCitationsSearchWithFormdataBody":       true,
	"OaRejectionsSearchWithFormdataBody":      true,
	"EnrichedCitationsSearchWithFormdataBody": true,
	"OaActionsListFields":                     true,
	"OaCitationsListFields":                   true,
	"OaRejectionsListFields":                  true,
	"EnrichedCitationsListFields":             true,
	// TSDR: methods returning raw *http.Response because content-negotiation
	// is broken upstream.
	"GetCaseDocsInfoXml":     true,
	"GetDocumentInfoXml":     true,
	"GetDocumentContentPdf":  true,
	"LoadXML":                true, // raw form used by GetTrademarkStatusJSON via tsdrJSONEditor
}

func generatedMethodNames(t *testing.T) []string {
	t.Helper()
	seen := map[string]bool{}
	collect := func(iface any) {
		typ := reflect.TypeOf(iface)
		for i := 0; i < typ.NumMethod(); i++ {
			m := typ.Method(i)
			if !m.IsExported() {
				continue
			}
			seen[m.Name] = true
		}
	}
	collect((*generated.ClientWithResponses)(nil))
	collect((*oa.ClientWithResponses)(nil))
	collect((*tsdrgen.ClientWithResponses)(nil))
	out := make([]string, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// wrappedMethodNames scans Go source for *call expressions* whose function
// is a selector matching one of the candidate names. We require a CallExpr
// (not a bare selector reference) so comments, type references, and other
// non-call uses don't accidentally satisfy coverage.
func wrappedMethodNames(t *testing.T, files []string, candidates map[string]bool) map[string]bool {
	t.Helper()
	used := map[string]bool{}
	fset := token.NewFileSet()
	for _, f := range files {
		path, _ := filepath.Abs(f)
		af, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}
		ast.Inspect(af, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if candidates[sel.Sel.Name] {
				used[sel.Sel.Name] = true
			}
			return true
		})
	}
	return used
}

func isUnwrappedByPattern(name string) bool {
	for _, p := range unwrappedPatterns {
		if strings.HasSuffix(name, p.match) {
			return true
		}
	}
	return false
}

func TestGeneratedClientCoverage(t *testing.T) {
	allMethods := generatedMethodNames(t)
	candidate := map[string]bool{}
	for _, n := range allMethods {
		candidate[n] = true
	}

	used := wrappedMethodNames(t, []string{
		"client.go",
		"office_action.go",
		"tsdr.go",
		"xml.go",
	}, candidate)

	missing := []string{}
	for _, n := range allMethods {
		if used[n] {
			continue
		}
		if extraUsedRawMethods[n] {
			continue
		}
		// Only consider response-returning endpoint methods.
		if !strings.HasSuffix(n, "WithResponse") {
			continue
		}
		if isUnwrappedByPattern(n) {
			continue
		}
		missing = append(missing, n)
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("Generated client methods are not wrapped or pattern-allow-listed:\n  %s",
			strings.Join(missing, "\n  "))
	}
}
