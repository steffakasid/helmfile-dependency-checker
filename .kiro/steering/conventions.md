# Development Conventions

## Commit Messages
All commits must follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Format
```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Formatting, missing semicolons, etc. (no code change)
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `test`: Adding or updating tests
- `chore`: Build process, dependency updates, tooling
- `perf`: Performance improvements
- `ci`: CI/CD configuration changes

### Scopes
- `parser`: Helmfile parsing module
- `repository`: Helm repository client module
- `checker`: Version/maintenance checker module
- `report`: Report generation module
- `config`: Configuration management module
- `cmd`: CLI entry point
- `deps`: Dependency updates

### Examples
```
feat(parser): add support for helmfile.d directory structure
fix(repository): handle timeout errors gracefully
docs(readme): update installation instructions
test(checker): add table-driven tests for version comparison
chore(deps): update gopkg.in/yaml.v3 to v3.0.1
ci: add golangci-lint to GitHub Actions workflow
```

### Rules
- Subject line (type + scope + description) must not exceed 50 characters
- Body lines must be wrapped at 72 characters
- Description must be lowercase
- Description must not end with a period
- Use imperative mood ("add" not "added" or "adds")
- Breaking changes must append `!` after type/scope or include `BREAKING CHANGE:` in footer

## Code Conventions
- Follow standard Go conventions (`gofmt`, `goimports`)
- All linting rules defined in `.golangci.yml` must pass
- Every public function must have a corresponding `*_test.go` file
- Use `testify/assert` for assertions, `testify/require` for error checks
