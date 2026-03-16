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

### CI/CD Integration

Exit with non-zero code when outdated or unmaintained charts are found:

```bash
hdc check helmfile.yaml --fail-on-outdated
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

checker:
  max_age_months: 12
  fail_on_outdated: false
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
export HELMFILE_CHECKER_CHECKER_FAIL_ON_OUTDATED=true
```

Configuration precedence: CLI flags > environment variables > config file > defaults.

## Configuration Reference

| Config Key | CLI Flag | Environment Variable | Type | Default | Description |
|---|---|---|---|---|---|
| `log.level` | `--log-level` | `HELMFILE_CHECKER_LOG_LEVEL` | string | `info` | Log verbosity: `debug`, `info`, `warn`, `error` |
| `log.format` | `--log-format` | `HELMFILE_CHECKER_LOG_FORMAT` | string | `text` | Log output format: `text`, `json` |
| `output.format` | `-o`, `--output` | `HELMFILE_CHECKER_OUTPUT_FORMAT` | string | `markdown` | Report format: `json`, `markdown`, `html` |
| `output.file` | `--output-file` | `HELMFILE_CHECKER_OUTPUT_FILE` | string | _(stdout)_ | Write report to file instead of stdout |
| `checker.max_age_months` | `--max-age` | `HELMFILE_CHECKER_CHECKER_MAX_AGE_MONTHS` | int | `12` | Months since last chart update before flagged as unmaintained |
| `checker.fail_on_outdated` | `--fail-on-outdated` | `HELMFILE_CHECKER_CHECKER_FAIL_ON_OUTDATED` | bool | `false` | Exit non-zero when outdated or unmaintained charts are found |
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
      --max-age int          Max chart age in months (default 12)
      --fail-on-outdated     Exit non-zero if issues found
      --concurrent int       Concurrent repo queries (default 5)
      --timeout int          Request timeout in seconds (default 30)

Global Flags:
  -c, --config string        Config file path
      --log-level string     Log level: debug, info, warn, error (default "info")
      --log-format string    Log format: text, json (default "text")
```

## CI/CD Integration Examples

HDC is designed for pipeline use. The `--fail-on-outdated` flag causes a non-zero exit code when outdated or unmaintained charts are detected, making it easy to gate deployments.

### GitLab CI

```yaml
check-helm-dependencies:
  stage: test
  image: golang:1.26
  before_script:
    - go install github.com/steffenrumpf/hdc/cmd@latest
  script:
    - hdc check helmfile.yaml --fail-on-outdated -o json --output-file hdc-report.json
  artifacts:
    paths:
      - hdc-report.json
    when: always
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
        run: hdc check helmfile.yaml --fail-on-outdated -o markdown --output-file report.md

      - name: Upload report
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: hdc-report
          path: report.md
```

### Generic Shell (any CI system)

```bash
#!/usr/bin/env bash
set -euo pipefail

# Install
curl -sLo hdc.tar.gz https://github.com/steffenrumpf/hdc/releases/latest/download/hdc_linux_amd64.tar.gz
tar xzf hdc.tar.gz && chmod +x hdc

# Run check — exits non-zero on outdated charts
./hdc check helmfile.yaml \
  --fail-on-outdated \
  --max-age 6 \
  -o json \
  --output-file hdc-report.json
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
