package report_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/steffenrumpf/hdc/internal/models"
	"github.com/steffenrumpf/hdc/internal/report"
)

var testResult = &models.Result{
	Findings: []models.Finding{
		{
			Release:        models.Release{Name: "redis", Chart: "bitnami/redis"},
			Status:         models.StatusOutdated,
			CurrentVersion: "16.13.0",
			LatestVersion:  "17.0.0",
			Message:        "newer version 17.0.0 available",
		},
	},
}

func TestNew_InvalidFormat(t *testing.T) {
	_, err := report.New("xml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported report format")
}

func TestWrite_JSON(t *testing.T) {
	w, err := report.New("json")
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, w.Write(&buf, testResult))

	out := buf.String()
	assert.Contains(t, out, `"LatestVersion"`)
	assert.Contains(t, out, "17.0.0")
}

func TestWrite_Markdown(t *testing.T) {
	w, err := report.New("markdown")
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, w.Write(&buf, testResult))

	out := buf.String()
	assert.Contains(t, out, "# HDC Dependency Report")
	assert.Contains(t, out, "⚠️ redis (bitnami/redis)")
	assert.Contains(t, out, "Version: 16.13.0 → 17.0.0")
	assert.Contains(t, out, "newer version 17.0.0 available")
}

func TestWrite_HTML(t *testing.T) {
	w, err := report.New("html")
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, w.Write(&buf, testResult))

	out := buf.String()
	assert.True(t, strings.HasPrefix(out, "<!DOCTYPE html>"))
	assert.Contains(t, out, "redis")
	assert.Contains(t, out, "17.0.0")
}
