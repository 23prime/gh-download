package download

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/23prime/gh-download/internal/config"
	"github.com/23prime/gh-download/internal/github"
	"github.com/cli/go-gh/v2/pkg/api"
)

func DownloadFromRelease(cfg config.Config) error {
	if cfg.Repository == "" {
		return fmt.Errorf("repository is required")
	}

	client, err := api.DefaultRESTClient()
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	if cfg.Releases {
		return github.ListReleases(client, cfg.Repository)
	}

	release, err := github.GetRelease(client, cfg.Repository, cfg.Tag)
	if err != nil {
		return fmt.Errorf("failed to get release: %w", err)
	}

	fmt.Printf("Release: %s", release.Name)
	if cfg.Tag != "" {
		fmt.Printf(" (tag: %s)", cfg.Tag)
	} else {
		fmt.Printf(" (latest)")
	}
	fmt.Printf(" from %s\n", cfg.Repository)

	if cfg.List {
		return github.ListAssets(release.Assets, cfg.Pattern)
	}

	if cfg.Archive != "" {
		return downloadArchive(client, cfg.Repository, cfg.Tag, cfg.Archive, cfg.Directory)
	}

	matchingAssets, err := github.FilterAssets(release.Assets, cfg.Pattern)
	if err != nil {
		return fmt.Errorf("failed to filter assets: %w", err)
	}

	if len(matchingAssets) == 0 {
		return fmt.Errorf("no assets found matching pattern '%s'", cfg.Pattern)
	}

	fmt.Printf("Found %d matching assets to download to %s:\n", len(matchingAssets), cfg.Directory)
	for _, asset := range matchingAssets {
		fmt.Printf("  - %s (%d bytes)\n", asset.Name, asset.Size)
	}

	return downloadAssets(matchingAssets, cfg.Directory)
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

	fullPath := filepath.Join(dir, filename)
	file, err := os.Create(fullPath)
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

	fmt.Printf("Downloaded archive: %s\n", fullPath)
	return nil
}

func downloadAssets(assets []github.Asset, dir string) error {
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

		fullPath := filepath.Join(dir, asset.Name)
		file, err := os.Create(fullPath)
		if err != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
			}
			return fmt.Errorf("failed to create file %s: %w", fullPath, err)
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
			return fmt.Errorf("failed to write %s: %w", fullPath, err)
		}

		fmt.Printf("done (%d bytes)\n", written)
	}

	fmt.Printf("Successfully downloaded %d assets to %s\n", len(assets), dir)
	return nil
}
