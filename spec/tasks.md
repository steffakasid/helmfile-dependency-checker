# Tasks

## 1. Project Setup & Configuration
- [x] 1.1 Initialize Go module, Makefile, golangci-lint, goreleaser config
- [x] 1.2 Define data models: Helmfile, Release, Repository, Result, Finding, Status (`internal/models`)
- [x] 1.3 Define interfaces: Parser, FileReader, Client, HTTPClient, Checker, RepositoryClient, Writer
- [x] 1.4 Configure mockery and generate mocks for all interfaces
- [x] 1.5 Implement Config struct and InitConfig/InitLogger (`internal/config`)
- [x] 1.6 Write config tests (`config_test.go`)
- [x] 1.7 Add `output.ignore_skipped` to config structs, defaults, and bindings
- [x] 1.8 Remove legacy `checker.fail_on_outdated` config support and references

## 2. Helmfile Parsing (US-001)
- [x] 2.1 Implement helmfile.yaml single-file parsing (`internal/parser`)
- [x] 2.2 Implement helmfile.d/*.yaml directory parsing
- [x] 2.3 Handle Go template expressions in helmfiles
- [x] 2.4 Write parser tests (`parser_test.go`)

## 3. Repository Client (US-002, US-003)
- [x] 3.1 Implement HTTP client for fetching repository index.yaml (`internal/repository`)
- [x] 3.2 Implement YAML index parsing with JSON re-encoding strategy
- [x] 3.3 Implement LatestVersion lookup from parsed index
- [x] 3.4 Write repository client tests (`client_test.go`)

## 4. Dependency Checker (US-002, US-004, US-007)
- [x] 4.1 Implement semantic version comparison (`internal/checker`)
- [x] 4.2 Implement latest version lookup and outdated detection
- [x] 4.3 Implement maintenance status check (last update age)
- [x] 4.4 Implement chart and repository exclusion rules
- [x] 4.5 Implement concurrent release checking with semaphore
- [x] 4.6 Write checker tests (`checker_test.go`)

## 5. Report Generation (US-005)
- [x] 5.1 Implement JSON report writer (`internal/report`)
- [x] 5.2 Implement Markdown report writer
- [x] 5.3 Implement HTML report writer
- [x] 5.4 Write report tests (`report_test.go`)
- [x] 5.5 Add `ignore_skipped` filtering so skipped findings are omitted from JSON, Markdown, and HTML output when enabled
- [x] 5.6 Add warning and error summary counts to CLI-facing output while preserving existing status-specific visual indicators

## 6. CLI Wiring (US-006)
- [x] 6.1 Implement root command with Cobra and global flags (`cmd/main.go`)
- [x] 6.2 Implement `check` subcommand with flag binding
- [x] 6.3 Implement `version` subcommand with ldflags injection
- [x] 6.4 Wire config → parser → checker → report pipeline
- [x] 6.5 Replace legacy `fail-on-outdated` behavior with severity-based exit code classification (`0` clean, `1` warnings only, `2` any errors)
- [x] 6.6 Add `--ignore-skipped` CLI flag and bind it with CLI-over-config precedence
- [x] 6.7 Remove `--fail-on-outdated` from CLI help, flag binding, and runtime behavior

## 7. Quality & Verification
- [x] 7.1 Ensure all linting rules pass (`make lint`)
- [x] 7.2 Ensure all tests pass (`make test`)
- [ ] 7.3 Verify >80% test coverage for core packages (parser, repository, checker, config)
- [ ] 7.4 Verify snapshot build succeeds (`make snapshot`)
- [x] 7.5 Add test coverage for exit code classification (`0`, `1`, `2`) and severity mapping (`outdated` warning, `unmaintained`/`unreachable` error)
- [x] 7.6 Add test coverage for `ignore_skipped` filtering and confirm skipped findings never affect exit codes
- [ ] 7.7 Add test coverage for CLI-over-config precedence and confirm removed legacy flags/config keys are rejected or absent

## 8. Documentation
- [x] 8.1 Update README.md with usage instructions and examples
- [x] 8.2 Add installation instructions (including Homebrew)
- [x] 8.3 Add configuration reference
- [x] 8.4 Add CI/CD integration examples
- [x] 8.5 Update README.md to document the `0`/`1`/`2` exit code contract and remove `--fail-on-outdated` references
- [x] 8.6 Document `--ignore-skipped`, CLI-over-config precedence, and authentication limitations consistently across README and CLI help

## 9. Homebrew Support via GoReleaser (US-008)
- [ ] 9.1 Create `homebrew-tap` repository on GitHub (`steffenrumpf/homebrew-tap`)
- [ ] 9.2 Add `brews` section to `.goreleaser.yml` with tap repo, name, description, install, and test stanzas
- [ ] 9.3 Verify `GITHUB_TOKEN` has write access to the tap repository
- [ ] 9.4 Run `goreleaser release --snapshot --clean` to validate the generated formula
- [ ] 9.5 Update README.md with Homebrew installation instructions (`brew install steffenrumpf/tap/hdc`)

## 10. Sub-Helmfile Parsing (US-009)
- [x] 10.1 Add `Helmfiles []any` field to `models.Helmfile` and add `SubHelmfileEntry` struct to `internal/models/helmfile.go`
- [x] 10.2 Add `Glob(pattern string) ([]string, error)` method to `FileReader` interface and `osFileReader` in `internal/parser/parser.go`
- [x] 10.3 Regenerate mocks for updated `FileReader` interface (`make generate` or mockery)
- [x] 10.4 Implement sub-helmfile resolution in parser: normalize `helmfiles:` entries (string vs map), resolve globs, parse referenced files, merge repos and releases
- [x] 10.5 Implement recursive sub-helmfile resolution with visited-path cycle detection
- [x] 10.6 Add test: glob pattern resolves and merges sub-helmfile releases and repositories
- [x] 10.7 Add test: explicit `path:` entry with selectors is parsed (selectors ignored)
- [x] 10.8 Add test: missing explicit path returns descriptive error
- [x] 10.9 Add test: glob matching zero files continues without error
- [x] 10.10 Add test: recursive sub-helmfile references are followed
- [x] 10.11 Add test: mixed string and map entries in same `helmfiles:` list
- [x] 10.12 Add integration test using existing `testdata/helmfiles/helmfile.yaml` (which already contains `helmfiles:` entries)
- [x] 10.13 Ensure linting and all tests pass (`make lint && make test`)

## 11. Exclude Local Chart References (US-010)
- [x] 11.1 Add `StatusSkipped Status = "skipped"` constant to `internal/models/result.go`
- [x] 11.2 Implement `isLocalChart(chart string) bool` helper in `internal/checker/checker.go` detecting `./`, `../`, `/` prefixes
- [x] 11.3 Add local chart check to `checkRelease` flow — return `StatusSkipped` finding before `splitChart` is called
- [x] 11.4 Update report writers (JSON, Markdown, HTML) in `internal/report/report.go` to handle `StatusSkipped` findings
- [x] 11.5 Add unit test: `isLocalChart` returns true for `./charts/mychart`, `../shared/chart`, `/absolute/path/chart`
- [x] 11.6 Add unit test: `isLocalChart` returns false for `bitnami/redis`, `oci://registry/chart`, `mychart`
- [x] 11.7 Add unit test: checker skips local chart release and returns `StatusSkipped` without calling repository client
- [x] 11.8 Add unit test: report output includes or correctly handles `StatusSkipped` findings
- [x] 11.9 Add property test: *for any* string with path prefix (`./`, `../`, `/`), `isLocalChart` returns true; for any string without, returns false (Property 10)
- [x] 11.10 Ensure linting and all tests pass (`make lint && make test`)
- [x] 11.11 Add support for omitting `StatusSkipped` findings from report output when `ignore_skipped` is enabled
- [x] 11.12 Add unit tests verifying `ignore_skipped` removes skipped findings from JSON, Markdown, and HTML output only
- [ ] 11.13 Add property test verifying `ignore_skipped` affects serialized output but not warning/error counts or derived exit codes (Property 14)

## 12. OCI Repository Support (US-011)
- [x] 12.1 Add `FetchOCITags(ociURL string) (*Index, error)` method to `Client` interface in `internal/repository/client.go`
- [x] 12.2 Implement `parseOCIURL(ociURL string) (host, repo, chartName string, err error)` helper in `internal/repository/client.go`
- [x] 12.3 Implement `FetchOCITags` on `repoClient`: construct `https://{host}/v2/{repo}/tags/list` URL, HTTP GET, parse JSON response, filter to valid semver tags, build `Index`
- [x] 12.4 Implement `isOCIRepo(repoURL string) bool` helper in `internal/checker/checker.go`
- [x] 12.5 Add OCI branch to `checkRelease` flow: detect OCI repo URL → call `FetchOCITags` → extract chart name from OCI URL → compare versions using existing `isNewer`/`parseSemver`
- [x] 12.6 Regenerate mocks for updated `Client` interface (`make generate` or mockery)
- [x] 12.7 Add unit test: `parseOCIURL` correctly extracts host, repo, and chart name from valid `oci://` URLs
- [x] 12.8 Add unit test: `FetchOCITags` constructs correct API URL and parses tags response
- [x] 12.9 Add unit test: `FetchOCITags` filters non-semver tags (e.g. `latest`, `dev`) and returns only valid versions
- [x] 12.10 Add unit test: `isOCIRepo` returns true for `oci://` URLs and false for `http://`/`https://` URLs
- [x] 12.11 Add unit test: checker handles OCI release end-to-end — outdated, up-to-date, and unreachable cases
- [x] 12.12 Add property test: *for any* list of mixed semver/non-semver tags, OCI client returns only valid semver entries with correct latest (Property 12)
- [x] 12.13 Add property test: *for any* URL with `oci://` prefix, `isOCIRepo` returns true; for any other scheme, returns false (Property 13)
- [x] 12.14 Ensure linting and all tests pass (`make lint && make test`)
