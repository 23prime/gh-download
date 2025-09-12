package config

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureOutput captures stdout during function execution
func captureOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestConfig_DefaultValues(t *testing.T) {
	var config Config

	// Test zero values
	if config.Repository != "" {
		t.Errorf("Expected Repository to be empty, got %q", config.Repository)
	}
	if config.Tag != "" {
		t.Errorf("Expected Tag to be empty, got %q", config.Tag)
	}
	if config.Pattern != "" {
		t.Errorf("Expected Pattern to be empty, got %q", config.Pattern)
	}
	if config.Directory != "" {
		t.Errorf("Expected Directory to be empty, got %q", config.Directory)
	}
	if config.Archive != "" {
		t.Errorf("Expected Archive to be empty, got %q", config.Archive)
	}
	if config.List != false {
		t.Errorf("Expected List to be false, got %t", config.List)
	}
	if config.Releases != false {
		t.Errorf("Expected Releases to be false, got %t", config.Releases)
	}
	if config.Help != false {
		t.Errorf("Expected Help to be false, got %t", config.Help)
	}
}

func TestPrintUsage(t *testing.T) {
	output := captureOutput(func() {
		PrintUsage()
	})

	// Test that output contains expected sections
	expectedSections := []string{
		"gh-download - Download files from GitHub releases",
		"Usage:",
		"gh download [repository] [tag] [flags]",
		"Arguments:",
		"repository",
		"tag",
		"Flags:",
		"-R, --repo string",
		"-t, --tag string",
		"-p, --pattern string",
		"-d, --dir string",
		"--archive string",
		"-l, --list",
		"-r, --releases",
		"-h, --help",
		"Examples:",
		"gh download owner/repo",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Expected output to contain %q, but it was missing", section)
		}
	}
}

func TestPrintUsage_OutputFormat(t *testing.T) {
	output := captureOutput(func() {
		PrintUsage()
	})

	// Test output structure
	lines := strings.Split(output, "\n")

	// Should have multiple lines
	if len(lines) < 10 {
		t.Errorf("Expected output to have at least 10 lines, got %d", len(lines))
	}

	// First line should be the title
	if !strings.Contains(lines[0], "gh-download") {
		t.Errorf("Expected first line to contain 'gh-download', got %q", lines[0])
	}

	// Should contain Usage section
	found := false
	for _, line := range lines {
		if strings.Contains(line, "Usage:") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected output to contain 'Usage:' section")
	}
}

func TestConfig_StructFields(t *testing.T) {
	config := Config{
		Repository: "owner/repo",
		Tag:        "v1.0.0",
		Pattern:    "*.tar.gz",
		Directory:  "./downloads",
		Archive:    "zip",
		List:       true,
		Releases:   false,
		Help:       true,
	}

	// Test that all fields are set correctly
	if config.Repository != "owner/repo" {
		t.Errorf("Expected Repository to be 'owner/repo', got %q", config.Repository)
	}
	if config.Tag != "v1.0.0" {
		t.Errorf("Expected Tag to be 'v1.0.0', got %q", config.Tag)
	}
	if config.Pattern != "*.tar.gz" {
		t.Errorf("Expected Pattern to be '*.tar.gz', got %q", config.Pattern)
	}
	if config.Directory != "./downloads" {
		t.Errorf("Expected Directory to be './downloads', got %q", config.Directory)
	}
	if config.Archive != "zip" {
		t.Errorf("Expected Archive to be 'zip', got %q", config.Archive)
	}
	if config.List != true {
		t.Errorf("Expected List to be true, got %t", config.List)
	}
	if config.Releases != false {
		t.Errorf("Expected Releases to be false, got %t", config.Releases)
	}
	if config.Help != true {
		t.Errorf("Expected Help to be true, got %t", config.Help)
	}
}

// NOTE: ParseArgs() testing is complex due to flag package's global state.
// In a real-world scenario, we might refactor ParseArgs to accept arguments
// or use dependency injection for better testability.
func TestPrintUsage_ContainsKeyElements(t *testing.T) {
	output := captureOutput(func() {
		PrintUsage()
	})

	// Test specific key elements that must be present
	keyElements := map[string]string{
		"title":           "gh-download",
		"usage_header":    "Usage:",
		"repo_flag":       "--repo",
		"pattern_flag":    "--pattern",
		"list_flag":       "--list",
		"releases_flag":   "--releases",
		"help_flag":       "--help",
		"example_basic":   "gh download owner/repo",
		"example_pattern": "*.tar.gz",
	}

	for element, expected := range keyElements {
		if !strings.Contains(output, expected) {
			t.Errorf("Missing key element '%s': expected to find '%s' in output", element, expected)
		}
	}
}
