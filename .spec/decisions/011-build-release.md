# ADR-011: Build and Release with GoReleaser

## Status
Accepted

## Context
We need a build and release strategy that supports cross-platform binaries, versioning, and automated release artifact generation.

## Decision
We will use [GoReleaser](https://goreleaser.com/) for building and releasing the application.

### Configuration File
`.goreleaser.yml` in project root.

### Key Features Used
- Cross-platform builds (Linux, macOS, Windows)
- Automated changelog generation from conventional commits
- Archive creation (tar.gz, zip)
- Checksum generation
- GitHub Releases integration

## Consequences

### Positive
- Single command to build all platform targets
- Automated changelog from conventional commits (aligns with ADR steering conventions)
- Reproducible builds
- Checksum verification for security
- Easy CI/CD integration

### Negative
- Additional tooling dependency
- Configuration maintenance needed

## Alternatives Considered
- **Custom Makefile**: Too much boilerplate for cross-platform builds
- **ko**: Focused on containers, not CLI binaries
- **Manual releases**: Error-prone, not reproducible

## Installation
```bash
# macOS
brew install goreleaser

# Or via Go
go install github.com/goreleaser/goreleaser/v2@latest
```
