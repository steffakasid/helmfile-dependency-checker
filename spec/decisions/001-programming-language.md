# ADR-001: Use Go as Primary Programming Language

## Status
Accepted

## Context
We need to choose a programming language for the helmfile-dependency-checker CLI tool that provides good performance, cross-platform compatibility, and ease of distribution.

## Decision
We will use Go (Golang) as the primary programming language for this project.

## Consequences

### Positive
- Single binary distribution with no runtime dependencies
- Excellent cross-platform support (Linux, macOS, Windows)
- Strong standard library for HTTP, YAML parsing, and file operations
- Fast compilation and execution
- Built-in concurrency support for parallel repository queries
- Static typing reduces runtime errors
- Easy CI/CD integration

### Negative
- Less flexible than dynamic languages for rapid prototyping
- Verbose error handling
- Limited generics support (though improved in Go 1.18+)

## Alternatives Considered
- **Python**: Requires runtime, slower execution, dependency management complexity
- **Rust**: Steeper learning curve, longer compilation times
- **Node.js**: Requires runtime, larger distribution size
