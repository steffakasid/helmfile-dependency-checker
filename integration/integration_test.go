//go:build integration

package integration_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLI_Precedence_ConfigFileVsFlags tests that CLI flags take precedence over config file settings
func TestCLI_Precedence_ConfigFileVsFlags(t *testing.T) {
	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", "hdc-test", "../cmd")
	require.NoError(t, buildCmd.Run(), "failed to build test binary")
	defer os.Remove("hdc-test")

	// Create a temporary config file
	configContent := `output:
  format: json
  ignore_skipped: false
checker:
  max_age_months: 6
  concurrent_requests: 2
`
	configFile, err := os.CreateTemp("", "hdc-config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(configFile.Name())

	_, err = configFile.WriteString(configContent)
	require.NoError(t, err)
	configFile.Close()

	// Test 1: Config file sets format=json, but --output markdown flag should override
	cmd := exec.Command("./hdc-test", "check", "../testdata/helmfile-integration.yaml",
		"--config", configFile.Name(),
		"--output", "markdown")
	output, _ := cmd.CombinedOutput() // Ignore error since we expect exit code 1 due to warnings

	// Should produce markdown output, not JSON
	assert.Contains(t, string(output), "# HDC Dependency Report")
	assert.NotContains(t, string(output), `"summary"`)

	// Test 2: Config file sets ignore_skipped=false, but --ignore-skipped flag should override
	cmd2 := exec.Command("./hdc-test", "check", "../testdata/helmfile-integration.yaml",
		"--config", configFile.Name(),
		"--ignore-skipped")
	output2, _ := cmd2.CombinedOutput() // Ignore error since we expect exit code 1 due to warnings

	// Should produce markdown output (from config), but with ignore_skipped behavior
	assert.Contains(t, string(output2), "# HDC Dependency Report")
}

// TestCLI_Precedence_FlagsVsDefaults tests that CLI flags override defaults
func TestCLI_Precedence_FlagsVsDefaults(t *testing.T) {
	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", "hdc-test", "../cmd")
	require.NoError(t, buildCmd.Run(), "failed to build test binary")
	defer os.Remove("hdc-test")

	// Test: Default output is markdown, but --output json flag should override
	cmd := exec.Command("./hdc-test", "check", "../testdata/helmfile-integration.yaml",
		"--output", "json")
	output, _ := cmd.CombinedOutput() // Ignore error since we expect exit code 1 due to warnings

	// Extract JSON from output (skip log lines)
	outputStr := string(output)
	startIdx := strings.Index(outputStr, "{")
	if startIdx == -1 {
		t.Fatalf("No JSON found in output: %s", outputStr)
	}

	// Find the end of the JSON object (simple approach: find matching brace)
	braceCount := 0
	endIdx := startIdx
	for i := startIdx; i < len(outputStr); i++ {
		if outputStr[i] == '{' {
			braceCount++
		} else if outputStr[i] == '}' {
			braceCount--
			if braceCount == 0 {
				endIdx = i + 1
				break
			}
		}
	}

	if braceCount != 0 {
		t.Fatalf("Unmatched braces in JSON output: %s", outputStr[startIdx:])
	}

	jsonPart := outputStr[startIdx:endIdx]

	// Should produce JSON output
	var jsonOutput map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(jsonPart), &jsonOutput))
	assert.Contains(t, jsonOutput, "summary")
	assert.Contains(t, jsonOutput, "findings")
}

// TestCLI_LegacyFlagsRejected tests that removed legacy flags are properly rejected
func TestCLI_LegacyFlagsRejected(t *testing.T) {
	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", "hdc-test", "../cmd")
	require.NoError(t, buildCmd.Run(), "failed to build test binary")
	defer os.Remove("hdc-test")

	// Test: --fail-on-outdated flag should be rejected
	cmd := exec.Command("./hdc-test", "check", "../testdata/helmfile-integration.yaml",
		"--fail-on-outdated")
	output, err := cmd.CombinedOutput()
	require.Error(t, err, "expected command to fail with unknown flag")

	// Should contain error about unknown flag
	assert.Contains(t, string(output), "unknown flag")
	assert.Contains(t, string(output), "fail-on-outdated")
}

// TestCLI_ConfigFileNotFound tests that invalid config file content is handled
func TestCLI_ConfigFileInvalid(t *testing.T) {
	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", "hdc-test", "../cmd")
	require.NoError(t, buildCmd.Run(), "failed to build test binary")
	defer os.Remove("hdc-test")

	// Create a config file with invalid YAML
	configFile, err := os.CreateTemp("", "hdc-config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(configFile.Name())

	_, err = configFile.WriteString("invalid: yaml: content: [\n")
	require.NoError(t, err)
	configFile.Close()

	// Test: Invalid config file should produce error
	cmd := exec.Command("./hdc-test", "check", "../testdata/helmfile-integration.yaml",
		"--config", configFile.Name())
	output, err := cmd.CombinedOutput()
	require.Error(t, err, "expected command to fail with invalid config file")

	// Should contain error about config parsing
	assert.Contains(t, string(output), "failed to read config")
}

// TestCLI_OutputFile tests that --output-file flag works correctly
func TestCLI_OutputFile(t *testing.T) {
	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", "hdc-test", "../cmd")
	require.NoError(t, buildCmd.Run(), "failed to build test binary")
	defer os.Remove("hdc-test")

	// Create temp output file
	outputFile, err := os.CreateTemp("", "hdc-output-*.txt")
	require.NoError(t, err)
	outputFile.Close()
	defer os.Remove(outputFile.Name())

	// Test: --output-file should write to file instead of stdout
	cmd := exec.Command("./hdc-test", "check", "../testdata/helmfile-integration.yaml",
		"--output", "markdown",
		"--output-file", outputFile.Name())
	output, _ := cmd.CombinedOutput() // Ignore error since we expect exit code 1 due to warnings

	// Stdout should be empty (output went to file)
	assert.Empty(t, strings.TrimSpace(string(output)))

	// File should contain the report
	fileContent, err := os.ReadFile(outputFile.Name())
	require.NoError(t, err)
	assert.Contains(t, string(fileContent), "# HDC Dependency Report")
}
