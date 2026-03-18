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
