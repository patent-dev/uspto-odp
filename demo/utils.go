package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

func printHeader(title string) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println(title)
	fmt.Println(strings.Repeat("=", 80))
}

func printSubHeader(title string) {
	fmt.Println()
	fmt.Println(title)
	fmt.Println(strings.Repeat("-", 80))
}

func printResult(key, value string) {
	fmt.Printf("%-25s: %s\n", key, value)
}

func printError(err error) {
	fmt.Printf("Error: %v\n", err)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func formatJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
