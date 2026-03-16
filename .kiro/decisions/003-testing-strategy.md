# ADR-003: Mandatory Unit Tests with Testify

## Status
Accepted

## Context
We need to establish a testing strategy that ensures code quality and reliability while maintaining developer productivity.

## Decision
Every file and function must have corresponding unit tests following these rules:

### Testing Requirements
1. **Coverage**: Every public function must have unit tests
2. **Coverage Threshold**: Minimum 80% coverage for core packages (`parser`, `repository`, `checker`, `config`)
3. **Test Files**: Tests placed alongside implementation as `*_test.go`
4. **Assertions**: Use `testify/assert` for general assertions
5. **Error Handling**: Use `testify/require` for error checks (fails fast)
6. **Table-Driven Tests**: Prefer table-driven tests for multiple scenarios
7. **Mocking**: Use mockery with the expecter pattern for interface-based mocks (`make generate`)

### Test Structure
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        // test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Testify Usage
- `assert.Equal()`, `assert.NotNil()`, etc. - continue test execution
- `require.NoError()`, `require.NotNil()`, etc. - stop test on failure

## Consequences

### Positive
- High code quality and reliability
- Easier refactoring with confidence
- Documentation through test examples
- Catches regressions early
- Testify provides readable assertions

### Negative
- More code to write and maintain
- Slower initial development
- Test maintenance overhead

## Alternatives Considered
- **Standard library only**: More verbose, less readable assertions
- **Optional tests**: Risk of untested code paths
- **Integration tests only**: Slower feedback, harder to debug
