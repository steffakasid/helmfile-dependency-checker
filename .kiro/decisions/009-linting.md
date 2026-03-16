# ADR-009: Code Quality with golangci-lint

## Status
Accepted

## Context
We need a comprehensive linting solution to maintain code quality, catch bugs early, and enforce consistent coding standards.

## Decision
We will use `golangci-lint` as our primary linting tool with a curated set of linters.

### golangci-lint Configuration
Configuration file: `.golangci.yml` in project root

```yaml
run:
  timeout: 5m
  tests: true
  modules-download-mode: readonly

linters:
  enable:
    - errcheck      # Check for unchecked errors
    - gosimple      # Simplify code
    - govet         # Vet examines Go source code
    - ineffassign   # Detect ineffectual assignments
    - staticcheck   # Advanced Go linter
    - unused        # Check for unused code
    - gofmt         # Check formatting
    - goimports     # Check imports
    - misspell      # Check for misspelled words
    - revive        # Fast, configurable, extensible linter
    - gosec         # Security issues
    - gocritic      # Opinionated linter
    - errname       # Check error naming conventions
    - errorlint     # Find code that will cause problems with error wrapping
    - exhaustive    # Check exhaustiveness of enum switch statements
    - gocyclo       # Cyclomatic complexity
    - godot         # Check if comments end in period
    - testpackage   # Makes you use separate _test package

linters-settings:
  gocyclo:
    min-complexity: 15
  
  revive:
    rules:
      - name: exported
        arguments:
          - "checkPrivateReceivers"
          - "sayRepetitiveInsteadOfStutters"
  
  errcheck:
    check-blank: true
    check-type-assertions: true

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck
        - gosec
        - gocyclo
  
  max-issues-per-linter: 0
  max-same-issues: 0

output:
  format: colored-line-number
  print-issued-lines: true
  print-linter-name: true
```

### Integration Points

**Local Development**
```bash
# Run linter
golangci-lint run

# Run with auto-fix
golangci-lint run --fix

# Run on specific files
golangci-lint run ./internal/parser/...
```

**Pre-commit Hook** (optional)
```bash
#!/bin/sh
golangci-lint run --new-from-rev=HEAD~1
```

**CI/CD Pipeline**
```yaml
# GitHub Actions / GitLab CI
- name: golangci-lint
  run: golangci-lint run --timeout=5m
```

**Makefile Integration**
```makefile
.PHONY: lint
lint:
	golangci-lint run

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix
```

### Key Linters Enabled

**Error Handling**
- `errcheck`: Ensures all errors are checked
- `errorlint`: Proper error wrapping usage
- `errname`: Error variable naming (err prefix)

**Code Quality**
- `staticcheck`: Advanced static analysis
- `gocritic`: Performance and style issues
- `revive`: Replacement for golint
- `gocyclo`: Complexity checks

**Security**
- `gosec`: Security vulnerabilities

**Style & Formatting**
- `gofmt`: Standard formatting
- `goimports`: Import organization
- `godot`: Comment punctuation

**Correctness**
- `govet`: Official Go tool
- `ineffassign`: Dead assignments
- `unused`: Unused code detection

## Consequences

### Positive
- Comprehensive code quality checks (50+ linters in one tool)
- Fast execution (runs linters in parallel)
- Consistent code style across team
- Catches bugs before runtime
- Security vulnerability detection
- Easy CI/CD integration
- Auto-fix capabilities for many issues
- Industry standard tool

### Negative
- Requires installation of golangci-lint binary
- Initial setup may find many issues in existing code
- Some linters can be opinionated
- Configuration maintenance needed

## Alternatives Considered
- **Individual linters**: Too many tools to manage, slow
- **go vet only**: Not comprehensive enough
- **staticcheck only**: Missing many useful checks
- **revive only**: Less comprehensive than golangci-lint

## Installation

```bash
# macOS
brew install golangci-lint

# Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Or via Go
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Notes
golangci-lint is the de facto standard for Go linting and is used by most major Go projects including Kubernetes, Prometheus, and Terraform.
