package config_test

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/steffenrumpf/hdc/internal/config"
)

func TestInitConfig_Defaults(t *testing.T) {
	viper.Reset()

	cfg, err := config.InitConfig("")
	require.NoError(t, err)

	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "text", cfg.Log.Format)
	assert.Equal(t, "markdown", cfg.Output.Format)
	assert.Equal(t, 12, cfg.Checker.MaxAgeMonths)
	assert.Equal(t, 5, cfg.Checker.ConcurrentRequests)
	assert.Equal(t, 30, cfg.Repositories.TimeoutSeconds)
}

func TestInitConfig_InvalidFile(t *testing.T) {
	viper.Reset()

	_, err := config.InitConfig("nonexistent.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config")
}

func TestInitLogger_DoesNotPanic(t *testing.T) {
	viper.Reset()

	cfg, err := config.InitConfig("")
	require.NoError(t, err)

	assert.NotPanics(t, func() { config.InitLogger(cfg) })
}
