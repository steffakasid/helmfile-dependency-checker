# Helmfile Dependency Checker - Requirements Specification

## Project Overview
A tool to verify that Helm chart dependencies declared in helmfiles are up-to-date and actively maintained.

## Project Name
**hdc** (Helmfile Dependency Checker)

## User Stories

### US-001: Parse Helmfile Dependencies
**As a** DevOps engineer  
**I want to** automatically parse helmfile.yaml files  
**So that** I can extract all Helm chart dependencies without manual inspection

**Acceptance Criteria:**
- AC-001.1: Parse helmfile.yaml and helmfile.d/*.yaml files
- AC-001.2: Extract chart name, version, and repository URL
- AC-001.3: Support multiple helmfile formats (single file and directory structure)

### US-002: Check Chart Version Currency
**As a** DevOps engineer  
**I want to** compare current chart versions against latest available versions  
**So that** I can identify outdated dependencies

**Acceptance Criteria:**
- AC-002.1: Query Helm repositories for latest chart versions
- AC-002.2: Compare installed vs. available versions
- AC-002.3: Report version differences with semantic versioning awareness

### US-003: Verify Repository Availability
**As a** DevOps engineer  
**I want to** verify that chart repositories are accessible and active  
**So that** I can ensure dependencies are from maintained sources

**Acceptance Criteria:**
- AC-003.1: Check HTTP/HTTPS repository accessibility
- AC-003.2: Validate repository index.yaml exists
- AC-003.3: Report unreachable or deprecated repositories

### US-004: Assess Chart Maintenance Status
**As a** DevOps engineer  
**I want to** evaluate if charts are actively maintained  
**So that** I can avoid using abandoned dependencies

**Acceptance Criteria:**
- AC-004.1: Check last update timestamp of charts
- AC-004.2: Identify charts not updated in configurable timeframe (default: 12 months)
- AC-004.3: Flag potentially abandoned charts

### US-005: Generate Dependency Report
**As a** DevOps engineer  
**I want to** receive a comprehensive report of all findings  
**So that** I can prioritize dependency updates

**Acceptance Criteria:**
- AC-005.1: Generate report in multiple formats (JSON, Markdown, HTML)
- AC-005.2: Include status for each dependency (up-to-date, outdated, unmaintained)
- AC-005.3: Provide actionable recommendations

### US-006: CI/CD Integration
**As a** DevOps engineer  
**I want to** integrate the checker into CI/CD pipelines  
**So that** I can automate dependency verification

**Acceptance Criteria:**
- AC-006.1: Exit with non-zero code when issues found
- AC-006.2: Support configuration via CLI arguments and config file
- AC-006.3: Provide machine-readable output for automation

### US-007: Configure Thresholds and Rules
**As a** DevOps engineer  
**I want to** customize verification rules and thresholds  
**So that** I can adapt the tool to my organization's policies

**Acceptance Criteria:**
- AC-007.1: Configure maintenance age threshold
- AC-007.2: Set version difference tolerance (major/minor/patch)
- AC-007.3: Exclude specific charts or repositories from checks

### US-008: Homebrew Installation Support
**As a** macOS or Linux user  
**I want to** install hdc via Homebrew (`brew install`)  
**So that** I can easily install and update the tool using my preferred package manager

**Acceptance Criteria:**
- AC-008.1: WHEN a new GitHub release is published, THE GoReleaser pipeline SHALL generate a Homebrew formula and publish it to a dedicated Homebrew tap repository
- AC-008.2: THE Homebrew formula SHALL support both macOS (darwin) and Linux platforms on amd64 and arm64 architectures
- AC-008.3: WHEN a user runs `brew install steffenrumpf/tap/hdc`, THE Homebrew formula SHALL install the hdc binary to the user's PATH
- AC-008.4: THE GoReleaser configuration SHALL include a `brews` section that references the correct tap repository, binary name, and project metadata

### US-009: Parse Sub-Helmfiles via `helmfiles:` Tag
**As a** DevOps engineer  
**I want to** have hdc follow and parse sub-helmfiles referenced via the `helmfiles:` key  
**So that** all Helm chart dependencies across split helmfile configurations are checked

**Acceptance Criteria:**
- AC-009.1: WHEN a helmfile.yaml contains a `helmfiles:` key with glob patterns, THE Parser SHALL resolve the glob patterns relative to the parent helmfile directory and parse all matching files
- AC-009.2: WHEN a helmfile.yaml contains a `helmfiles:` key with explicit path entries (using `path:` key), THE Parser SHALL parse the referenced helmfile at the specified path
- AC-009.3: WHEN sub-helmfiles are parsed, THE Parser SHALL merge all repositories and releases from sub-helmfiles into the parent Helmfile result
- AC-009.4: WHEN a `helmfiles:` entry contains `selectors:` or `selectorsInherited:` fields, THE Parser SHALL ignore selector fields and parse the full content of the referenced file
- AC-009.5: IF a referenced sub-helmfile path does not exist, THEN THE Parser SHALL return a descriptive error including the missing file path
- AC-009.6: IF a glob pattern matches zero files, THEN THE Parser SHALL continue without error
- AC-009.7: THE Parser SHALL support both string entries (glob patterns) and map entries (with `path:` key) in the same `helmfiles:` list
- AC-009.8: WHEN sub-helmfiles contain their own `helmfiles:` keys, THE Parser SHALL recursively follow and parse nested sub-helmfile references
- AC-009.9: THE Helmfile model SHALL include a `Helmfiles` field to capture raw `helmfiles:` entries from the YAML structure

## Non-Functional Requirements

### NFR-001: Performance
- Process typical helmfile (<50 charts) in under 30 seconds
- Support concurrent repository queries

### NFR-002: Compatibility
- Support Helm 3.x chart repositories
- Compatible with helmfile v0.140.0+
- Cross-platform (Linux, macOS, Windows)

### NFR-003: Security
- Support authenticated Helm repositories
- Handle credentials securely (no logging)
- Validate SSL certificates by default

### NFR-004: Maintainability
- Modular architecture for easy extension
- Comprehensive error handling
- Logging with configurable verbosity

## Technical Constraints
- Must not require Helm or helmfile binaries installed
- Standalone executable preferred
- Minimal external dependencies
