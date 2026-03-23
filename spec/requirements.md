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
- AC-006.3: WHEN the same setting is provided via both CLI arguments and config file, THE CLI argument SHALL take precedence
- AC-006.4: Provide machine-readable output for automation

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

### US-010: Exclude Local Chart References from Dependency Checking
**As a** DevOps engineer  
**I want to** have hdc skip local chart references (e.g. `./charts/mychart`) during dependency checking  
**So that** locally maintained charts are not flagged as outdated or unresolvable

**Acceptance Criteria:**
- AC-010.1: WHEN a release references a chart using a relative path (e.g. `./charts/mychart`), THE Checker SHALL exclude that release from dependency checking
- AC-010.2: WHEN a release references a chart using an absolute local path, THE Checker SHALL exclude that release from dependency checking
- AC-010.3: WHEN a local chart reference is excluded, THE Report SHALL include the release as `skipped` with a clear reason by default
- AC-010.4: THE CLI SHALL provide an `--ignore-skipped` option that omits `skipped` findings from report output
- AC-010.5: WHEN `--ignore-skipped` is enabled, skipped findings SHALL be omitted from both human-readable and machine-readable report output, but dependency checking behavior SHALL remain unchanged
- AC-010.6: `skipped` findings SHALL remain informational only and MUST NOT influence the exit code, regardless of whether `--ignore-skipped` is enabled
- AC-010.7: THE Parser SHALL identify local chart references by detecting path prefixes such as `./`, `../`, or `/` in the chart field of a release

### US-011: Support OCI Repository References
**As a** DevOps engineer  
**I want to** have hdc handle OCI-based Helm chart repository references (e.g. `oci://registry.example.com/charts/mychart`)  
**So that** charts hosted in OCI registries are included in dependency checking

**Acceptance Criteria:**
- AC-011.1: WHEN a release references a repository with an `oci://` scheme, THE Repository_Client SHALL fetch chart version metadata from the OCI registry
- AC-011.2: WHEN an OCI registry is queried, THE Repository_Client SHALL retrieve the list of available tags for the chart and determine the latest version using semantic versioning
- AC-011.3: IF an OCI registry is unreachable or returns an error, THEN THE Checker SHALL report the failure with a descriptive error including the OCI reference URL
- AC-011.4: THE Parser SHALL recognize `oci://` prefixed repository URLs and pass them to the OCI-aware Repository_Client
- AC-011.5: WHEN OCI chart versions are retrieved, THE Checker SHALL compare the current version against the latest available version using the same semantic versioning logic as HTTP/HTTPS repositories

## Non-Functional Requirements

### NFR-001: Performance
- Process typical helmfile (<50 charts) in under 30 seconds
- Support concurrent repository queries

### NFR-002: Compatibility
- Support Helm 3.x chart repositories
- Compatible with helmfile v0.140.0+
- Cross-platform (Linux, macOS, Windows)

### NFR-003: Security
- General authentication for Helm repositories SHALL NOT be supported in this version
- OCI registries MAY be accessed using anonymous bearer-token challenge flows when required by the registry, but this SHALL be treated as a compatibility mechanism rather than a user-configurable authentication feature
- The checker SHALL NOT require users to configure repository credentials
- Validate SSL certificates by default

### NFR-004: Maintainability
- Modular architecture for easy extension
- Comprehensive error handling
- Logging with configurable verbosity

## Technical Constraints
- Must not require Helm or helmfile binaries installed
- Standalone executable preferred
- Minimal external dependencies

## Distinct Exit Codes for Warnings and Errors

### User story
As a CI/CD engineer, I want the dependency checker to return distinct exit codes for warnings and errors so that GitLab CI can distinguish yellow jobs from red failures without parsing human-readable output.

### Acceptance criteria
- When the checker completes without warnings or errors, it must exit with code `0`.
- When the checker finds one or more warnings and no errors, it must exit with code `1`.
- When the checker finds one or more errors, it must exit with code `2`, even if warnings are also present.
- The three-level exit code scheme SHALL be the default and only supported CI/CD exit code behavior; legacy boolean fail modes SHALL NOT be preserved.
- The checker SHALL classify findings with status `outdated` as warnings.
- The checker SHALL classify findings with status `unmaintained` and `unreachable` as errors.
- The checker SHALL treat findings with status `ok` and `skipped` as informational only and they MUST NOT influence the exit code.
- Human-readable reports SHALL preserve the warning versus error distinction with different visual indicators, matching the status icons used by the current implementation.
- The CLI output must still include the warning and error counts so the exit code is not the only signal.
- The exit code behavior must be documented as part of the CLI requirements for CI/CD usage.
