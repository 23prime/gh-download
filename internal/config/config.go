package config

import (
	"flag"
	"fmt"
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

func ParseArgs() Config {
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

func PrintUsage() {
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
