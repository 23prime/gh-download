package tests

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Helper function to run the main program with arguments
func runGhDownload(t *testing.T, args ...string) (string, string, int) {
	t.Helper()

	// Build command to run main.go from the parent directory
	cmdArgs := append([]string{"run", "../main.go"}, args...)
	cmd := exec.Command("go", cmdArgs...)

	// Capture stdout and stderr
	stdout, err := cmd.Output()
	stderr := ""
	exitCode := 0

	if exitError, ok := err.(*exec.ExitError); ok {
		stderr = string(exitError.Stderr)
		exitCode = exitError.ExitCode()
	} else if err != nil {
		t.Logf("Command execution error: %v", err)
		stderr = err.Error()
		exitCode = 1
	}

	return string(stdout), stderr, exitCode
}

// Helper function to create temporary directory for downloads
func createTempDir(t *testing.T) string {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "gh-download-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})
	return tempDir
}

func TestIntegration_Help(t *testing.T) {
	stdout, stderr, exitCode := runGhDownload(t, "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	expectedStrings := []string{
		"gh-download - Download files from GitHub releases",
		"Usage:",
		"--repo",
		"--help",
		"Examples:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected help output to contain %q, but it was missing", expected)
		}
	}
}

func TestIntegration_ListAssets(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	stdout, stderr, exitCode := runGhDownload(t, "--repo", "cli/cli", "--list")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	expectedStrings := []string{
		"Release:",
		"cli/cli",
		"Assets matching pattern '*':",
		"Total:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected list output to contain %q, but it was missing", expected)
		}
	}
}

func TestIntegration_ListReleases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	stdout, stderr, exitCode := runGhDownload(t, "--repo", "cli/cli", "--releases")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	expectedStrings := []string{
		"Releases for cli/cli:",
		"Assets:",
		"Total:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected releases output to contain %q, but it was missing", expected)
		}
	}
}

func TestIntegration_ArchiveDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := createTempDir(t)

	stdout, stderr, exitCode := runGhDownload(t,
		"--repo", "cli/cli",
		"--archive", "zip",
		"--dir", tempDir)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "Downloaded archive:") {
		t.Errorf("Expected download confirmation, but it was missing")
	}

	// Check if zip file was actually downloaded
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	found := false
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".zip") && strings.Contains(file.Name(), "cli-cli") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected zip file to be downloaded, but none found in %s", tempDir)
	}
}

func TestIntegration_AssetDownload_WithPattern(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := createTempDir(t)

	// Try to download tar.gz assets (common in cli/cli releases)
	stdout, stderr, exitCode := runGhDownload(t,
		"--repo", "cli/cli",
		"--pattern", "*.tar.gz",
		"--dir", tempDir)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	expectedStrings := []string{
		"Release:",
		"Found",
		"matching assets to download",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(stdout, expected) {
			t.Errorf("Expected download output to contain %q, but it was missing", expected)
		}
	}
}

func TestIntegration_ErrorHandling_InvalidRepo(t *testing.T) {
	stdout, stderr, exitCode := runGhDownload(t, "--repo", "nonexistent/nonexistent")

	if exitCode == 0 {
		t.Errorf("Expected non-zero exit code for invalid repository")
	}

	// Should contain error message in stderr
	if !strings.Contains(stderr, "Error:") && !strings.Contains(stdout, "Error:") {
		t.Errorf("Expected error message, but none found. Stdout: %s, Stderr: %s", stdout, stderr)
	}
}

func TestIntegration_ErrorHandling_EmptyRepo(t *testing.T) {
	stdout, stderr, exitCode := runGhDownload(t, "--repo", "")

	if exitCode == 0 {
		t.Errorf("Expected non-zero exit code for empty repository")
	}

	// Should contain error about repository being required
	errorOutput := stdout + stderr
	if !strings.Contains(errorOutput, "repository is required") {
		t.Errorf("Expected 'repository is required' error, but got: %s", errorOutput)
	}
}

func TestIntegration_SpecificTag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Use a known stable tag from cli/cli
	stdout, stderr, exitCode := runGhDownload(t,
		"--repo", "cli/cli",
		"--tag", "v2.0.0",
		"--list")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "v2.0.0") {
		t.Errorf("Expected output to contain tag v2.0.0")
	}

	if !strings.Contains(stdout, "Assets matching pattern '*':") {
		t.Errorf("Expected assets listing")
	}
}
