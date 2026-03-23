# ADR-002: Prefer Pure Go with Minimal External Dependencies

## Status
Accepted

## Context
We need to decide on the dependency strategy for the project, balancing between using external libraries and maintaining simplicity.

## Decision
We will prefer pure Go implementations using the standard library wherever possible, minimizing external dependencies.

### Allowed External Dependencies
- **gopkg.in/yaml.v3**: For YAML parsing (helmfile format)
- **github.com/spf13/cobra**: For CLI framework (industry standard)
- **github.com/spf13/viper**: For configuration management (integrates with Cobra)
- **github.com/stretchr/testify**: For testing assertions only

### Standard Library Usage
- `net/http`: HTTP client for repository queries
- `encoding/json`: JSON parsing for Helm repository index
- `os`, `path/filepath`: File system operations
- `sync`: Concurrency primitives
- `time`: Timestamp handling
- `log/slog`: Structured logging with configurable levels

## Consequences

### Positive
- Smaller binary size
- Fewer security vulnerabilities from dependencies
- Faster compilation
- Reduced maintenance burden
- Better long-term stability
- No dependency version conflicts

### Negative
- May need to implement some utilities from scratch
- Potentially more code to write and maintain
- Less feature-rich than specialized libraries

## Alternatives Considered
- **Heavy framework approach**: Using libraries like Helm SDK (rejected due to size and complexity)
- **Zero dependencies**: Including CLI parsing (rejected as too much boilerplate)
