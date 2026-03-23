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
	assert.Contains(t, out, "**Summary:** 1 checked, 0 ok, 1 warnings, 0 errors, 0 skipped")
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
	w, err := report.New("json")
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
	w, err := report.New("markdown")
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, w.Write(&buf, skippedResult))

	out := buf.String()
	assert.Contains(t, out, "⏭️ local-chart (./charts/mychart)")
	assert.Contains(t, out, "local chart reference")
	assert.Contains(t, out, "**Summary:** 1 checked, 0 ok, 0 warnings, 0 errors, 1 skipped")
}

func TestWrite_HTML_StatusSkipped(t *testing.T) {
	w, err := report.New("html")
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, w.Write(&buf, skippedResult))

	out := buf.String()
	assert.Contains(t, out, "local-chart")
	assert.Contains(t, out, "./charts/mychart")
	assert.Contains(t, out, "skipped")
	assert.Contains(t, out, `style="color:#888;background:#f5f5f5"`)
}

var mixedResult = &models.Result{
	Findings: []models.Finding{
		{
			Release: models.Release{Name: "redis", Chart: "bitnami/redis"},
			Status:  models.StatusOutdated,
			Message: "newer version available",
		},
		{
			Release: models.Release{Name: "local-chart", Chart: "./charts/mychart"},
			Status:  models.StatusSkipped,
			Message: "local chart reference",
		},
		{
			Release: models.Release{Name: "nginx", Chart: "bitnami/nginx"},
			Status:  models.StatusOK,
			Message: "up-to-date",
		},
	},
}

func TestFilterSkipped_RemovesSkippedFindings(t *testing.T) {
	filtered := report.FilterSkipped(mixedResult)

	require.Len(t, filtered.Findings, 2)
	assert.Equal(t, models.StatusOutdated, filtered.Findings[0].Status)
	assert.Equal(t, models.StatusOK, filtered.Findings[1].Status)
}

func TestFilterSkipped_JSON_OmitsSkipped(t *testing.T) {
	w, err := report.New("json")
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, w.Write(&buf, report.FilterSkipped(mixedResult)))

	out := buf.String()
	assert.Contains(t, out, "redis")
	assert.NotContains(t, out, "local-chart")
	assert.Contains(t, out, "nginx")
}

func TestFilterSkipped_Markdown_OmitsSkipped(t *testing.T) {
	w, err := report.New("markdown")
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, w.Write(&buf, report.FilterSkipped(mixedResult)))

	out := buf.String()
	assert.Contains(t, out, "redis")
	assert.NotContains(t, out, "local-chart")
	assert.Contains(t, out, "nginx")
}

func TestFilterSkipped_HTML_OmitsSkipped(t *testing.T) {
	w, err := report.New("html")
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, w.Write(&buf, report.FilterSkipped(mixedResult)))

	out := buf.String()
	assert.Contains(t, out, "redis")
	assert.NotContains(t, out, "local-chart")
	assert.Contains(t, out, "nginx")
}

func TestFilterSkipped_AllSkipped(t *testing.T) {
	filtered := report.FilterSkipped(skippedResult)
	assert.Empty(t, filtered.Findings)
}

func TestFilterSkipped_NoneSkipped(t *testing.T) {
	filtered := report.FilterSkipped(testResult)
	require.Len(t, filtered.Findings, 1)
	assert.Equal(t, models.StatusOutdated, filtered.Findings[0].Status)
}

func TestCountFindings(t *testing.T) {
	c := report.CountFindings(mixedResult)

	assert.Equal(t, 3, c.Total)
	assert.Equal(t, 1, c.OK)
	assert.Equal(t, 1, c.Warnings)
	assert.Equal(t, 0, c.Errors)
	assert.Equal(t, 1, c.Skipped)
}

func TestCountFindings_WithErrors(t *testing.T) {
	result := &models.Result{
		Findings: []models.Finding{
			{Status: models.StatusUnmaintained},
			{Status: models.StatusUnreachable},
			{Status: models.StatusOutdated},
			{Status: models.StatusOK},
		},
	}
	c := report.CountFindings(result)

	assert.Equal(t, 4, c.Total)
	assert.Equal(t, 1, c.OK)
	assert.Equal(t, 1, c.Warnings)
	assert.Equal(t, 2, c.Errors)
	assert.Equal(t, 0, c.Skipped)
}
