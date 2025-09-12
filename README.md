# gh-download

An extension of [GitHub CLI](https://cli.github.com) for downloading a file from releases.

## Usage

### Install extension

```sh
gh extension install 23prime/gh-download
```

### Basic Usage

Download all assets from the latest release:

```sh
gh download owner/repo
```

Download all assets from a specific release:

```sh
gh download owner/repo v1.0.0
```

### Advanced Options

Download only specific files using patterns:

```sh
gh download --repo owner/repo --pattern "*.tar.gz"
gh download -R owner/repo -p "*.deb"
```

Download to a specific directory:

```sh
gh download --repo owner/repo --dir ./downloads
```

Download source code archive:

```sh
gh download --repo owner/repo --archive zip
gh download --repo owner/repo --archive tar.gz
```

### List Operations

List all releases without downloading:

```sh
gh download --repo owner/repo --releases
```

List assets from a release without downloading:

```sh
gh download --repo owner/repo --list
gh download --repo owner/repo --tag v1.0.0 --list --pattern "*.tar.gz"
```

### Command Reference

```txt
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
```

## For developers

### Pre requirements

- [Taskfile](https://taskfile.dev)
- [mise](https://mise.jdx.dev).

### Get start development

1. Setup project.

    ```sh
    task setup
    ```

2. Run application.

   ```sh
   task go:run
   ```

3. Check project.

    ```sh
    task check
    ```

For more information, run `task list`.

### Release

1. Build Go binary.

    ```sh
    task go:build
    ```

2. Check binary.

    ```sh
    $ file gh-download
    gh-download: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, BuildID[sha1]=2ad75d301c397993a80765a9131754b7d32b9ca2, with debug_info, not stripped
    ```

3. Tag.

    ```sh
    task tag:1.0.0
    ```
