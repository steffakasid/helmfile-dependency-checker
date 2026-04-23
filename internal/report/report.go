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

type jsonReport struct {
	Summary struct {
		Total    int `json:"total"`
		OK       int `json:"ok"`
		Warnings int `json:"warnings"`
		Errors   int `json:"errors"`
		Skipped  int `json:"skipped"`
	} `json:"summary"`
	Findings []models.Finding `json:"findings"`
}

func (j *jsonWriter) Write(w io.Writer, result *models.Result) error {
	findings := j.filterFindings(result.Findings)

	report := jsonReport{}
	report.Findings = findings

	// Calculate summary
	for _, f := range result.Findings { // Use original findings for summary, not filtered
		report.Summary.Total++
		switch f.Status {
		case models.StatusOK:
			report.Summary.OK++
		case models.StatusOutdated:
			report.Summary.Warnings++
		case models.StatusUnmaintained, models.StatusUnreachable:
			report.Summary.Errors++
		case models.StatusSkipped:
			report.Summary.Skipped++
		}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	if err := enc.Encode(report); err != nil {
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

	// Add summary section
	if err := m.writeSummary(w, result); err != nil {
		return err
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

func (m *markdownWriter) writeSummary(w io.Writer, result *models.Result) error {
	total := len(result.Findings)
	ok := 0
	warnings := 0
	errors := 0
	skipped := 0

	for _, f := range result.Findings {
		switch f.Status {
		case models.StatusOK:
			ok++
		case models.StatusOutdated:
			warnings++
		case models.StatusUnmaintained, models.StatusUnreachable:
			errors++
		case models.StatusSkipped:
			skipped++
		}
	}

	summary := fmt.Sprintf("**Summary:** %d releases checked", total)
	if ok > 0 {
		summary += fmt.Sprintf(", %d ✅ OK", ok)
	}
	if warnings > 0 {
		summary += fmt.Sprintf(", %d ⚠️ warnings", warnings)
	}
	if errors > 0 {
		summary += fmt.Sprintf(", %d ❌ errors", errors)
	}
	if skipped > 0 && !m.ignoreSkipped {
		summary += fmt.Sprintf(", %d ⏭️ skipped", skipped)
	}

	if _, err := fmt.Fprintf(w, "%s\n\n", summary); err != nil {
		return fmt.Errorf("failed to write markdown summary: %w", err)
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

	// Generate summary
	summaryHTML := h.generateSummaryHTML(result)

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
%s
<table border="1" cellpadding="4" cellspacing="0">
<thead><tr><th>Release</th><th>Chart</th><th>Current</th><th>Latest</th><th>Status</th><th>Message</th></tr></thead>
<tbody>
%s</tbody>
</table>
</body>
</html>`, summaryHTML, rows.String())

	if _, err := fmt.Fprint(w, html); err != nil {
		return fmt.Errorf("failed to write html report: %w", err)
	}

	return nil
}

func (h *htmlWriter) generateSummaryHTML(result *models.Result) string {
	total := len(result.Findings)
	ok := 0
	warnings := 0
	errors := 0
	skipped := 0

	for _, f := range result.Findings {
		switch f.Status {
		case models.StatusOK:
			ok++
		case models.StatusOutdated:
			warnings++
		case models.StatusUnmaintained, models.StatusUnreachable:
			errors++
		case models.StatusSkipped:
			skipped++
		}
	}

	summary := fmt.Sprintf("<p><strong>Summary:</strong> %d releases checked", total)
	if ok > 0 {
		summary += fmt.Sprintf(", %d ✅ OK", ok)
	}
	if warnings > 0 {
		summary += fmt.Sprintf(", %d ⚠️ warnings", warnings)
	}
	if errors > 0 {
		summary += fmt.Sprintf(", %d ❌ errors", errors)
	}
	if skipped > 0 && !h.ignoreSkipped {
		summary += fmt.Sprintf(", %d ⏭️ skipped", skipped)
	}
	summary += "</p>"

	return summary
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
