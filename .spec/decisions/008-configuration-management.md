# ADR-008: Configuration Management with Viper

## Status
Accepted

## Context
We need to manage configuration from multiple sources (CLI flags, config file, environment variables) with a clear precedence order.

## Decision
We will use `github.com/spf13/viper` for configuration management, which integrates seamlessly with Cobra.

### Configuration Sources (Precedence Order)
1. **CLI flags** (highest priority)
2. **Environment variables** (prefix: `HELMFILE_CHECKER_`)
3. **Config file** (`.helmfile-checker.yaml` in current dir or `~/.config/helmfile-checker/config.yaml`)
4. **Defaults** (lowest priority)

### Configuration Structure
```yaml
# .helmfile-checker.yaml
log:
  level: info        # debug, info, warn, error
  format: text       # text, json

output:
  format: markdown   # json, markdown, html
  file: ""          # empty = stdout
  ignore_skipped: false

checker:
  max_age_months: 12
  concurrent_requests: 5

repositories:
  timeout_seconds: 30
  skip_tls_verify: false
  
exclude:
  charts: []
  repositories: []
```

### Configuration Model
```go
type Config struct {
    Log struct {
        Level  string
        Format string
    }
    Output struct {
        Format        string
        File          string
        IgnoreSkipped bool
    }
    Checker struct {
        MaxAgeMonths       int
        ConcurrentRequests int
    }
    Repositories struct {
        TimeoutSeconds int
        SkipTLSVerify  bool
    }
    Exclude struct {
        Charts       []string
        Repositories []string
    }
}
```

### CLI Flags Mapping
```go
// Persistent flags
--log-level, -l      → log.level
--log-format         → log.format
--config, -c         → config file path

// Command flags
--output, -o         → output.format
--output-file        → output.file
--ignore-skipped     → output.ignore_skipped
--max-age            → checker.max_age_months
--concurrent         → checker.concurrent_requests
--timeout            → repositories.timeout_seconds
--skip-tls-verify    → repositories.skip_tls_verify
```

### Environment Variables
```bash
HELMFILE_CHECKER_LOG_LEVEL=debug
HELMFILE_CHECKER_OUTPUT_FORMAT=json
HELMFILE_CHECKER_CHECKER_MAX_AGE_MONTHS=6
```

### Viper Integration
```go
func InitConfig(cfgFile string) (*Config, error) {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        viper.SetConfigName(".helmfile-checker")
        viper.SetConfigType("yaml")
        viper.AddConfigPath(".")
        viper.AddConfigPath("$HOME/.config/helmfile-checker")
    }
    
    viper.SetEnvPrefix("HELMFILE_CHECKER")
    viper.AutomaticEnv()
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    
    // Set defaults
    viper.SetDefault("log.level", "info")
    viper.SetDefault("log.format", "text")
    viper.SetDefault("checker.max_age_months", 12)
    viper.SetDefault("checker.concurrent_requests", 5)
    viper.SetDefault("repositories.timeout_seconds", 30)
    
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, fmt.Errorf("failed to read config: %w", err)
        }
    }
    
    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    return &cfg, nil
}
```

## Consequences

### Positive
- Single source of truth for configuration
- Flexible configuration sources
- Clear precedence order
- Integrates with Cobra
- Environment variable support for CI/CD
- Type-safe configuration struct

### Negative
- Adds external dependency (justified by value)
- Viper can be complex for simple use cases
- Need to maintain config schema

## Alternatives Considered
- **Manual flag parsing**: Too much boilerplate, no config file support
- **envconfig**: Only environment variables, no config file
- **Custom solution**: Reinventing the wheel
- **koanf**: Less popular, similar features

## Notes
Viper is maintained by the same author as Cobra and is the de facto standard for Go CLI configuration.
