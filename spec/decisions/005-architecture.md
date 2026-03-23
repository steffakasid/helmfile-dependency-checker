# ADR-005: Modular Architecture with Clear Separation of Concerns

## Status
Accepted

## Context
We need to define the internal architecture and module structure for maintainability and testability.

## Decision
We will organize the codebase into focused modules with clear responsibilities.

### Module Structure
```
internal/
├── config/          # Configuration management
│   ├── config.go
│   └── config_test.go
├── parser/          # Helmfile parsing
│   ├── parser.go
│   └── parser_test.go
├── repository/      # Helm repository interaction
│   ├── client.go
│   ├── client_test.go
│   ├── index.go
│   └── index_test.go
├── checker/         # Version and maintenance checking
│   ├── checker.go
│   └── checker_test.go
├── report/          # Report generation
│   ├── report.go
│   ├── report_test.go
│   ├── json.go
│   ├── markdown.go
│   └── html.go
└── models/          # Shared data structures
    ├── chart.go
    ├── helmfile.go
    └── result.go

cmd/
└── main.go          # CLI entry point
```

### Module Responsibilities

**config**: Configuration management and initialization
- Load configuration from multiple sources
- Initialize logger based on config
- Provide configuration to other modules

**parser**: Parse helmfile.yaml files and extract chart dependencies
- Input: File path
- Output: List of chart references

**repository**: Query Helm repositories for chart information
- Fetch repository index.yaml
- Parse chart metadata
- Handle HTTP operations

**checker**: Compare versions and assess maintenance status
- Semantic version comparison
- Last update timestamp evaluation
- Generate findings

**report**: Format and output results
- Multiple output formats (JSON, Markdown, HTML)
- Consistent structure across formats

**models**: Shared data structures used across modules
- Chart, Release, Repository types
- Result and Finding types

### Design Principles
1. **Single Responsibility**: Each module has one clear purpose
2. **Dependency Injection**: Use interfaces for testability
3. **No Circular Dependencies**: Strict module hierarchy
4. **Minimal Public API**: Export only what's necessary

## Consequences

### Positive
- Easy to test individual components
- Clear code organization
- Easy to extend with new features
- Parallel development possible
- Reusable components

### Negative
- More files to navigate
- Potential over-engineering for small features
- Need to maintain interfaces

## Alternatives Considered
- **Flat structure**: All code in one package (rejected: hard to maintain)
- **Feature-based**: Group by feature not layer (rejected: more coupling)
