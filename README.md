# HDC - Helmfile Dependency Checker

A standalone CLI tool that verifies Helm chart dependencies declared in helmfiles are up-to-date and actively maintained. No Helm or helmfile binaries required.

## Features

- Parse `helmfile.yaml` and `helmfile.d/*.yaml` files (including Go template expressions)
- Compare installed chart versions against latest available versions
- Detect unmaintained charts based on configurable age threshold
- Generate reports in JSON, Markdown, or HTML format
- Exclude specific charts or repositories from checks
- Concurrent repository queries for fast execution
- CI/CD friendly with non-zero exit codes on issues

## Exit Codes

HDC uses severity-based exit codes for CI/CD integration:

- **Exit code 0**: No issues found (all charts are up-to-date and actively maintained)
- **Exit code 1**: Warnings only (some charts are outdated but maintained)
- **Exit code 2**: Errors found (some charts are unmaintained or unreachable)

**Note:** Skipped releases (local charts) never affect the exit code.

## Limitations

- General authentication for HTTP/HTTPS Helm repositories is not supported.
- OCI registries may work when they support the standard bearer-token challenge flow, but this is treated as limited compatibility support, not as a full authentication feature.
- User-configured repository credentials are not supported. Authentication is read-only and limited to anonymous access or standard bearer-token challenges.

## Installation

### Homebrew (macOS / Linux)

```bash
brew install steffenrumpf/tap/hdc
```

To update to the latest version:

```bash
brew upgrade hdc
```

### From Release Archives

