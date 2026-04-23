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
func New(format string, ignoreSkipped bool) (Writer, error) {
	switch strings.ToLower(format) {
	case "json":
		return &jsonWriter{ignoreSkipped: ignoreSkipped}, nil
	case "markdown":
		return &markdownWriter{ignoreSkipped: ignoreSkipped}, nil
	case "html":
		return &htmlWriter{ignoreSkipped: ignoreSkipped}, nil
	default:
		return nil, fmt.Errorf("unsupported report format %q", format)
	}
}

// --- JSON ---

type jsonWriter struct {
	ignoreSkipped bool
}

func (j *jsonWriter) Write(w io.Writer, result *models.Result) error {
	findings := j.filterFindings(result.Findings)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	if err := enc.Encode(findings); err != nil {
		return fmt.Errorf("failed to write json report: %w", err)
	}

	return nil
}

func (j *jsonWriter) filterFindings(findings []models.Finding) []models.Finding {
	if !j.ignoreSkipped {
		return findings
	}

	var filtered []models.Finding
	for _, f := range findings {
		if f.Status != models.StatusSkipped {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// --- Markdown ---

type markdownWriter struct {
	ignoreSkipped bool
}

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

	findings := m.filterFindings(result.Findings)
	for _, f := range findings {
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

	return nil
}

func (m *markdownWriter) filterFindings(findings []models.Finding) []models.Finding {
	if !m.ignoreSkipped {
		return findings
	}

	var filtered []models.Finding
	for _, f := range findings {
		if f.Status != models.StatusSkipped {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// --- HTML ---

type htmlWriter struct {
	ignoreSkipped bool
}

func (h *htmlWriter) Write(w io.Writer, result *models.Result) error {
	findings := h.filterFindings(result.Findings)

	rows := &strings.Builder{}
	for _, f := range findings {
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

func (h *htmlWriter) filterFindings(findings []models.Finding) []models.Finding {
	if !h.ignoreSkipped {
		return findings
	}

	var filtered []models.Finding
	for _, f := range findings {
		if f.Status != models.StatusSkipped {
			filtered = append(filtered, f)
		}
	}
	return filtered
}
