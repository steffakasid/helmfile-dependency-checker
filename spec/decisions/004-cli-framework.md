# ADR-004: Use Cobra for CLI Framework

## Status
Accepted

## Context
We need a CLI framework to handle command-line arguments, flags, and subcommands in a user-friendly way.

## Decision
We will use `github.com/spf13/cobra` as the CLI framework.

### CLI Structure
```
hdc check [flags] <helmfile-path>
hdc version
```

### Key Features Used
- Command and subcommand support
- Flag parsing (persistent and local flags)
- Automatic help generation
- POSIX-compliant flags
- Shell completion support

### Example Flags
- `--output, -o`: Output format (json, markdown, html)
- `--max-age`: Maximum chart age in months (default: 12)
- `--fail-on-outdated`: Exit with error if outdated charts found
- `--concurrent, -c`: Number of concurrent repository queries
- `--verbose, -v`: Verbose logging

## Consequences

### Positive
- Industry-standard CLI framework (used by kubectl, helm, etc.)
- Excellent documentation and community support
- Automatic help and usage generation
- Easy to extend with new commands
- Built-in shell completion

### Negative
- Adds external dependency (but justified by value)
- Opinionated structure
- Slightly larger binary size

## Alternatives Considered
- **flag (stdlib)**: Too basic, no subcommand support
- **urfave/cli**: Less popular, different conventions
- **Custom implementation**: Too much boilerplate
