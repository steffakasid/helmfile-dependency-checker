package models

// Status represents the check result for a single release.
type Status string

const (
	StatusOK           Status = "ok"
	StatusOutdated     Status = "outdated"
	StatusUnmaintained Status = "unmaintained"
	StatusUnreachable  Status = "unreachable"
	StatusSkipped      Status = "skipped"
)

// Finding holds the check result for a single release.
type Finding struct {
	Release        Release
	Status         Status
	CurrentVersion string
	LatestVersion  string
	LastUpdated    string
	Message        string
}

// Result is the aggregated output of a full check run.
type Result struct {
	Findings []Finding
}

// HasIssues returns true when any finding is not StatusOK.
func (r *Result) HasIssues() bool {
	for _, f := range r.Findings {
		if f.Status != StatusOK {
			return true
		}
	}
	return false
}

// HasErrors returns true when any finding has errors (unmaintained or unreachable).
func (r *Result) HasErrors() bool {
	for _, f := range r.Findings {
		if f.Status == StatusUnmaintained || f.Status == StatusUnreachable {
			return true
		}
	}
	return false
}

// HasWarnings returns true when any finding has warnings (outdated) but no errors.
func (r *Result) HasWarnings() bool {
	hasWarnings := false
	for _, f := range r.Findings {
		if f.Status == StatusUnmaintained || f.Status == StatusUnreachable {
			return false // errors take precedence
		}
		if f.Status == StatusOutdated {
			hasWarnings = true
		}
	}
	return hasWarnings
}

// ExitCode returns the appropriate exit code based on finding severity:
// 0 = no issues or only skipped
// 1 = warnings only (outdated charts)
// 2 = errors (unmaintained or unreachable charts).
func (r *Result) ExitCode() int {
	if r.HasErrors() {
		return 2
	}
	if r.HasWarnings() {
		return 1
	}
	return 0
}
