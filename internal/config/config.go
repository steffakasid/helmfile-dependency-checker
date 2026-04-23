package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for hdc.
type Config struct {
	Log struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"`
	} `mapstructure:"log"`
	Output struct {
		Format        string `mapstructure:"format"`
		File          string `mapstructure:"file"`
		IgnoreSkipped bool   `mapstructure:"ignore_skipped"`
	} `mapstructure:"output"`
	Checker struct {
		MaxAgeMonths       int  `mapstructure:"max_age_months"`
		FailOnOutdated     bool `mapstructure:"fail_on_outdated"`
		ConcurrentRequests int  `mapstructure:"concurrent_requests"`
	} `mapstructure:"checker"`
	Repositories struct {
		TimeoutSeconds int  `mapstructure:"timeout_seconds"`
		SkipTLSVerify  bool `mapstructure:"skip_tls_verify"`
	} `mapstructure:"repositories"`
	Exclude struct {
		Charts       []string `mapstructure:"charts"`
		Repositories []string `mapstructure:"repositories"`
	} `mapstructure:"exclude"`
}

// InitConfig loads configuration from file, environment, and defaults.
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

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "text")
	viper.SetDefault("output.format", "markdown")
	viper.SetDefault("checker.max_age_months", 12)
	viper.SetDefault("checker.concurrent_requests", 5)
	viper.SetDefault("repositories.timeout_seconds", 30)

	if err := viper.ReadInConfig(); err != nil {
		var configNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configNotFound) {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// InitLogger configures the global slog logger based on cfg.
func InitLogger(cfg *Config) {
	var level slog.Level

	switch cfg.Log.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if cfg.Log.Format == "json" {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	slog.SetDefault(slog.New(handler))
}
