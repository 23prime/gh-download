package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
)

type Config struct {
	Repository string
	Tag        string
	Pattern    string
	Directory  string
	Archive    string
	List       bool
	Releases   bool
	Help       bool
}

func main() {
	config := parseArgs()

	if config.Help {
		printUsage()
		return
	}

	if err := downloadFromRelease(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseArgs() Config {
	var config Config

	flag.StringVar(&config.Repository, "repo", "", "Repository in format owner/repo (required)")
	flag.StringVar(&config.Repository, "R", "", "Repository in format owner/repo (shorthand)")
	flag.StringVar(&config.Tag, "tag", "", "Release tag (defaults to latest)")
	flag.StringVar(&config.Tag, "t", "", "Release tag (shorthand)")
	flag.StringVar(&config.Pattern, "pattern", "*", "Glob pattern to match asset names")
	flag.StringVar(&config.Pattern, "p", "*", "Glob pattern to match asset names (shorthand)")
	flag.StringVar(&config.Directory, "dir", ".", "Directory to download files to")
	flag.StringVar(&config.Directory, "d", ".", "Directory to download files to (shorthand)")
	flag.StringVar(&config.Archive, "archive", "", "Download source archive (zip or tar.gz)")
	flag.BoolVar(&config.List, "list", false, "List release assets without downloading")
	flag.BoolVar(&config.List, "l", false, "List release assets without downloading (shorthand)")
	flag.BoolVar(&config.Releases, "releases", false, "List all releases")
	flag.BoolVar(&config.Releases, "r", false, "List all releases (shorthand)")
	flag.BoolVar(&config.Help, "help", false, "Show help")
	flag.BoolVar(&config.Help, "h", false, "Show help (shorthand)")

	flag.Parse()

	args := flag.Args()
	if len(args) > 0 && config.Repository == "" {
		config.Repository = args[0]
	}
	if len(args) > 1 && config.Tag == "" {
		config.Tag = args[1]
	}

	return config
}

func printUsage() {
	fmt.Println(`gh-download - Download files from GitHub releases

Usage:
  gh download [repository] [tag] [flags]

Arguments:
  repository    Repository in format owner/repo
  tag           Release tag (optional, defaults to latest)

Flags:
  -R, --repo string      Repository in format owner/repo
  -t, --tag string       Release tag (defaults to latest)
  -p, --pattern string   Glob pattern to match asset names (default "*")
  -d, --dir string       Directory to download files to (default ".")
      --archive string   Download source archive (zip or tar.gz)
  -l, --list             List release assets without downloading
  -r, --releases         List all releases
  -h, --help             Show help

Examples:
  gh download owner/repo                       # Download all assets from latest release
  gh download owner/repo v1.0.0                # Download all assets from v1.0.0
  gh download -R owner/repo -p "*.tar.gz"      # Download only .tar.gz files
  gh download --repo owner/repo --archive zip  # Download source code as zip
  gh download --repo owner/repo --list         # List all assets without downloading
  gh download --repo owner/repo --releases     # List all releases`)
}

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

func downloadFromRelease(config Config) error {
	if config.Repository == "" {
		return fmt.Errorf("repository is required")
	}

	client, err := api.DefaultRESTClient()
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	if config.Releases {
		return listReleases(client, config.Repository)
	}

	release, err := getRelease(client, config.Repository, config.Tag)
	if err != nil {
		return fmt.Errorf("failed to get release: %w", err)
	}

	fmt.Printf("Release: %s", release.Name)
	if config.Tag != "" {
		fmt.Printf(" (tag: %s)", config.Tag)
	} else {
		fmt.Printf(" (latest)")
	}
	fmt.Printf(" from %s\n", config.Repository)

	if config.List {
		return listAssets(release.Assets, config.Pattern)
	}

	if config.Archive != "" {
		return downloadArchive(client, config.Repository, config.Tag, config.Archive, config.Directory)
	}

	matchingAssets, err := filterAssets(release.Assets, config.Pattern)
	if err != nil {
		return fmt.Errorf("failed to filter assets: %w", err)
	}

	if len(matchingAssets) == 0 {
		return fmt.Errorf("no assets found matching pattern '%s'", config.Pattern)
	}

	fmt.Printf("Found %d matching assets to download to %s:\n", len(matchingAssets), config.Directory)
	for _, asset := range matchingAssets {
		fmt.Printf("  - %s (%d bytes)\n", asset.Name, asset.Size)
	}

	return downloadAssets(matchingAssets, config.Directory)
}

func getRelease(client *api.RESTClient, repo, tag string) (*Release, error) {
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

func downloadArchive(client *api.RESTClient, repo, tag, archiveFormat, dir string) error {
	if archiveFormat != "zip" && archiveFormat != "tar.gz" {
		return fmt.Errorf("archive format must be 'zip' or 'tar.gz'")
	}

	tagRef := tag
	if tagRef == "" {
		tagRef = "HEAD"
	}

	var endpoint string
	var filename string
	if archiveFormat == "zip" {
		endpoint = fmt.Sprintf("repos/%s/zipball/%s", repo, tagRef)
		filename = fmt.Sprintf("%s-%s.zip", strings.ReplaceAll(repo, "/", "-"), tagRef)
	} else {
		endpoint = fmt.Sprintf("repos/%s/tarball/%s", repo, tagRef)
		filename = fmt.Sprintf("%s-%s.tar.gz", strings.ReplaceAll(repo, "/", "-"), tagRef)
	}

	resp, err := client.Request("GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to download archive: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	filepath := filepath.Join(dir, filename)
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close file: %v\n", closeErr)
		}
	}()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Downloaded archive: %s\n", filepath)
	return nil
}

func filterAssets(assets []Asset, pattern string) ([]Asset, error) {
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

func listAssets(assets []Asset, pattern string) error {
	matchingAssets, err := filterAssets(assets, pattern)
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

func listReleases(client *api.RESTClient, repo string) error {
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

func downloadAssets(assets []Asset, dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create download client once with octet-stream header
	opts := api.ClientOptions{
		Headers: map[string]string{"Accept": "application/octet-stream"},
	}
	downloadClient, err := api.NewRESTClient(opts)
	if err != nil {
		return fmt.Errorf("failed to create download client: %w", err)
	}

	for _, asset := range assets {
		fmt.Printf("Downloading %s... ", asset.Name)

		resp, err := downloadClient.Request("GET", asset.URL, nil)
		if err != nil {
			return fmt.Errorf("failed to download %s: %w", asset.Name, err)
		}

		filepath := filepath.Join(dir, asset.Name)
		file, err := os.Create(filepath)
		if err != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
			}
			return fmt.Errorf("failed to create file %s: %w", filepath, err)
		}

		written, err := io.Copy(file, resp.Body)

		// Close resources immediately after use
		if closeErr := file.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close file: %v\n", closeErr)
		}
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}

		if err != nil {
			return fmt.Errorf("failed to write %s: %w", filepath, err)
		}

		fmt.Printf("done (%d bytes)\n", written)
	}

	fmt.Printf("Successfully downloaded %d assets to %s\n", len(assets), dir)
	return nil
}
