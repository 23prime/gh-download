package github

import (
	"fmt"
	"path"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
)

type Release struct {
	ID          int     `json:"id"`
	TagName     string  `json:"tag_name"`
	Name        string  `json:"name"`
	Body        string  `json:"body"`
	Draft       bool    `json:"draft"`
	Prerelease  bool    `json:"prerelease"`
	CreatedAt   string  `json:"created_at"`
	PublishedAt string  `json:"published_at"`
	Assets      []Asset `json:"assets"`
}

type Asset struct {
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	ContentType        string `json:"content_type"`
	Size               int    `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
	URL                string `json:"url"`
}

func GetRelease(client *api.RESTClient, repo, tag string) (*Release, error) {
	var endpoint string
	if tag == "" {
		endpoint = fmt.Sprintf("repos/%s/releases/latest", repo)
	} else {
		endpoint = fmt.Sprintf("repos/%s/releases/tags/%s", repo, tag)
	}

	var release Release
	err := client.Get(endpoint, &release)
	if err != nil {
		return nil, err
	}

	return &release, nil
}

func FilterAssets(assets []Asset, pattern string) ([]Asset, error) {
	if pattern == "*" || pattern == "" {
		return assets, nil
	}

	var matched []Asset
	for _, asset := range assets {
		match, err := path.Match(pattern, asset.Name)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern '%s': %w", pattern, err)
		}
		if match {
			matched = append(matched, asset)
		}
	}

	return matched, nil
}

func ListAssets(assets []Asset, pattern string) error {
	matchingAssets, err := FilterAssets(assets, pattern)
	if err != nil {
		return fmt.Errorf("failed to filter assets: %w", err)
	}

	if len(matchingAssets) == 0 {
		fmt.Printf("No assets found matching pattern '%s'\n", pattern)
		return nil
	}

	fmt.Printf("\nAssets matching pattern '%s':\n", pattern)
	for i, asset := range matchingAssets {
		fmt.Printf("%d. %s\n", i+1, asset.Name)
		fmt.Printf("   Size: %d bytes\n", asset.Size)
		fmt.Printf("   Content-Type: %s\n", asset.ContentType)
		if i < len(matchingAssets)-1 {
			fmt.Println()
		}
	}

	fmt.Printf("\nTotal: %d assets\n", len(matchingAssets))
	return nil
}

func ListReleases(client *api.RESTClient, repo string) error {
	endpoint := fmt.Sprintf("repos/%s/releases", repo)

	var releases []Release
	err := client.Get(endpoint, &releases)
	if err != nil {
		return fmt.Errorf("failed to get releases: %w", err)
	}

	if len(releases) == 0 {
		fmt.Printf("No releases found for %s\n", repo)
		return nil
	}

	fmt.Printf("Releases for %s:\n\n", repo)

	for i, release := range releases {
		fmt.Printf("%d. %s", i+1, release.Name)
		if release.TagName != "" && release.TagName != release.Name {
			fmt.Printf(" (%s)", release.TagName)
		}

		var status []string
		if release.Draft {
			status = append(status, "draft")
		}
		if release.Prerelease {
			status = append(status, "prerelease")
		}
		if len(status) > 0 {
			fmt.Printf(" [%s]", strings.Join(status, ", "))
		}
		fmt.Printf("\n")

		if release.PublishedAt != "" {
			fmt.Printf("   Published: %s\n", formatDate(release.PublishedAt))
		}

		fmt.Printf("   Assets: %d\n", len(release.Assets))

		if i < len(releases)-1 {
			fmt.Println()
		}
	}

	fmt.Printf("\nTotal: %d releases\n", len(releases))
	return nil
}

func formatDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	// Parse ISO 8601 date format and return a readable format
	// Input format: "2023-12-01T10:30:00Z"
	if len(dateStr) >= 10 {
		return dateStr[:10] // Return just the date part (YYYY-MM-DD)
	}
	return dateStr
}
