# ADR-007: Structured Logging with Standard Library

## Status
Accepted

## Context
We need a logging solution that provides configurable log levels, structured output, and integrates well with our pure Go approach.

## Decision
We will use Go's standard library `log/slog` (available since Go 1.21) for structured logging with configurable log levels.

### Log Levels
- **DEBUG**: Detailed information for debugging (HTTP requests, parsing details)
- **INFO**: General informational messages (processing files, found charts)
- **WARN**: Warning messages (deprecated repositories, slow responses)
- **ERROR**: Error conditions (failed requests, parsing errors)

### Logger Configuration
```go
// Configurable via CLI flags and config file
type LogConfig struct {
    Level  string // "debug", "info", "warn", "error"
    Format string // "text", "json"
}
```

### Usage Pattern
```go
import "log/slog"

slog.Debug("parsing helmfile", "path", helmfilePath)
slog.Info("found charts", "count", len(charts))
slog.Warn("repository slow to respond", "url", repoURL, "duration", duration)
slog.Error("failed to fetch repository", "url", repoURL, "error", err)
```

### Logger Initialization
```go
func InitLogger(cfg LogConfig) {
    var level slog.Level
    switch cfg.Level {
    case "debug":
        level = slog.LevelDebug
    case "info":
        level = slog.LevelInfo
    case "warn":
        level = slog.LevelWarn
    case "error":
        level = slog.LevelError
    default:
        level = slog.LevelInfo
    }
    
    var handler slog.Handler
    if cfg.Format == "json" {
        handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
    } else {
        handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
    }
    
    slog.SetDefault(slog.New(handler))
}
```

## Consequences

### Positive
- No external dependencies (stdlib only)
- Structured logging with key-value pairs
- Built-in log level support
- JSON output for machine parsing
- Performance optimized
- Context propagation support

### Negative
- Requires Go 1.21+ (using Go 1.26.1)
- Less feature-rich than specialized libraries
- No built-in log rotation (not needed for CLI tool)

## Alternatives Considered
- **zerolog**: Fast but external dependency
- **zap**: Feature-rich but external dependency
- **logrus**: Popular but external dependency, slower
- **log (old stdlib)**: No structured logging or levels
