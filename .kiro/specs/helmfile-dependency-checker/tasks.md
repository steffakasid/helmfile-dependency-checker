# Tasks

## 1. Project Setup & Configuration
- [x] 1.1 Initialize Go module, Makefile, golangci-lint, goreleaser config
- [x] 1.2 Define data models: Helmfile, Release, Repository, Result, Finding, Status (`internal/models`)
- [x] 1.3 Define interfaces: Parser, FileReader, Client, HTTPClient, Checker, RepositoryClient, Writer
- [x] 1.4 Configure mockery and generate mocks for all interfaces
- [x] 1.5 Implement Config struct and InitConfig/InitLogger (`internal/config`)
- [x] 1.6 Write config tests (`config_test.go`)

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

## 6. CLI Wiring (US-006)
- [x] 6.1 Implement root command with Cobra and global flags (`cmd/main.go`)
- [x] 6.2 Implement `check` subcommand with flag binding
- [x] 6.3 Implement `version` subcommand with ldflags injection
- [x] 6.4 Wire config → parser → checker → report pipeline
- [x] 6.5 Handle exit codes for CI/CD integration (fail-on-outdated)

## 7. Quality & Verification
- [x] 7.1 Ensure all linting rules pass (`make lint`)
- [x] 7.2 Ensure all tests pass (`make test`)
- [x] 7.3 Verify >80% test coverage for core packages (parser, repository, checker, config)
- [x] 7.4 Verify snapshot build succeeds (`make snapshot`)

## 8. Documentation
- [x] 8.1 Update README.md with usage instructions and examples
- [x] 8.2 Add installation instructions (including Homebrew)
- [x] 8.3 Add configuration reference
- [x] 8.4 Add CI/CD integration examples

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
