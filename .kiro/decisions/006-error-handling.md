# ADR-006: Explicit Error Handling with Context

## Status
Accepted

## Context
We need a consistent error handling strategy that provides useful debugging information while maintaining clean code.

## Decision
We will use explicit error handling with contextual information.

### Error Handling Rules
1. **Always check errors**: Never ignore returned errors
2. **Always wrap errors**: Use `fmt.Errorf("context: %w", err)` — never return bare errors
3. **No sensitive data in errors**: Error messages must not contain credentials, tokens, passwords, or personal data — sanitize before wrapping
4. **Custom errors**: Define sentinel errors for specific conditions
5. **No panic**: Avoid panic in library code, only in main for fatal errors
6. **Logging**: Use structured logging for error context — apply same sanitization rules

### Error Wrapping Pattern
```go
// Always wrap with context
func ParseHelmfile(path string) (*Helmfile, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read helmfile %s: %w", path, err)
    }
    // ...
}

// Sanitize sensitive data before including in errors
func fetchWithAuth(url, token string) (*http.Response, error) {
    resp, err := http.Get(url)
    if err != nil {
        // Do NOT include token in error message
        return nil, fmt.Errorf("failed to fetch repository %s: %w", url, err)
    }
    return resp, nil
}
```

### Sensitive Data Rules
- Never include credentials, tokens, or passwords in error messages
- Never include credentials, tokens, or passwords in log messages
- Repository URLs may be included but must be stripped of any embedded credentials (e.g. `https://user:password@repo.example.com` → `https://repo.example.com`)
- Use placeholder text if context about sensitive fields is needed: `"authentication failed for repository %s"` not `"token %s rejected"`

### Sentinel Errors
```go
var (
    ErrInvalidHelmfile = errors.New("invalid helmfile format")
    ErrRepositoryUnreachable = errors.New("repository unreachable")
    ErrChartNotFound = errors.New("chart not found in repository")
)
```

### Error Testing
```go
func TestParseHelmfile_InvalidFile(t *testing.T) {
    _, err := ParseHelmfile("nonexistent.yaml")
    require.Error(t, err)
    assert.Contains(t, err.Error(), "failed to read helmfile")
}
```

## Consequences

### Positive
- Clear error messages for debugging
- Easy to trace error origins
- Testable error conditions
- Standard Go idioms

### Negative
- Verbose error checking code
- Need to maintain error messages
- Error wrapping can create long messages

## Alternatives Considered
- **pkg/errors**: Additional dependency (rejected per ADR-002)
- **Panic-based**: Not idiomatic Go, hard to test
- **Error codes**: Less descriptive than messages
