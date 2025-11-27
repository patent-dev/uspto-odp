package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileFormat represents the format of a file
type FileFormat string

const (
	FormatJSON FileFormat = "json"
	FormatCSV  FileFormat = "csv"
	FormatXML  FileFormat = "xml"
	FormatText FileFormat = "txt"
)

// ExampleSaver saves request/response pairs to disk
type ExampleSaver struct {
	baseDir string
}

// NewExampleSaver creates a new ExampleSaver with the specified base directory
func NewExampleSaver(baseDir string) *ExampleSaver {
	return &ExampleSaver{baseDir: baseDir}
}

// SaveExample saves a request description and response data to the examples directory
func (s *ExampleSaver) SaveExample(endpointName string, requestDesc string, response []byte, format FileFormat) error {
	// Create endpoint directory
	dir := filepath.Join(s.baseDir, endpointName)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Save request description
	requestFile := filepath.Join(dir, "request.txt")
	if err := os.WriteFile(requestFile, []byte(requestDesc), 0600); err != nil {
		return fmt.Errorf("failed to save request: %w", err)
	}

	// Determine response filename based on format
	responseFile := filepath.Join(dir, fmt.Sprintf("response.%s", format))

	// Save response
	if err := os.WriteFile(responseFile, response, 0600); err != nil {
		return fmt.Errorf("failed to save response: %w", err)
	}

	return nil
}

// SaveJSONExample saves a JSON response with pretty formatting
func (s *ExampleSaver) SaveJSONExample(endpointName string, requestDesc string, response interface{}) error {
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	return s.SaveExample(endpointName, requestDesc, data, FormatJSON)
}

// DetectFormat detects the format of data
func DetectFormat(data []byte) FileFormat {
	if len(data) == 0 {
		return FormatText
	}

	// Check for XML
	trimmed := strings.TrimSpace(string(data))
	if strings.HasPrefix(trimmed, "<?xml") || strings.HasPrefix(trimmed, "<") {
		return FormatXML
	}

	// Check for JSON
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		if json.Valid(data) {
			return FormatJSON
		}
	}

	// Check for CSV (has commas and newlines)
	if strings.Contains(trimmed, ",") && strings.Contains(trimmed, "\n") {
		return FormatCSV
	}

	return FormatText
}

// FormatRequestDescription formats a request description with parameters
func FormatRequestDescription(method string, params map[string]string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Method: %s\n\n", method))
	if len(params) > 0 {
		sb.WriteString("Parameters:\n")
		for k, v := range params {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}
	return sb.String()
}
