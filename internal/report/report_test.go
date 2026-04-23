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
	_, err := report.New("xml", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported report format")
}

func TestWrite_JSON(t *testing.T) {
	w, err := report.New("json", false)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, w.Write(&buf, testResult))

	out := buf.String()
	assert.Contains(t, out, `"summary"`)
	assert.Contains(t, out, `"findings"`)
	assert.Contains(t, out, `"total": 1`)
	assert.Contains(t, out, `"warnings": 1`)
	assert.Contains(t, out, `"LatestVersion"`)
	assert.Contains(t, out, "17.0.0")
}

func TestWrite_Markdown(t *testing.T) {
	w, err := report.New("markdown", false)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, w.Write(&buf, testResult))

	out := buf.String()
	assert.Contains(t, out, "# HDC Dependency Report")
	assert.Contains(t, out, "**Summary:** 1 releases checked, 1 ⚠️ warnings")
	assert.Contains(t, out, "⚠️ redis (bitnami/redis)")
	assert.Contains(t, out, "Version: 16.13.0 → 17.0.0")
	assert.Contains(t, out, "newer version 17.0.0 available")
}

func TestWrite_HTML(t *testing.T) {
	w, err := report.New("html", false)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, w.Write(&buf, testResult))

	out := buf.String()
	assert.True(t, strings.HasPrefix(out, "<!DOCTYPE html>"))
	assert.Contains(t, out, "<strong>Summary:</strong> 1 releases checked, 1 ⚠️ warnings")
	assert.Contains(t, out, "redis")
	assert.Contains(t, out, "17.0.0")
}

var skippedResult = &models.Result{
	Findings: []models.Finding{
		{
			Release:        models.Release{Name: "local-chart", Chart: "./charts/mychart"},
			Status:         models.StatusSkipped,
			CurrentVersion: "1.0.0",
			LatestVersion:  "",
			Message:        "local chart reference",
		},
	},
}

func TestWrite_JSON_StatusSkipped(t *testing.T) {
	w, err := report.New("json", false)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, w.Write(&buf, skippedResult))

	out := buf.String()
	assert.Contains(t, out, `"Status"`)
	assert.Contains(t, out, "skipped")
	assert.Contains(t, out, "local-chart")
	assert.Contains(t, out, "./charts/mychart")
}

func TestWrite_Markdown_StatusSkipped(t *testing.T) {
	w, err := report.New("markdown", false)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, w.Write(&buf, skippedResult))

	out := buf.String()
	assert.Contains(t, out, "⏭️ local-chart (./charts/mychart)")
	assert.Contains(t, out, "local chart reference")
}

func TestWrite_HTML_StatusSkipped(t *testing.T) {
	w, err := report.New("html", false)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, w.Write(&buf, skippedResult))

	out := buf.String()
	assert.Contains(t, out, "local-chart")
	assert.Contains(t, out, "./charts/mychart")
	assert.Contains(t, out, "skipped")
	assert.Contains(t, out, `style="color:#888;background:#f5f5f5"`)
}
func TestWrite_JSON_IgnoreSkipped(t *testing.T) {
	w, err := report.New("json", true)
	require.NoError(t, err)

	resultWithSkipped := &models.Result{
		Findings: []models.Finding{
			{Status: models.StatusOK},
			{Status: models.StatusSkipped},
			{Status: models.StatusOutdated},
		},
	}

	var buf bytes.Buffer
	require.NoError(t, w.Write(&buf, resultWithSkipped))

	out := buf.String()
	// Summary should include all findings
	assert.Contains(t, out, `"total": 3`)
	assert.Contains(t, out, `"ok": 1`)
	assert.Contains(t, out, `"warnings": 1`)
	assert.Contains(t, out, `"skipped": 1`)
	// Findings array should exclude skipped
	assert.Contains(t, out, `"Status": "ok"`)
	assert.NotContains(t, out, `"Status": "skipped"`)
	assert.Contains(t, out, `"Status": "outdated"`)
}

func TestWrite_Markdown_IgnoreSkipped(t *testing.T) {
	w, err := report.New("markdown", true)
	require.NoError(t, err)

	resultWithSkipped := &models.Result{
		Findings: []models.Finding{
			{Release: models.Release{Name: "ok", Chart: "ok/chart"}, Status: models.StatusOK},
			{Release: models.Release{Name: "skipped", Chart: "skipped/chart"}, Status: models.StatusSkipped},
			{Release: models.Release{Name: "outdated", Chart: "outdated/chart"}, Status: models.StatusOutdated},
		},
	}

	var buf bytes.Buffer
	require.NoError(t, w.Write(&buf, resultWithSkipped))

	out := buf.String()
	assert.Contains(t, out, "ok")
	assert.NotContains(t, out, "skipped")
	assert.Contains(t, out, "outdated")
}

func TestWrite_HTML_IgnoreSkipped(t *testing.T) {
	w, err := report.New("html", true)
	require.NoError(t, err)

	resultWithSkipped := &models.Result{
		Findings: []models.Finding{
			{Release: models.Release{Name: "ok", Chart: "ok/chart"}, Status: models.StatusOK},
			{Release: models.Release{Name: "skipped", Chart: "skipped/chart"}, Status: models.StatusSkipped},
			{Release: models.Release{Name: "outdated", Chart: "outdated/chart"}, Status: models.StatusOutdated},
		},
	}

	var buf bytes.Buffer
	require.NoError(t, w.Write(&buf, resultWithSkipped))

	out := buf.String()
	assert.Contains(t, out, "ok")
	assert.NotContains(t, out, "skipped")
	assert.Contains(t, out, "outdated")
}