Download the appropriate archive for your platform from the [Releases](https://github.com/steffenrumpf/hdc/releases) page.

**macOS (Apple Silicon)**
```bash
curl -Lo hdc.tar.gz https://github.com/steffenrumpf/hdc/releases/latest/download/hdc_darwin_arm64.tar.gz
tar xzf hdc.tar.gz
sudo mv hdc /usr/local/bin/
```

**macOS (Intel)**
```bash
curl -Lo hdc.tar.gz https://github.com/steffenrumpf/hdc/releases/latest/download/hdc_darwin_amd64.tar.gz
tar xzf hdc.tar.gz
sudo mv hdc /usr/local/bin/
```

**Linux (amd64)**
```bash
curl -Lo hdc.tar.gz https://github.com/steffenrumpf/hdc/releases/latest/download/hdc_linux_amd64.tar.gz
tar xzf hdc.tar.gz
sudo mv hdc /usr/local/bin/
```

**Linux (arm64)**
```bash
curl -Lo hdc.tar.gz https://github.com/steffenrumpf/hdc/releases/latest/download/hdc_linux_arm64.tar.gz
tar xzf hdc.tar.gz
sudo mv hdc /usr/local/bin/
```

**Windows (amd64)**

Download `hdc_windows_amd64.zip` from the [Releases](https://github.com/steffenrumpf/hdc/releases) page, extract it, and add `hdc.exe` to your `PATH`.

### From Source

Requires Go 1.26+:

```bash
go install github.com/steffenrumpf/hdc/cmd@latest
```

## Usage

### Basic Check

```bash
hdc check helmfile.yaml
```

### Output Formats

```bash
# Markdown (default)
hdc check helmfile.yaml

# JSON
hdc check helmfile.yaml -o json

# HTML
hdc check helmfile.yaml -o html

# Write to file
hdc check helmfile.yaml -o json --output-file report.json
```

### Filtering Output

```bash
# Omit skipped releases (local charts) from report
hdc check helmfile.yaml --ignore-skipped
```

### Directory-Based Helmfiles

HDC supports both single-file and directory-based helmfile layouts:

```bash
# Single file
hdc check helmfile.yaml

# Directory
hdc check helmfile.d/
```

### Tuning

```bash
# Custom maintenance age threshold (default: 12 months)
hdc check helmfile.yaml --max-age 6

# Adjust concurrent repository queries (default: 5)
hdc check helmfile.yaml --concurrent 10

# Set request timeout in seconds (default: 30)
hdc check helmfile.yaml --timeout 60
```

### Logging

```bash
hdc check helmfile.yaml --log-level debug --log-format json
```

### Version

```bash
hdc version
```

## Configuration File

HDC looks for `.helmfile-checker.yaml` in the current directory or `$HOME/.config/helmfile-checker/`. You can also specify a config file explicitly with `-c`:

```bash
hdc check helmfile.yaml -c my-config.yaml
```

### Example `.helmfile-checker.yaml`

```yaml
log:
  level: info
  format: text

output:
  format: markdown
  ignore_skipped: false

checker:
  max_age_months: 12
  concurrent_requests: 5

repositories:
  timeout_seconds: 30
  skip_tls_verify: false

exclude:
  charts:
    - stable/prometheus
  repositories:
    - https://charts.helm.sh/stable
```

### Environment Variables

All configuration keys can be set via environment variables with the prefix `HELMFILE_CHECKER_` and dots replaced by underscores:

```bash
export HELMFILE_CHECKER_CHECKER_MAX_AGE_MONTHS=6
export HELMFILE_CHECKER_OUTPUT_FORMAT=json
export HELMFILE_CHECKER_OUTPUT_IGNORE_SKIPPED=true
```

### Configuration Precedence

Settings are resolved in this order (highest to lowest priority):

1. **CLI flags** (e.g., `--output`, `--ignore-skipped`)
2. **Environment variables** (e.g., `HELMFILE_CHECKER_OUTPUT_FORMAT`)
3. **Config file** (`.helmfile-checker.yaml` or path specified with `-c`)
4. **Defaults** (built-in values)

This means CLI flags always override config file and environment variable settings.

## Configuration Reference

| Config Key | CLI Flag | Environment Variable | Type | Default | Description |
|---|---|---|---|---|---|
| `log.level` | `--log-level` | `HELMFILE_CHECKER_LOG_LEVEL` | string | `info` | Log verbosity: `debug`, `info`, `warn`, `error` |
| `log.format` | `--log-format` | `HELMFILE_CHECKER_LOG_FORMAT` | string | `text` | Log output format: `text`, `json` |
| `output.format` | `-o`, `--output` | `HELMFILE_CHECKER_OUTPUT_FORMAT` | string | `markdown` | Report format: `json`, `markdown`, `html` |
| `output.file` | `--output-file` | `HELMFILE_CHECKER_OUTPUT_FILE` | string | _(stdout)_ | Write report to file instead of stdout |
| `output.ignore_skipped` | `--ignore-skipped` | `HELMFILE_CHECKER_OUTPUT_IGNORE_SKIPPED` | bool | `false` | Omit skipped findings (local charts) from report output |
| `checker.max_age_months` | `--max-age` | `HELMFILE_CHECKER_CHECKER_MAX_AGE_MONTHS` | int | `12` | Months since last chart update before flagged as unmaintained |
| `checker.concurrent_requests` | `--concurrent` | `HELMFILE_CHECKER_CHECKER_CONCURRENT_REQUESTS` | int | `5` | Number of concurrent repository index fetches |
| `repositories.timeout_seconds` | `--timeout` | `HELMFILE_CHECKER_REPOSITORIES_TIMEOUT_SECONDS` | int | `30` | HTTP timeout in seconds for repository requests |
| `repositories.skip_tls_verify` | — | `HELMFILE_CHECKER_REPOSITORIES_SKIP_TLS_VERIFY` | bool | `false` | Skip TLS certificate verification (not recommended) |
| `exclude.charts` | — | — | list | `[]` | Charts to exclude from checks (format: `repo/chart`) |
| `exclude.repositories` | — | — | list | `[]` | Repository URLs to skip entirely |

> **Note:** List values (`exclude.charts`, `exclude.repositories`) can only be set via config file, not through environment variables or CLI flags.

## Example Helmfile

```yaml
repositories:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami
  - name: ingress-nginx
    url: https://kubernetes.github.io/ingress-nginx

releases:
  - name: nginx-ingress
    namespace: ingress
    chart: ingress-nginx/ingress-nginx
    version: 4.0.1

  - name: postgresql
    namespace: database
    chart: bitnami/postgresql
    version: 11.6.0
```

## CLI Reference

```
hdc check <helmfile-path> [flags]

Flags:
  -o, --output string        Output format: json, markdown, html (default "markdown")
      --output-file string   Write report to file instead of stdout
      --ignore-skipped       Omit skipped findings from report (default false)
      --max-age int          Max chart age in months (default 12)
      --concurrent int       Concurrent repo queries (default 5)
      --timeout int          Request timeout in seconds (default 30)

Global Flags:
  -c, --config string        Config file path
      --log-level string     Log level: debug, info, warn, error (default "info")
      --log-format string    Log format: text, json (default "text")
```

## Exit Codes

HDC uses severity-based exit codes for CI/CD integration:

- **Exit code 0**: No issues found (all charts are up-to-date and actively maintained)
- **Exit code 1**: Warnings only (some charts are outdated but maintained)
- **Exit code 2**: Errors found (some charts are unmaintained or unreachable)

**Note:** Skipped findings (local charts) never affect the exit code.

## CI/CD Integration Examples

HDC is designed for pipeline use. Exit codes enable easy integration with CI systems to gate deployments based on severity.

### GitLab CI

```yaml
check-helm-dependencies:
  stage: test
  image: golang:1.26
  before_script:
    - go install github.com/steffenrumpf/hdc/cmd@latest
  script:
    - hdc check helmfile.yaml -o json --output-file hdc-report.json
  artifacts:
    paths:
      - hdc-report.json
    when: always
  allow_failure: true
```

To fail on warnings (outdated charts), add to your CI/CD script:
```bash
hdc check helmfile.yaml -o json --output-file hdc-report.json
EXIT_CODE=$?
if [ $EXIT_CODE -ge 1 ]; then
  echo "Dependency issues found (exit code: $EXIT_CODE)"
  exit $EXIT_CODE
fi
```

### GitHub Actions

```yaml
name: Helm Dependency Check
on:
  pull_request:
  schedule:
    - cron: '0 8 * * 1' # weekly on Monday

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install hdc
        run: |
          curl -Lo hdc.tar.gz https://github.com/steffenrumpf/hdc/releases/latest/download/hdc_linux_amd64.tar.gz
          tar xzf hdc.tar.gz
          sudo mv hdc /usr/local/bin/

      - name: Check dependencies
        run: hdc check helmfile.yaml -o markdown --output-file report.md

      - name: Upload report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: hdc-report
          path: report.md

      - name: Fail on errors (but allow warnings)
        run: |
          EXIT_CODE=$?
          if [ $EXIT_CODE -eq 2 ]; then
            echo "❌ Errors found in dependencies"
            exit 1
          fi
```

### Generic Shell (any CI system)

```bash
#!/usr/bin/env bash
set -euo pipefail

# Install
curl -sLo hdc.tar.gz https://github.com/steffenrumpf/hdc/releases/latest/download/hdc_linux_amd64.tar.gz
tar xzf hdc.tar.gz && chmod +x hdc

# Run check
./hdc check helmfile.yaml \
  --max-age 6 \
  -o json \
  --output-file hdc-report.json

EXIT_CODE=$?
echo "HDC exit code: $EXIT_CODE"

# 0 = no issues, 1 = warnings, 2 = errors
if [ $EXIT_CODE -eq 2 ]; then
  echo "❌ Errors found - failing build"
  exit 1
elif [ $EXIT_CODE -eq 1 ]; then
  echo "⚠️  Warnings found but build continues"
fi

echo "✅ Check complete"
```

### Scheduled Reporting (no failure on outdated)

For informational runs that produce a report without failing the pipeline:

```bash
hdc check helmfile.yaml -o html --output-file report.html
```

## Development

```bash
make build       # Build binary for current platform
make test        # Run tests with coverage
make lint        # Run golangci-lint
make lint-fix    # Run golangci-lint with auto-fix
make snapshot    # Build snapshot release with GoReleaser
make generate    # Generate mocks with mockery
make clean       # Remove build artifacts
```

## License

See [LICENSE](LICENSE).
