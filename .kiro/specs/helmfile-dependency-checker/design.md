# Helmfile Dependency Checker - Technical Design

## Overview
HDC is a standalone Go CLI tool that parses helmfile.yaml files, queries Helm chart repositories, and produces dependency health reports. It requires no Helm or helmfile binaries.

## Architecture

### High-Level Design

```
┌─────────┐     ┌────────┐     ┌────────────┐     ┌──────────┐     ┌────────┐
│   CLI   │────▶│ Config │────▶│   Parser   │────▶│ Checker  │────▶│ Report │
│  (Cobra)│     │ (Viper)│     │            │     │          │     │        │
└─────────┘     └────────┘     └────────────┘     └──────────┘     └────────┘
                                                       │
                                                       ▼
                                                  ┌──────────┐
                                                  │Repository│
                                                  │  Client  │
                                                  └──────────┘
```

### Module Structure

```
cmd/
└── main.go              # CLI entry point (Cobra root + subcommands)

internal/
├── config/              # Configuration management (Viper)
│   ├── config.go
│   └── config_test.go
├── models/              # Shared data structures
│   ├── helmfile.go      # Helmfile, Release, Repository types
│   └── result.go        # Result, Finding, Status types
├── parser/              # Helmfile parsing
│   ├── parser.go
│   └── parser_test.go
├── repository/          # Helm repository HTTP client
│   ├── client.go        # Client interface + HTTP implementation
│   ├── index.go         # YAML index parsing
│   └── client_test.go
├── checker/             # Version comparison + maintenance checks
│   ├── checker.go
│   └── checker_test.go
└── report/              # Multi-format output (JSON, Markdown, HTML)
    ├── report.go
    └── report_test.go
```

### Module Responsibilities

| Module | Responsibility | Key Interfaces |
|--------|---------------|----------------|
| `config` | Load config from file/env/defaults, init logger | `Config` struct |
| `models` | Shared types across all modules | `Helmfile`, `Release`, `Repository`, `Result`, `Finding`, `Status` |
| `parser` | Parse helmfile.yaml / helmfile.d/ into `Helmfile` | `Parser`, `FileReader` |
| `repository` | Fetch + parse Helm repo index.yaml over HTTP | `Client`, `HTTPClient` |
| `checker` | Compare versions, check maintenance age, apply exclusions | `Checker`, `RepositoryClient` |
| `report` | Format `Result` as JSON/Markdown/HTML | `Writer` |
| `cmd` | Wire everything together, handle CLI flags + exit codes | Cobra commands |

## Data Models

### Core Types (models package)

```go
type Repository struct {
    Name string `yaml:"name"`
    URL  string `yaml:"url"`
}

type Release struct {
    Name      string `yaml:"name"`
    Namespace string `yaml:"namespace"`
    Chart     string `yaml:"chart"`  // format: "repo/chart"
    Version   string `yaml:"version"`
}

type Helmfile struct {
    Repositories []Repository `yaml:"repositories"`
    Releases     []Release    `yaml:"releases"`
    Helmfiles    []any        `yaml:"helmfiles"`
}

type SubHelmfileEntry struct {
    Path               string   `yaml:"path"`
    Selectors          []string `yaml:"selectors"`
    SelectorsInherited bool     `yaml:"selectorsInherited"`
}

type Status string // "ok", "outdated", "unmaintained", "unreachable"

type Finding struct {
    Release        Release
    Status         Status
    CurrentVersion string
    LatestVersion  string
    LastUpdated    string
    Message        string
}

type Result struct {
    Findings []Finding
}
```

### Configuration (config package)

```go
type Config struct {
    Log          struct{ Level, Format string }
    Output       struct{ Format, File string }
    Checker      struct{ MaxAgeMonths int; FailOnOutdated bool; ConcurrentRequests int }
    Repositories struct{ TimeoutSeconds int; SkipTLSVerify bool }
    Exclude      struct{ Charts, Repositories []string }
}
```

Config sources (precedence): CLI flags > env vars (`HELMFILE_CHECKER_*`) > config file (`.helmfile-checker.yaml`) > defaults.

## Interfaces

All module boundaries use interfaces for testability. Mocks are generated via mockery.

```go
// parser
type Parser interface { Parse(path string) (*models.Helmfile, error) }
type FileReader interface { ReadFile(path string) ([]byte, error); ReadDir(path string) ([]os.DirEntry, error); Glob(pattern string) ([]string, error) }

// repository
type Client interface { FetchIndex(repoURL string) (*Index, error) }
type HTTPClient interface { Get(url string) (*http.Response, error) }

// checker
type Checker interface { Check(hf *models.Helmfile) (*models.Result, error) }
type RepositoryClient interface { FetchIndex(repoURL string) (*Index, error) }

// report
type Writer interface { Write(w io.Writer, result *models.Result) error }
```

## Key Algorithms

### Version Comparison
- Parse semver `vX.Y.Z` or `X.Y.Z` into `[major, minor, patch]`
- Strip pre-release suffixes before comparison
- Fall back to string inequality for non-semver versions

### Maintenance Check
- Compare `latest.Created` timestamp against `MaxAgeMonths` threshold
- Charts exceeding threshold are flagged as `unmaintained`

