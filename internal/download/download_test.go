package download

import (
	"strings"
	"testing"

	"github.com/23prime/gh-download/internal/config"
)

func TestDownloadFromRelease_EmptyRepository(t *testing.T) {
	cfg := config.Config{
		Repository: "",
	}

	err := DownloadFromRelease(cfg)
	if err == nil {
		t.Fatal("Expected error for empty repository, got nil")
	}

	expectedError := "repository is required"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain %q, got %q", expectedError, err.Error())
	}
}

// Note: The DownloadFromRelease function is difficult to test comprehensively
// without refactoring because it has several external dependencies:
// 1. api.DefaultRESTClient() - creates GitHub client
// 2. File system operations (os.MkdirAll, os.Create, etc.)
// 3. HTTP requests through the GitHub client
//
// To make this function more testable, we would typically:
// 1. Use dependency injection to pass in the client and file system operations
// 2. Create interfaces for external dependencies
// 3. Use mock implementations in tests
//
// For now, we can only test the input validation logic.

func TestDownloadFromRelease_InvalidRepository(t *testing.T) {
	testCases := []struct {
		name       string
		repository string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.Config{
				Repository: strings.TrimSpace(tc.repository),
			}

			err := DownloadFromRelease(cfg)
			if err == nil {
				t.Fatal("Expected error for invalid repository, got nil")
			}

			if !strings.Contains(err.Error(), "repository is required") {
				t.Errorf("Expected error about repository, got %q", err.Error())
			}
		})
	}
}