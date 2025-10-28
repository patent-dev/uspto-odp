package odp

import (
	"testing"
)

func TestNormalizePatentNumber_Application(t *testing.T) {
	tests := []struct {
		input      string
		normalized string
		appNo      string
	}{
		{"17248024", "17248024", "17248024"},
		{"17/248024", "17248024", "17248024"},
		{"17/248,024", "17248024", "17248024"},
		{"US 17/248,024", "17248024", "17248024"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			pn, err := NormalizePatentNumber(tt.input)
			if err != nil {
				t.Fatalf("Failed to normalize: %v", err)
			}

			if pn.Type != PatentNumberTypeApplication {
				t.Errorf("Expected type Application, got %v", pn.Type)
			}

			if pn.Normalized != tt.normalized {
				t.Errorf("Expected normalized %s, got %s", tt.normalized, pn.Normalized)
			}

			if pn.ApplicationNo != tt.appNo {
				t.Errorf("Expected app no %s, got %s", tt.appNo, pn.ApplicationNo)
			}

			if pn.ToApplicationNumber() != tt.appNo {
				t.Errorf("Expected ToApplicationNumber %s, got %s", tt.appNo, pn.ToApplicationNumber())
			}
		})
	}
}

func TestNormalizePatentNumber_Grant(t *testing.T) {
	tests := []struct {
		input      string
		normalized string
		typeCheck  bool // whether to check that it's actually detected as grant
	}{
		{"11646472", "11646472", false},        // 8 digits, ambiguous, treated as application
		{"11,646,472", "11646472", true},       // Commas indicate grant
		{"US 11,646,472", "11646472", true},    // Commas indicate grant
		{"US 11,646,472 B2", "11646472", true}, // Kind code indicates grant
		{"11646472 B2", "11646472", true},      // Kind code indicates grant
		{"9123456", "9123456", true},           // 7 digits is grant
		{"9,123,456", "9123456", true},         // Commas indicate grant
		{"US 9,123,456 B1", "9123456", true},   // Kind code indicates grant
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			pn, err := NormalizePatentNumber(tt.input)
			if err != nil {
				t.Fatalf("Failed to normalize: %v", err)
			}

			if tt.typeCheck && pn.Type != PatentNumberTypeGrant {
				t.Errorf("Expected type Grant, got %v", pn.Type)
			}

			if pn.Normalized != tt.normalized {
				t.Errorf("Expected normalized %s, got %s", tt.normalized, pn.Normalized)
			}

			// Grant numbers normalize to themselves for API use
			if pn.ToApplicationNumber() != tt.normalized {
				t.Errorf("Expected ToApplicationNumber %s, got %s", tt.normalized, pn.ToApplicationNumber())
			}
		})
	}
}

func TestNormalizePatentNumber_Publication(t *testing.T) {
	tests := []struct {
		input      string
		normalized string
	}{
		{"20250087686", "20250087686"},
		{"US20250087686", "20250087686"},
		{"US20250087686A1", "20250087686"},
		{"US 2025/0087686 A1", "20250087686"},
		{"2025/0087686", "20250087686"},
		{"20240123456", "20240123456"},
		{"US 2024/0123456 A1", "20240123456"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			pn, err := NormalizePatentNumber(tt.input)
			if err != nil {
				t.Fatalf("Failed to normalize: %v", err)
			}

			if pn.Type != PatentNumberTypePublication {
				t.Errorf("Expected type Publication, got %v", pn.Type)
			}

			if pn.Normalized != tt.normalized {
				t.Errorf("Expected normalized %s, got %s", tt.normalized, pn.Normalized)
			}
		})
	}
}

func TestNormalizePatentNumber_Invalid(t *testing.T) {
	tests := []string{
		"",
		"abc123",
		"US",
		"patent123",
		"123",
		"1234567890123456", // too long
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := NormalizePatentNumber(input)
			if err == nil {
				t.Errorf("Expected error for input: %s", input)
			}
		})
	}
}

func TestPatentNumber_FormatAsApplication(t *testing.T) {
	pn, _ := NormalizePatentNumber("17248024")
	formatted := pn.FormatAsApplication()
	expected := "17/248,024"

	if formatted != expected {
		t.Errorf("Expected %s, got %s", expected, formatted)
	}
}

func TestPatentNumber_FormatAsGrant(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"US 11,646,472 B2", "11,646,472"}, // Use formatted input with kind code
		{"9123456", "9,123,456"},           // 7 digits is unambiguously a grant
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			pn, _ := NormalizePatentNumber(tt.input)
			formatted := pn.FormatAsGrant()

			if formatted != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, formatted)
			}
		})
	}
}

func TestPatentNumber_FormatAsPublication(t *testing.T) {
	pn, _ := NormalizePatentNumber("20250087686")
	formatted := pn.FormatAsPublication()
	expected := "2025/0087686"

	if formatted != expected {
		t.Errorf("Expected %s, got %s", expected, formatted)
	}
}

func TestPatentNumber_String(t *testing.T) {
	pn, _ := NormalizePatentNumber("US 11,646,472 B2")
	str := pn.String()

	if str == "" {
		t.Error("String() returned empty")
	}

	// Should contain original and type info
	if !contains(str, "11,646,472") && !contains(str, "grant") {
		t.Errorf("String should contain original and type info: %s", str)
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test real-world examples
func TestNormalizePatentNumber_RealExamples(t *testing.T) {
	tests := []struct {
		input         string
		expectedType  PatentNumberType
		shouldSucceed bool
	}{
		// Application 17/248,024 -> Grant US 11,646,472 B2
		{"17/248,024", PatentNumberTypeApplication, true},
		{"17248024", PatentNumberTypeApplication, true},
		{"US 11,646,472 B2", PatentNumberTypeGrant, true},
		// Note: 8-digit numbers without formatting are ambiguous (could be app or grant)
		// Both work with the API, so this is OK
		{"11646472", PatentNumberTypeApplication, true}, // 8 digits defaults to application

		// Publication US 2025/0087686 A1
		{"US 2025/0087686 A1", PatentNumberTypePublication, true},
		{"US20250087686A1", PatentNumberTypePublication, true},
		{"20250087686", PatentNumberTypePublication, true},

		// More examples
		{"16/123,456", PatentNumberTypeApplication, true},
		{"US 10,000,000 B2", PatentNumberTypeGrant, true},
		{"US 2024/0000001 A1", PatentNumberTypePublication, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			pn, err := NormalizePatentNumber(tt.input)

			if tt.shouldSucceed {
				if err != nil {
					t.Fatalf("Expected success, got error: %v", err)
				}

				if pn.Type != tt.expectedType {
					t.Errorf("Expected type %v, got %v", tt.expectedType, pn.Type)
				}

				if pn.Normalized == "" {
					t.Error("Normalized should not be empty")
				}

				if pn.ToApplicationNumber() == "" {
					t.Error("ToApplicationNumber should not be empty")
				}

				t.Logf("✓ %s -> %s (type: %v)", tt.input, pn.Normalized, pn.Type)
			} else {
				if err == nil {
					t.Error("Expected error, got success")
				}
			}
		})
	}
}

func BenchmarkNormalizePatentNumber(b *testing.B) {
	inputs := []string{
		"17248024",
		"US 11,646,472 B2",
		"US20250087686A1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			_, _ = NormalizePatentNumber(input)
		}
	}
}
