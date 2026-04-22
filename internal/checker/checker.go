package checker

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/steffenrumpf/hdc/internal/models"
	"github.com/steffenrumpf/hdc/internal/repository"
)

// ErrChartNotFound is returned when a chart is absent from the repository index.
var ErrChartNotFound = fmt.Errorf("chart not found in repository index")

// RepositoryClient fetches a Helm repository index.
type RepositoryClient interface {
	FetchIndex(repoURL string) (*repository.Index, error)
	FetchOCITags(ociURL string) (*repository.Index, error)
}

// Checker runs dependency checks against a parsed Helmfile.
type Checker interface {
	Check(hf *models.Helmfile) (*models.Result, error)
}

// Config holds checker-specific settings.
type Config struct {
	MaxAgeMonths       int
	ConcurrentRequests int
	ExcludeCharts      []string
	ExcludeRepos       []string
}

type checker struct {
	client repository.Client
	cfg    Config
}

// New returns a Checker.
func New(client repository.Client, cfg Config) Checker {
	return &checker{client: client, cfg: cfg}
}

// Check evaluates all releases in hf and returns a Result.
func (c *checker) Check(hf *models.Helmfile) (*models.Result, error) {
	repoByName := make(map[string]models.Repository, len(hf.Repositories))
	for _, r := range hf.Repositories {
		repoByName[r.Name] = r
	}

	sem := make(chan struct{}, max(c.cfg.ConcurrentRequests, 1))
	findings := make([]models.Finding, len(hf.Releases))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i, rel := range hf.Releases {
		if c.isExcluded(rel) {
			findings[i] = models.Finding{Release: rel, Status: models.StatusOK, Message: "excluded"}
			continue
		}

		wg.Add(1)
		go func(idx int, r models.Release) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			f := c.checkRelease(r, repoByName)

			mu.Lock()
			findings[idx] = f
			mu.Unlock()
		}(i, rel)
	}

	wg.Wait()
	return &models.Result{Findings: findings}, nil
}

func (c *checker) checkRelease(rel models.Release, repoByName map[string]models.Repository) models.Finding {
	if isLocalChart(rel.Chart) {
		return models.Finding{
			Release: rel,
			Status:  models.StatusSkipped,
			Message: fmt.Sprintf("local chart reference %q skipped", rel.Chart),
		}
	}

	// OCI URL directly in the chart field (e.g. oci://registry/org/chart).
	if IsOCIRepo(rel.Chart) {
		return c.checkOCIRelease(rel, rel.Chart, "")
	}

	repoName, chartName, ok := splitChart(rel.Chart)
	if !ok {
		return models.Finding{
			Release: rel,
			Status:  models.StatusUnreachable,
			Message: fmt.Sprintf("cannot determine repository for chart %q", rel.Chart),
		}
	}

	repo, ok := repoByName[repoName]
	if !ok {
		return models.Finding{
			Release: rel,
			Status:  models.StatusUnreachable,
			Message: fmt.Sprintf("repository %q not declared in helmfile", repoName),
		}
	}

	repoURL := repo.URL
	if c.isRepoExcluded(repoURL) {
		return models.Finding{Release: rel, Status: models.StatusOK, Message: "excluded"}
	}

	// Check if this is an OCI repository (either by URL scheme or oci: true flag)
	if IsOCIRepo(repoURL) || IsOCIFromFlag(repo) {
		// For repositories with oci: true, construct the full OCI URL
		if IsOCIFromFlag(repo) && !IsOCIRepo(repoURL) {
			repoURL = "oci://" + repoURL
		}
		return c.checkOCIRelease(rel, repoURL, chartName)
	}

	return c.checkHTTPRelease(rel, repoName, repoURL, chartName)
}

// checkOCIRelease handles version checking for OCI-based repositories.
func (c *checker) checkOCIRelease(rel models.Release, repoURL, chartName string) models.Finding {
	var ociURL string

	if chartName == "" {
		// Full OCI URL in chart field — use it directly and extract chart name.
		ociURL = strings.TrimRight(repoURL, "/")
		parts := strings.Split(ociURL, "/")
		chartName = parts[len(parts)-1]
	} else {
		ociURL = strings.TrimRight(repoURL, "/") + "/" + chartName
	}

	idx, err := c.client.FetchOCITags(ociURL)
	if err != nil {
		slog.Warn("failed to fetch OCI tags", "url", ociURL, "error", err)
		return models.Finding{
			Release: rel,
			Status:  models.StatusUnreachable,
			Message: fmt.Sprintf("failed to fetch OCI tags from %q: %s", ociURL, err),
		}
	}

	versions, ok := idx.Entries[chartName]
	if !ok || len(versions) == 0 {
		return models.Finding{
			Release: rel,
			Status:  models.StatusUnreachable,
			Message: fmt.Sprintf("chart %q not found in OCI registry %q", chartName, ociURL),
		}
	}

	// Find latest version by semver comparison since OCI tags have no timestamps.
	latest := versions[0]
	for _, v := range versions[1:] {
		if isNewer(v.Version, latest.Version) {
			latest = v
		}
	}

	status := models.StatusOK
	message := "up-to-date"

	if rel.Version != "" && rel.Version != latest.Version {
		if isNewer(latest.Version, rel.Version) {
			status = models.StatusOutdated
			message = fmt.Sprintf("newer version %s available", latest.Version)
		}
	}

	return models.Finding{
		Release:        rel,
		Status:         status,
		CurrentVersion: rel.Version,
		LatestVersion:  latest.Version,
		Message:        message,
	}
}

