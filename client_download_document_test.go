package odp

import "testing"

func TestValidateDocumentDownloadURL(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = "test"
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	base := c.config.BaseURL

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid document pdf", base + "/api/v1/download/applications/18957107/M8N07YQ8WFYGX53.pdf", false},
		{"empty", "", true},
		{"off-host", "https://evil.example/x.pdf", true},
		{"bulk datasets path", base + "/api/v1/datasets/products/files/x.zip", true},
		{"metadata path, not download", base + "/api/v1/patent/applications/18957107/documents", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := c.validateDocumentDownloadURL(tt.url); (err != nil) != tt.wantErr {
				t.Errorf("validateDocumentDownloadURL(%q) err=%v, wantErr=%v", tt.url, err, tt.wantErr)
			}
		})
	}
}