### Concurrency
- Concurrent repository fetches bounded by `ConcurrentRequests` semaphore
- Results collected via mutex-protected slice

## CLI Design

```
hdc check <helmfile-path> [flags]
  -o, --output         Output format: json, markdown, html (default: markdown)
  --output-file        Write report to file instead of stdout
  --max-age            Max chart age in months (default: 12)
  --fail-on-outdated   Exit non-zero if issues found
  --concurrent         Concurrent repo queries (default: 5)
  --timeout            Request timeout in seconds (default: 30)

hdc version            Print version info

Global flags:
  -c, --config         Config file path
  --log-level          Log level: debug, info, warn, error
  --log-format         Log format: text, json
```

## Build & Release
- GoReleaser for cross-platform builds and releases
- Version injected via ldflags at build time
- Mockery for interface mock generation

### Homebrew Distribution

GoReleaser's `brews` section automates Homebrew formula generation and publishing on each release.

**Tap Repository:** `github.com/steffenrumpf/homebrew-tap`  
This is a separate GitHub repository that hosts the generated formula. Users add it via `brew tap steffenrumpf/tap`.

**GoReleaser `brews` Configuration:**

```yaml
brews:
  - repository:
      owner: steffenrumpf
      name: homebrew-tap
    name: hdc
    homepage: "https://github.com/steffenrumpf/hdc"
    description: "Helmfile Dependency Checker - verify Helm chart dependencies are up-to-date"
    license: "MIT"
    install: |
      bin.install "hdc"
    test: |
      system "#{bin}/hdc", "version"
```

**How it works:**
1. On `goreleaser release`, GoReleaser builds archives for all OS/arch combos (already configured).
2. The `brews` section generates a Ruby formula file (`hdc.rb`) containing download URLs and SHA256 checksums for each archive.
3. GoReleaser pushes `hdc.rb` to the `homebrew-tap` repository via GitHub token.
4. Users install with: `brew tap steffenrumpf/tap && brew install hdc` (or shorthand `brew install steffenrumpf/tap/hdc`).

**Prerequisites:**
- The `homebrew-tap` repository must exist on GitHub before the first release.
- The `GITHUB_TOKEN` used by GoReleaser must have write access to the tap repository.

**Supported platforms:** darwin/amd64, darwin/arm64, linux/amd64, linux/arm64 (matching existing `builds` config).

## Sub-Helmfile Resolution (US-009)

### Overview

The `helmfiles:` key in helmfile.yaml references other helmfile files that should be parsed and merged. HDC follows these references to collect all repositories and releases across a split helmfile configuration.

### Model Changes

The `Helmfile` struct in `internal/models/helmfile.go` gains a new field to capture raw `helmfiles:` entries:

```go
type SubHelmfileEntry struct {
    Path               string   `yaml:"path"`
    Selectors          []string `yaml:"selectors"`
    SelectorsInherited bool     `yaml:"selectorsInherited"`
}

type Helmfile struct {
    Repositories []Repository       `yaml:"repositories"`
    Releases     []Release          `yaml:"releases"`
    Helmfiles    []any              `yaml:"helmfiles"`
}
```

The `Helmfiles` field uses `[]any` because entries can be either plain strings (glob patterns) or maps (with `path:`, `selectors:`, etc.). The parser normalizes these into `SubHelmfileEntry` values during processing.

### Parser Changes

The `parseFile` method is extended with a post-parse step that resolves `helmfiles:` entries:

```
parseFile(path)
  ├── read & unmarshal YAML
  ├── for each entry in hf.Helmfiles:
  │     ├── string entry → treat as glob pattern
  │     │     └── filepath.Glob(basedir/pattern) → parse each match
  │     └── map entry → extract "path" key
  │           └── parse basedir/path
  ├── merge sub-helmfile repos + releases into parent
  └── return merged Helmfile
```

### Glob Resolution

- Glob patterns are resolved relative to the directory of the parent helmfile using `filepath.Glob`.
- If a glob matches zero files, no error is raised (AC-009.6).
- The `FileReader` interface gains a `Glob(pattern string) ([]string, error)` method to keep filesystem operations mockable.

### Selector Handling

Selectors (`selectors:`, `selectorsInherited:`) are parsed but intentionally ignored. HDC needs to check ALL dependencies regardless of selector filtering (AC-009.4).

### Recursive Resolution

Sub-helmfiles may themselves contain `helmfiles:` keys. The parser recursively follows these references (AC-009.8). A visited-path set prevents infinite loops from circular references.

### Error Handling

- Missing explicit path → error with file path in message (AC-009.5)
- Missing glob match → silent continue (AC-009.6)
- Parse error in sub-helmfile → propagated with context (which parent referenced it)

### FileReader Interface Extension

```go
type FileReader interface {
    ReadFile(path string) ([]byte, error)
    ReadDir(path string) ([]os.DirEntry, error)
    Glob(pattern string) ([]string, error)
}
```

The `osFileReader` implementation delegates to `filepath.Glob`. Mocks can return controlled results for testing.

## Design Principles
1. Single Responsibility: each module has one clear purpose
2. Dependency Injection: interfaces at all boundaries
3. No Circular Dependencies: strict module hierarchy (models ← parser/repository ← checker ← report ← cmd)
4. Minimal Public API: export only what's necessary