// checkHTTPRelease handles version checking for HTTP/HTTPS-based repositories.
func (c *checker) checkHTTPRelease(rel models.Release, repoName, repoURL, chartName string) models.Finding {
	idx, err := c.client.FetchIndex(repoURL)
	if err != nil {
		slog.Warn("failed to fetch repository index", "repo", repoName, "error", err)
		return models.Finding{
			Release: rel,
			Status:  models.StatusUnreachable,
			Message: fmt.Sprintf("failed to fetch repository %q: %s", repoName, err),
		}
	}

	latest, err := idx.LatestVersion(chartName)
	if err != nil {
		return models.Finding{
			Release: rel,
			Status:  models.StatusUnreachable,
			Message: fmt.Sprintf("chart %q not found in repository %q", chartName, repoName),
		}
	}

	status := models.StatusOK
	message := "up-to-date"

	if rel.Version != "" && rel.Version != latest.Version {
		if isNewer(latest.Version, rel.Version) {
			status = models.StatusOutdated
			message = fmt.Sprintf("newer version %s available", latest.Version)
		}
	}

	if !latest.Created.IsZero() {
		age := time.Since(latest.Created)
		maxAge := time.Duration(c.cfg.MaxAgeMonths) * 30 * 24 * time.Hour
		if age > maxAge {
			status = models.StatusUnmaintained
			message = fmt.Sprintf("latest release is %.0f months old", age.Hours()/24/30)
		}
	}

	return models.Finding{
		Release:        rel,
		Status:         status,
		CurrentVersion: rel.Version,
		LatestVersion:  latest.Version,
		LastUpdated:    latest.Created.Format(time.RFC3339),
		Message:        message,
	}
}

// splitChart splits "repo/chart" into (repo, chart, true).
func splitChart(chart string) (string, string, bool) {
	parts := strings.SplitN(chart, "/", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

func (c *checker) isExcluded(rel models.Release) bool {
	for _, ex := range c.cfg.ExcludeCharts {
		if ex == rel.Chart || ex == rel.Name {
			return true
		}
	}

	return false
}

func (c *checker) isRepoExcluded(repoURL string) bool {
	for _, ex := range c.cfg.ExcludeRepos {
		if ex == repoURL {
			return true
		}
	}

	return false
}

// IsOCIRepo returns true if the repository URL uses the oci:// scheme.
func IsOCIRepo(repoURL string) bool {
	return strings.HasPrefix(repoURL, "oci://")
}

// IsOCIFromFlag returns true if the repository has oci: true set.
func IsOCIFromFlag(repo models.Repository) bool {
	return repo.OCI
}

// isLocalChart returns true if the chart field is a local filesystem path.
func isLocalChart(chart string) bool {
	return strings.HasPrefix(chart, "./") ||
		strings.HasPrefix(chart, "../") ||
		strings.HasPrefix(chart, "/")
}

// isNewer returns true if candidate is a higher semver than current.
// Falls back to string inequality for non-semver versions.
func isNewer(candidate, current string) bool {
	cv := parseSemver(candidate)
	cur := parseSemver(current)

	if cv == nil || cur == nil {
		return candidate != current
	}

	if cv[0] != cur[0] {
		return cv[0] > cur[0]
	}

	if cv[1] != cur[1] {
		return cv[1] > cur[1]
	}

	return cv[2] > cur[2]
}

// parseSemver parses "vX.Y.Z" or "X.Y.Z" into [major, minor, patch].
func parseSemver(v string) []int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)

	if len(parts) != 3 {
		return nil
	}

	nums := make([]int, 3)
	for i, p := range parts {
		// strip pre-release suffix
		p = strings.SplitN(p, "-", 2)[0]
		n := 0
		for _, ch := range p {
			if ch < '0' || ch > '9' {
				return nil
			}

			n = n*10 + int(ch-'0')
		}

		nums[i] = n
	}

	return nums
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}
