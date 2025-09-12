package github

import (
	"bytes"
	"fmt"
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

// MockHTTPClient implements HTTPClient interface for testing
type MockHTTPClient struct {
	GetFunc func(endpoint string, response interface{}) error
}

func (m *MockHTTPClient) Get(endpoint string, response interface{}) error {
	if m.GetFunc != nil {
		return m.GetFunc(endpoint, response)
	}
	return nil
}

func TestGetRelease_LatestRelease(t *testing.T) {
	mockRelease := Release{
		ID:      12345,
		TagName: "v1.0.0",
		Name:    "Release v1.0.0",
		Assets: []Asset{
			{ID: 1, Name: "app.tar.gz", Size: 1024},
		},
	}

	mockClient := &MockHTTPClient{
		GetFunc: func(endpoint string, response interface{}) error {
			expectedEndpoint := "repos/owner/repo/releases/latest"
			if endpoint != expectedEndpoint {
				t.Errorf("Expected endpoint %q, got %q", expectedEndpoint, endpoint)
			}

			// Simulate API response by copying mock data
			if release, ok := response.(*Release); ok {
				*release = mockRelease
			}
			return nil
		},
	}

	release, err := GetRelease(mockClient, "owner/repo", "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if release.ID != mockRelease.ID {
		t.Errorf("Expected ID %d, got %d", mockRelease.ID, release.ID)
	}
	if release.TagName != mockRelease.TagName {
		t.Errorf("Expected TagName %q, got %q", mockRelease.TagName, release.TagName)
	}
	if release.Name != mockRelease.Name {
		t.Errorf("Expected Name %q, got %q", mockRelease.Name, release.Name)
	}
}

func TestGetRelease_SpecificTag(t *testing.T) {
	mockRelease := Release{
		ID:      67890,
		TagName: "v2.0.0",
		Name:    "Release v2.0.0",
		Assets:  []Asset{},
	}

	mockClient := &MockHTTPClient{
		GetFunc: func(endpoint string, response interface{}) error {
			expectedEndpoint := "repos/owner/repo/releases/tags/v2.0.0"
			if endpoint != expectedEndpoint {
				t.Errorf("Expected endpoint %q, got %q", expectedEndpoint, endpoint)
			}

			if release, ok := response.(*Release); ok {
				*release = mockRelease
			}
			return nil
		},
	}

	release, err := GetRelease(mockClient, "owner/repo", "v2.0.0")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if release.TagName != "v2.0.0" {
		t.Errorf("Expected TagName 'v2.0.0', got %q", release.TagName)
	}
}

func TestGetRelease_APIError(t *testing.T) {
	mockClient := &MockHTTPClient{
		GetFunc: func(endpoint string, response interface{}) error {
			return fmt.Errorf("API error: 404 Not Found")
		},
	}

	release, err := GetRelease(mockClient, "owner/repo", "v1.0.0")
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
	if release != nil {
		t.Errorf("Expected nil release on error, got %+v", release)
	}

	expectedError := "API error: 404 Not Found"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestGetRelease_ResponseParsing(t *testing.T) {
	mockRelease := Release{
		ID:          99999,
		TagName:     "v3.0.0",
		Name:        "Major Release v3.0.0",
		Body:        "This is a major release with breaking changes",
		Draft:       false,
		Prerelease:  true,
		CreatedAt:   "2023-12-01T10:00:00Z",
		PublishedAt: "2023-12-01T12:00:00Z",
		Assets: []Asset{
			{
				ID:                 1001,
				Name:               "app-linux-amd64.tar.gz",
				ContentType:        "application/x-gtar",
				Size:               2048576,
				BrowserDownloadURL: "https://github.com/owner/repo/releases/download/v3.0.0/app-linux-amd64.tar.gz",
				URL:                "https://api.github.com/repos/owner/repo/releases/assets/1001",
			},
			{
				ID:                 1002,
				Name:               "app-windows-amd64.zip",
				ContentType:        "application/zip",
				Size:               1843200,
				BrowserDownloadURL: "https://github.com/owner/repo/releases/download/v3.0.0/app-windows-amd64.zip",
				URL:                "https://api.github.com/repos/owner/repo/releases/assets/1002",
			},
		},
	}

	mockClient := &MockHTTPClient{
		GetFunc: func(endpoint string, response interface{}) error {
			if release, ok := response.(*Release); ok {
				*release = mockRelease
			}
			return nil
		},
	}

	release, err := GetRelease(mockClient, "owner/repo", "v3.0.0")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify all fields are correctly parsed
	if release.ID != mockRelease.ID {
		t.Errorf("Expected ID %d, got %d", mockRelease.ID, release.ID)
	}
	if release.TagName != mockRelease.TagName {
		t.Errorf("Expected TagName %q, got %q", mockRelease.TagName, release.TagName)
	}
	if release.Name != mockRelease.Name {
		t.Errorf("Expected Name %q, got %q", mockRelease.Name, release.Name)
	}
	if release.Body != mockRelease.Body {
		t.Errorf("Expected Body %q, got %q", mockRelease.Body, release.Body)
	}
	if release.Draft != mockRelease.Draft {
		t.Errorf("Expected Draft %t, got %t", mockRelease.Draft, release.Draft)
	}
	if release.Prerelease != mockRelease.Prerelease {
		t.Errorf("Expected Prerelease %t, got %t", mockRelease.Prerelease, release.Prerelease)
	}
	if release.CreatedAt != mockRelease.CreatedAt {
		t.Errorf("Expected CreatedAt %q, got %q", mockRelease.CreatedAt, release.CreatedAt)
	}
	if release.PublishedAt != mockRelease.PublishedAt {
		t.Errorf("Expected PublishedAt %q, got %q", mockRelease.PublishedAt, release.PublishedAt)
	}

	// Verify assets
	if len(release.Assets) != len(mockRelease.Assets) {
		t.Errorf("Expected %d assets, got %d", len(mockRelease.Assets), len(release.Assets))
	}

	for i, expectedAsset := range mockRelease.Assets {
		if i >= len(release.Assets) {
			t.Errorf("Missing asset at index %d", i)
			continue
		}
		actualAsset := release.Assets[i]

		if actualAsset.ID != expectedAsset.ID {
			t.Errorf("Asset %d: Expected ID %d, got %d", i, expectedAsset.ID, actualAsset.ID)
		}
		if actualAsset.Name != expectedAsset.Name {
			t.Errorf("Asset %d: Expected Name %q, got %q", i, expectedAsset.Name, actualAsset.Name)
		}
		if actualAsset.Size != expectedAsset.Size {
			t.Errorf("Asset %d: Expected Size %d, got %d", i, expectedAsset.Size, actualAsset.Size)
		}
	}
}

func TestFilterAssets_AllAssets(t *testing.T) {
	assets := []Asset{
		{Name: "app.tar.gz"},
		{Name: "app.zip"},
		{Name: "checksums.txt"},
	}

	// Test with "*" pattern
	filtered, err := FilterAssets(assets, "*")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(filtered) != 3 {
		t.Errorf("Expected 3 assets, got %d", len(filtered))
	}

	// Test with empty pattern
	filtered, err = FilterAssets(assets, "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(filtered) != 3 {
		t.Errorf("Expected 3 assets, got %d", len(filtered))
	}
}

func TestFilterAssets_SpecificPattern(t *testing.T) {
	assets := []Asset{
		{Name: "app-linux.tar.gz"},
		{Name: "app-windows.zip"},
		{Name: "app-macos.tar.gz"},
		{Name: "checksums.txt"},
	}

	// Test with "*.tar.gz" pattern
	filtered, err := FilterAssets(assets, "*.tar.gz")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(filtered) != 2 {
		t.Errorf("Expected 2 assets, got %d", len(filtered))
	}

	expectedNames := []string{"app-linux.tar.gz", "app-macos.tar.gz"}
	for i, asset := range filtered {
		if asset.Name != expectedNames[i] {
			t.Errorf("Expected asset name %q, got %q", expectedNames[i], asset.Name)
		}
	}
}

func TestFilterAssets_NoMatches(t *testing.T) {
	assets := []Asset{
		{Name: "app.tar.gz"},
		{Name: "app.zip"},
	}

	filtered, err := FilterAssets(assets, "*.exe")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(filtered) != 0 {
		t.Errorf("Expected 0 assets, got %d", len(filtered))
	}
}

func TestFilterAssets_InvalidPattern(t *testing.T) {
	assets := []Asset{
		{Name: "app.tar.gz"},
	}

	_, err := FilterAssets(assets, "[")
	if err == nil {
		t.Fatal("Expected error for invalid pattern, got nil")
	}

	expectedError := "invalid pattern '['"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain %q, got %q", expectedError, err.Error())
	}
}

func TestFilterAssets_ComplexPattern(t *testing.T) {
	assets := []Asset{
		{Name: "app-v1.0.0-linux-amd64.tar.gz"},
		{Name: "app-v1.0.0-windows-amd64.zip"},
		{Name: "app-v1.0.0-darwin-amd64.tar.gz"},
		{Name: "checksums-v1.0.0.txt"},
	}

	// Test with "app-*-linux-*" pattern
	filtered, err := FilterAssets(assets, "app-*-linux-*")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(filtered) != 1 {
		t.Errorf("Expected 1 asset, got %d", len(filtered))
	}
	if filtered[0].Name != "app-v1.0.0-linux-amd64.tar.gz" {
		t.Errorf("Expected 'app-v1.0.0-linux-amd64.tar.gz', got %q", filtered[0].Name)
	}
}

func TestListAssets_WithMatches(t *testing.T) {
	assets := []Asset{
		{Name: "app-linux.tar.gz", Size: 1024, ContentType: "application/x-gtar"},
		{Name: "app-windows.zip", Size: 2048, ContentType: "application/zip"},
		{Name: "checksums.txt", Size: 256, ContentType: "text/plain"},
	}

	output := captureOutput(func() {
		err := ListAssets(assets, "*.tar.gz")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	// Check output contains expected elements
	expectedStrings := []string{
		"Assets matching pattern '*.tar.gz':",
		"1. app-linux.tar.gz",
		"Size: 1024 bytes",
		"Content-Type: application/x-gtar",
		"Total: 1 assets",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it was missing", expected)
		}
	}
}

func TestListAssets_NoMatches(t *testing.T) {
	assets := []Asset{
		{Name: "app.tar.gz", Size: 1024, ContentType: "application/x-gtar"},
		{Name: "app.zip", Size: 2048, ContentType: "application/zip"},
	}

	output := captureOutput(func() {
		err := ListAssets(assets, "*.exe")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	expectedOutput := "No assets found matching pattern '*.exe'"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output to contain %q, got %q", expectedOutput, output)
	}
}

func TestListAssets_AllAssets(t *testing.T) {
	assets := []Asset{
		{Name: "app.tar.gz", Size: 1024, ContentType: "application/x-gtar"},
		{Name: "app.zip", Size: 2048, ContentType: "application/zip"},
	}

	output := captureOutput(func() {
		err := ListAssets(assets, "*")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	expectedStrings := []string{
		"Assets matching pattern '*':",
		"1. app.tar.gz",
		"2. app.zip",
		"Total: 2 assets",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it was missing", expected)
		}
	}
}

func TestListAssets_InvalidPattern(t *testing.T) {
	assets := []Asset{
		{Name: "app.tar.gz", Size: 1024, ContentType: "application/x-gtar"},
	}

	err := ListAssets(assets, "[")
	if err == nil {
		t.Fatal("Expected error for invalid pattern, got nil")
	}

	expectedError := "failed to filter assets"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain %q, got %q", expectedError, err.Error())
	}
}

func TestListReleases_WithReleases(t *testing.T) {
	mockReleases := []Release{
		{
			Name:        "Release v1.0.0",
			TagName:     "v1.0.0",
			Draft:       false,
			Prerelease:  false,
			PublishedAt: "2023-12-01T10:00:00Z",
			Assets:      []Asset{{Name: "app.tar.gz"}, {Name: "app.zip"}},
		},
		{
			Name:        "Release v0.9.0",
			TagName:     "v0.9.0",
			Draft:       true,
			Prerelease:  true,
			PublishedAt: "2023-11-15T15:30:00Z",
			Assets:      []Asset{{Name: "app.tar.gz"}},
		},
	}

	mockClient := &MockHTTPClient{
		GetFunc: func(endpoint string, response interface{}) error {
			expectedEndpoint := "repos/owner/repo/releases"
			if endpoint != expectedEndpoint {
				t.Errorf("Expected endpoint %q, got %q", expectedEndpoint, endpoint)
			}

			if releases, ok := response.(*[]Release); ok {
				*releases = mockReleases
			}
			return nil
		},
	}

	output := captureOutput(func() {
		err := ListReleases(mockClient, "owner/repo")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	// Check output contains expected elements
	expectedStrings := []string{
		"Releases for owner/repo:",
		"1. Release v1.0.0 (v1.0.0)",
		"Published: 2023-12-01",
		"Assets: 2",
		"2. Release v0.9.0 (v0.9.0) [draft, prerelease]",
		"Published: 2023-11-15",
		"Assets: 1",
		"Total: 2 releases",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it was missing", expected)
		}
	}
}

func TestListReleases_NoReleases(t *testing.T) {
	mockClient := &MockHTTPClient{
		GetFunc: func(endpoint string, response interface{}) error {
			if releases, ok := response.(*[]Release); ok {
				*releases = []Release{}
			}
			return nil
		},
	}

	output := captureOutput(func() {
		err := ListReleases(mockClient, "owner/repo")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	expectedOutput := "No releases found for owner/repo"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output to contain %q, got %q", expectedOutput, output)
	}
}

func TestListReleases_APIError(t *testing.T) {
	mockClient := &MockHTTPClient{
		GetFunc: func(endpoint string, response interface{}) error {
			return fmt.Errorf("API error: 404 Not Found")
		},
	}

	err := ListReleases(mockClient, "owner/repo")
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	expectedError := "failed to get releases"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain %q, got %q", expectedError, err.Error())
	}
}

func TestListReleases_SameTitleAndTag(t *testing.T) {
	mockReleases := []Release{
		{
			Name:        "v2.0.0",
			TagName:     "v2.0.0",
			Draft:       false,
			Prerelease:  false,
			PublishedAt: "2024-01-01T00:00:00Z",
			Assets:      []Asset{},
		},
	}

	mockClient := &MockHTTPClient{
		GetFunc: func(endpoint string, response interface{}) error {
			if releases, ok := response.(*[]Release); ok {
				*releases = mockReleases
			}
			return nil
		},
	}

	output := captureOutput(func() {
		err := ListReleases(mockClient, "owner/repo")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	// When name and tag are the same, tag should not be shown in parentheses
	if strings.Contains(output, "v2.0.0 (v2.0.0)") {
		t.Error("Expected tag not to be shown when it's the same as name")
	}
	if !strings.Contains(output, "1. v2.0.0") {
		t.Error("Expected release name to be shown")
	}
}
