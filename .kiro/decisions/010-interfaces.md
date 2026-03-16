# ADR-010: Interface Design for Testability

## Status
Accepted

## Context
External dependencies like HTTP clients and filesystem operations make unit testing hard without interfaces. We need a strategy that allows mocking in tests without introducing heavy mocking frameworks.

## Decision
Define narrow interfaces at the boundary of each module for any external or cross-module dependency. Generate mocks using [mockery](https://github.com/vektra/mockery) with the expecter pattern — no manual mock structs.

### Interface Locations
Interfaces are defined in the package that **consumes** them, not the package that implements them (Go best practice: "accept interfaces, return structs").

### Key Interfaces

**HTTP Client** (`internal/repository`)
```go
type HTTPClient interface {
    Get(url string) (*http.Response, error)
}
```

**Filesystem** (`internal/parser`)
```go
type FileReader interface {
    ReadFile(path string) ([]byte, error)
    ReadDir(path string) ([]os.DirEntry, error)
}
```

**Repository Client** (`internal/checker`)
```go
type RepositoryClient interface {
    FetchIndex(repoURL string) (*Index, error)
}
```

**Checker** (`internal/report` / `cmd`)
```go
type Checker interface {
    Check(releases []models.Release) ([]models.Finding, error)
}
```

**Parser** (`cmd`)
```go
type Parser interface {
    Parse(path string) (*models.Helmfile, error)
}
```

### Mock Generation
Mocks are generated via `make generate` (runs `go tool mockery`) and placed in a `mocks/` subdirectory next to the interface definition. Configuration lives in `.mockery.yaml` at the project root.

### Mock Pattern
```go
// Use the generated constructor — it registers cleanup automatically
reader := mocks.NewMockFileReader(t)

// Use the expecter API for type-safe call expectations
reader.EXPECT().ReadFile("helmfile.yaml").Return([]byte(content), nil)
```

### Rules
1. Interfaces should be small — prefer 1-2 methods (Interface Segregation Principle)
2. Define interfaces in the consuming package
3. Production structs implement interfaces implicitly
4. Generated mocks live in `mocks/` — never edit them manually, always re-generate
5. Always use `NewMock*(t)` constructors (auto-registers `AssertExpectations` cleanup)
6. Always use the expecter API (`EXPECT()`) over raw `On()` calls
7. Use `var _ InterfaceName = (*StructName)(nil)` compile-time checks where helpful

## Consequences

### Positive
- Type-safe expectations via expecter API — no stringly-typed `On("MethodName")` calls
- Auto-registered `AssertExpectations` via `NewMock*(t)` constructors — no manual cleanup
- Mocks are always in sync with interfaces — regenerate on change
- Encourages small, focused interfaces
- Compile-time interface satisfaction checks
- Clear boundaries between modules

### Negative
- Generated files must be committed and kept in sync
- `make generate` must be re-run after any interface change

## Alternatives Considered
- **Manual mocks**: No tooling dependency but verbose boilerplate and error-prone maintenance
- **gomock**: Similar generation approach but less idiomatic with testify
- **No interfaces**: Tight coupling, hard to unit test
