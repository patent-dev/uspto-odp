package odp

import (
	"os"
	"strings"
	"testing"
)

// These tests use real patent grant XML pulled verbatim from the USPTO ODP API
// (see testdata/). They guard the claim-ref handling: dependent claims reference
// their parent via a <claim-ref> child element inside <claim-text>, e.g.
//
//	<claim-text>2. The system according to <claim-ref idref="CLM-00001">claim 1</claim-ref>, wherein ...</claim-text>
//
// A naive xml:",chardata" mapping concatenates only the text around the child and
// drops "claim 1", yielding "2. The system according to , wherein ...". These
// tests assert the reference text is preserved in document order.

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return data
}

func TestExtractClaims_PreservesClaimRef_Raytheon(t *testing.T) {
	doc, err := ParseGrantXML(readFixture(t, "grant_us10000000b2_14643719.xml"))
	if err != nil {
		t.Fatalf("parse grant XML: %v", err)
	}
	claims := doc.GetClaims()
	if claims == nil {
		t.Fatal("GetClaims returned nil")
	}
	texts := claims.ExtractAllClaimsText()
	if len(texts) != 20 {
		t.Fatalf("want 20 claims, got %d", len(texts))
	}

	// Dependent claim 2 references claim 1 via <claim-ref>.
	want2 := "2. The system according to claim 1, wherein the two-dimensional array of detector elements comprises a large format array."
	if texts[1] != want2 {
		t.Errorf("claim 2 mismatch:\n got: %q\nwant: %q", texts[1], want2)
	}

	// Independent claim 1 has deeply nested <claim-text>; all parts must survive.
	for _, frag := range []string{
		"1. A laser detection and ranging (LADAR) system, comprising:",
		"a two-dimensional array of detector elements",
		"local processing circuitry coupled to an output",
		"a data bus coupled to one or more outputs",
		"a processor coupled to the data bus",
	} {
		if !strings.Contains(texts[0], frag) {
			t.Errorf("claim 1 missing nested fragment %q\ngot: %q", frag, texts[0])
		}
	}

	// Regression signature of the bug: a dropped claim-ref leaves "according to ,".
	for i, txt := range texts {
		if strings.Contains(txt, "according to ,") {
			t.Errorf("claim %d still shows dropped claim-ref: %q", i+1, txt)
		}
	}
}

func TestExtractClaims_PreservesClaimRef_PolyPlus(t *testing.T) {
	doc, err := ParseGrantXML(readFixture(t, "grant_us11646472b2_17248024.xml"))
	if err != nil {
		t.Fatalf("parse grant XML: %v", err)
	}
	claims := doc.GetClaims()
	if claims == nil {
		t.Fatal("GetClaims returned nil")
	}
	texts := claims.ExtractAllClaimsText()
	if len(texts) != 17 {
		t.Fatalf("want 17 claims, got %d", len(texts))
	}

	want2 := "2. The method of claim 1 wherein the lithium anode is lithium metal."
	if texts[1] != want2 {
		t.Errorf("claim 2 mismatch:\n got: %q\nwant: %q", texts[1], want2)
	}

	// The bug surfaced here as "The method of  wherein" (double space, no ref).
	for i, txt := range texts {
		if strings.Contains(txt, "method of  wherein") || strings.Contains(txt, "method of  comprising") {
			t.Errorf("claim %d still shows dropped claim-ref: %q", i+1, txt)
		}
	}
}
