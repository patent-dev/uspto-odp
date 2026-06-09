//go:build integration
// +build integration

package odp

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/patent-dev/uspto-odp/generated"
)

// TestDownloadPatentDocument_Integration exercises the real list -> download flow:
// it discovers a document download URL from the file wrapper of application
// 18957107 and fetches it, rather than pinning a volatile document identifier.
func TestDownloadPatentDocument_Integration(t *testing.T) {
	apiKey := os.Getenv("USPTO_API_KEY")
	if apiKey == "" {
		t.Skip("USPTO_API_KEY not set")
	}
	cfg := DefaultConfig()
	cfg.APIKey = apiKey
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	docs, err := c.GetPatentDocuments(ctx, "18957107")
	if err != nil {
		t.Fatalf("GetPatentDocuments: %v", err)
	}
	downloadURL := firstDocumentDownloadURL(docs)
	if downloadURL == "" {
		t.Skip("no downloadable documents for the test application")
	}

	var buf bytes.Buffer
	if err := c.DownloadPatentDocument(ctx, downloadURL, &buf); err != nil {
		t.Fatalf("DownloadPatentDocument(%s): %v", downloadURL, err)
	}
	if buf.Len() < 1000 || !bytes.HasPrefix(buf.Bytes(), []byte("%PDF")) {
		t.Fatalf("expected a PDF >=1000 bytes, got %d bytes", buf.Len())
	}
	t.Logf("downloaded %d-byte file-wrapper document from %s", buf.Len(), downloadURL)

	// An off-host URL must be rejected before any authenticated request goes out.
	if err := c.DownloadPatentDocument(ctx, "https://evil.example/x.pdf", new(bytes.Buffer)); err == nil {
		t.Error("expected off-host URL to be rejected")
	}
}

func firstDocumentDownloadURL(docs *generated.DocumentBag) string {
	if docs == nil || docs.DocumentBag == nil {
		return ""
	}
	for _, d := range *docs.DocumentBag {
		if d.DownloadOptionBag == nil {
			continue
		}
		for _, o := range *d.DownloadOptionBag {
			if o.DownloadUrl != nil && *o.DownloadUrl != "" {
				return *o.DownloadUrl
			}
		}
	}
	return ""
}
