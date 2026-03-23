package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/steffenrumpf/hdc/internal/models"
)

// Writer formats and writes a Result to an io.Writer.
type Writer interface {
	Write(w io.Writer, result *models.Result) error
}

// New returns a Writer for the given format ("json", "markdown", "html").
func New(format string) (Writer, error) {
	switch strings.ToLower(format) {
	case "json":
		return &jsonWriter{}, nil
	case "markdown":
		return &markdownWriter{}, nil
	case "html":
		return &htmlWriter{}, nil
	default:
		return nil, fmt.Errorf("unsupported report format %q", format)
	}
}

// FilterSkipped returns a new Result with StatusSkipped findings removed.
func FilterSkipped(result *models.Result) *models.Result {
	filtered := make([]models.Finding, 0, len(result.Findings))
	for _, f := range result.Findings {
		if f.Status != models.StatusSkipped {
			filtered = append(filtered, f)
		}
	}

	return &models.Result{Findings: filtered}
}

// Counts holds summary counters for a Result.
type Counts struct {
	OK           int
	Warnings     int
	Errors       int
	Skipped      int
	Total        int
}

// CountFindings returns severity-classified counts from a Result.
func CountFindings(result *models.Result) Counts {
	var c Counts
	c.Total = len(result.Findings)

	for _, f := range result.Findings {
		switch f.Status {
		case models.StatusOK:
			c.OK++
		case models.StatusOutdated:
			c.Warnings++
		case models.StatusUnmaintained, models.StatusUnreachable:
			c.Errors++
		case models.StatusSkipped:
			c.Skipped++
		}
	}

	return c
}

// --- JSON ---

type jsonWriter struct{}

func (j *jsonWriter) Write(w io.Writer, result *models.Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	if err := enc.Encode(result.Findings); err != nil {
		return fmt.Errorf("failed to write json report: %w", err)
	}

	return nil
}

// --- Markdown ---

type markdownWriter struct{}

var statusIcon = map[models.Status]string{
	models.StatusOK:           "✅",
	models.StatusOutdated:     "⚠️",
	models.StatusUnmaintained: "🔴",
	models.StatusUnreachable:  "❌",
	models.StatusSkipped:      "⏭️",
}

func (m *markdownWriter) Write(w io.Writer, result *models.Result) error {
	if _, err := fmt.Fprint(w, "# HDC Dependency Report\n\n"); err != nil {
		return fmt.Errorf("failed to write markdown report: %w", err)
	}

	for _, f := range result.Findings {
		icon := statusIcon[f.Status]
		if icon == "" {
			icon = "❓"
		}

		if _, err := fmt.Fprintf(w, "%s %s (%s)\n", icon, f.Release.Name, f.Release.Chart); err != nil {
			return fmt.Errorf("failed to write markdown report: %w", err)
		}
		if _, err := fmt.Fprintf(w, "  Version: %s → %s\n", f.CurrentVersion, f.LatestVersion); err != nil {
			return fmt.Errorf("failed to write markdown report: %w", err)
		}
		if f.Message != "" {
			if _, err := fmt.Fprintf(w, "  %s\n", f.Message); err != nil {
				return fmt.Errorf("failed to write markdown report: %w", err)
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return fmt.Errorf("failed to write markdown report: %w", err)
		}
	}

	c := CountFindings(result)
	if _, err := fmt.Fprintf(w, "---\n**Summary:** %d checked, %d ok, %d warnings, %d errors, %d skipped\n",
		c.Total, c.OK, c.Warnings, c.Errors, c.Skipped); err != nil {
		return fmt.Errorf("failed to write markdown report: %w", err)
	}

	return nil
}

// --- HTML ---

type htmlWriter struct{}

func (h *htmlWriter) Write(w io.Writer, result *models.Result) error {
	rows := &strings.Builder{}
	for _, f := range result.Findings {
		style := ""
		if f.Status == models.StatusSkipped {
			style = ` style="color:#888;background:#f5f5f5"`
		}
		fmt.Fprintf(rows,
			"<tr%s><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>\n",
			style,
			f.Release.Name, f.Release.Chart,
			f.CurrentVersion, f.LatestVersion,
			f.Status, f.Message,
		)
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>HDC Dependency Report</title></head>
<body>
<h1>HDC Dependency Report</h1>
<table border="1" cellpadding="4" cellspacing="0">
<thead><tr><th>Release</th><th>Chart</th><th>Current</th><th>Latest</th><th>Status</th><th>Message</th></tr></thead>
<tbody>
%s</tbody>
</table>
</body>
</html>`, rows.String())

	if _, err := fmt.Fprint(w, html); err != nil {
		return fmt.Errorf("failed to write html report: %w", err)
	}

	return nil
}
