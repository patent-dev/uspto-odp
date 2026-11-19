package odp

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/patent-dev/uspto-odp/generated/tmd"
)

func TestDownloadTrademarkDecisionFileRejectsForeignURI(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = "test"
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// fileDownloadURI of the recorded Administrative record.
	raw, err := os.ReadFile(filepath.Join("testdata", "strictdecode", "get_trademark_decision.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rec tmd.TrademarkDecisionResponse
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatal(err)
	}
	live := *rec.TrademarkDecisionsDataBag.FileDownloadURI
	if !strings.HasPrefix(live, c.config.BaseURL+trademarkDecisionFilePath) {
		t.Fatalf("recorded fileDownloadURI %q does not match the accepted prefix", live)
	}

	for _, uri := range []string{
		"",
		"https://evil.example" + trademarkDecisionFilePath + "x.pdf",
		c.config.BaseURL + "/api/v1/download/applications/18957107/M8N07YQ8WFYGX53.pdf",
		c.config.BaseURL + "/api/v1/trademark/decision-proceeding/search",
	} {
		var buf bytes.Buffer
		if err := c.DownloadTrademarkDecisionFile(context.Background(), uri, &buf); err == nil || !strings.Contains(err.Error(), "invalid trademark decision fileDownloadURI") {
			t.Errorf("DownloadTrademarkDecisionFile(%q) err=%v, want URI rejection", uri, err)
		}
	}
}
