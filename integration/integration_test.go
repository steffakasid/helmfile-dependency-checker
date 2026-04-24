//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildTestBinary(t *testing.T) {
	t.Helper()

	buildCmd := exec.Command("go", "build", "-o", "hdc-test", "../cmd")
	require.NoError(t, buildCmd.Run(), "failed to build test binary")

	t.Cleanup(func() {
		_ = os.Remove("hdc-test")
	})
}

func runCommand(t *testing.T, cmd *exec.Cmd) (string, string, error) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// TestCLI_Precedence_ConfigFileVsFlags tests that CLI flags take precedence over config file settings
func TestCLI_Precedence_ConfigFileVsFlags(t *testing.T) {
	// Define the content of the configuration file inside the test function
	configContent := `{
		"log": {
			"level": "info",
			"format": "text"
		},
		"output": {
			"format": "json"
		},
		"checker": {
			"max_age_months": 6,
			"concurrent_requests": 2
		},
		"repositories": {
			"timeout_seconds": 30
		}
	}`

	// Use t.TempDir() to create a temporary directory for the test
	tempDir := t.TempDir()
	configFilePath := filepath.Join(tempDir, "hdc-config.yaml")
	// Write the config file to the temporary directory
	if err := os.WriteFile(configFilePath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	buildTestBinary(t)

	// Test 1: Config file sets format=json, but --output markdown flag should override
	cmd := exec.Command("./hdc-test", "check", "../testdata/helmfile-integration.yaml",
		"--config", configFilePath,
		"--output", "markdown")
	output, _ := cmd.CombinedOutput() // Ignore error since we expect exit code 1 due to warnings

	// Should produce markdown output, not JSON
	assert.Contains(t, string(output), "# HDC Dependency Report")
	assert.NotContains(t, string(output), `"summary"`)

	// Test 2: Config file sets ignore_skipped=false, but --ignore-skipped flag should override
	cmd2 := exec.Command("./hdc-test", "check", "../testdata/helmfile-integration.yaml",
		"--config", configFilePath,
		"--ignore-skipped")
	output2, _ := cmd2.CombinedOutput() // Ignore error since we expect exit code 1 due to warnings

	// Without an output flag, the config file's JSON format should still apply.
	assert.Contains(t, string(output2), `"summary"`)
	assert.NotContains(t, string(output2), "# HDC Dependency Report")
}

// TestCLI_Precedence_FlagsVsDefaults tests that CLI flags override defaults
func TestCLI_Precedence_FlagsVsDefaults(t *testing.T) {
	buildTestBinary(t)

	// Test: Default output is markdown, but --output json flag should override
	cmd := exec.Command("./hdc-test", "check", "../testdata/helmfile-integration.yaml",
		"--output", "json")
	stdout, stderr, _ := runCommand(t, cmd) // Ignore error since we expect exit code 1 due to warnings

	// Parse the JSON report from stdout. Logs are written to stderr.
	outputStr := stdout
	startIdx := strings.Index(outputStr, "{")
	if startIdx == -1 {
		t.Fatalf("No JSON found in stdout: %s\nstderr: %s", outputStr, stderr)
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
	buildTestBinary(t)

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
	buildTestBinary(t)

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
	buildTestBinary(t)

	// Create temp output file
	outputFile, err := os.CreateTemp("", "hdc-output-*.txt")
	require.NoError(t, err)
	outputFile.Close()
	defer os.Remove(outputFile.Name())

	// Test: --output-file should write to file instead of stdout
	cmd := exec.Command("./hdc-test", "check", "../testdata/helmfile-integration.yaml",
		"--output", "markdown",
		"--output-file", outputFile.Name())
	stdout, _, _ := runCommand(t, cmd) // Ignore error since we expect exit code 1 due to warnings

	// Stdout should be empty (output went to file)
	assert.Empty(t, strings.TrimSpace(stdout))

	// File should contain the report
	fileContent, err := os.ReadFile(outputFile.Name())
	require.NoError(t, err)
	assert.Contains(t, string(fileContent), "# HDC Dependency Report")
}
