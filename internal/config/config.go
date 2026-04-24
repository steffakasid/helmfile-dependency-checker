package config

import (
	"errors"
	"fmt"
	"log"
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
		MaxAgeMonths       int `mapstructure:"max_age_months"`
		ConcurrentRequests int `mapstructure:"concurrent_requests"`
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

// Simplify InitConfig by relying on viper's direct unmarshaling capabilities
func InitConfig(cfgFile string) (*Config, error) {
	log.Printf("Config file path: %s", cfgFile)

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

	if err := viper.ReadInConfig(); err != nil {
		var configNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configNotFound) {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "text")
	viper.SetDefault("output.format", "markdown")
	viper.SetDefault("checker.max_age_months", 12)
	viper.SetDefault("checker.concurrent_requests", 5)
	viper.SetDefault("repositories.timeout_seconds", 30)

	log.Printf("Pre-Unmarshal log.level: %s", viper.Get("log.level"))
	log.Printf("Pre-Unmarshal log.format: %s", viper.Get("log.format"))
	log.Printf("Pre-Unmarshal output.format: %s", viper.Get("output.format"))
	log.Printf("Pre-Unmarshal checker.max_age_months: %d", viper.Get("checker.max_age_months"))
	log.Printf("Pre-Unmarshal checker.concurrent_requests: %d", viper.Get("checker.concurrent_requests"))
	log.Printf("Pre-Unmarshal repositories.timeout_seconds: %d", viper.Get("repositories.timeout_seconds"))

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	log.Printf("Using config file: %s", viper.ConfigFileUsed())
	log.Printf("Config values: %+v", cfg)
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

// SetDefaultConfig sets default values for configuration fields.
func SetDefaultConfig() {
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "text")
	viper.SetDefault("output.format", "markdown")
	viper.SetDefault("checker.max_age_months", 12)
	viper.SetDefault("checker.concurrent_requests", 5)
	viper.SetDefault("repositories.timeout_seconds", 30)

	log.Printf("Default log.level: %s", viper.Get("log.level"))
	log.Printf("Default log.format: %s", viper.Get("log.format"))
	log.Printf("Default output.format: %s", viper.Get("output.format"))
	log.Printf("Default checker.max_age_months: %d", viper.Get("checker.max_age_months"))
	log.Printf("Default checker.concurrent_requests: %d", viper.Get("checker.concurrent_requests"))
	log.Printf("Default repositories.timeout_seconds: %d", viper.Get("repositories.timeout_seconds"))
}
