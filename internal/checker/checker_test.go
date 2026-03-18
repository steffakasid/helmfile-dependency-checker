package checker_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/steffenrumpf/hdc/internal/checker"
	"github.com/steffenrumpf/hdc/internal/models"
	"github.com/steffenrumpf/hdc/internal/repository"
	"github.com/steffenrumpf/hdc/internal/repository/mocks"
)

var testHelmfile = &models.Helmfile{
	Repositories: []models.Repository{
		{Name: "bitnami", URL: "https://charts.bitnami.com/bitnami"},
	},
	Releases: []models.Release{
		{Name: "redis", Chart: "bitnami/redis", Version: "16.13.0"},
	},
}

func newIndex(version string, created time.Time) *repository.Index {
	return &repository.Index{
		Entries: map[string][]repository.ChartVersion{
			"redis": {{Version: version, Created: created}},
		},
	}
}

func TestCheck_UpToDate(t *testing.T) {
	client := mocks.NewMockClient(t)
	client.EXPECT().
		FetchIndex("https://charts.bitnami.com/bitnami").
		Return(newIndex("16.13.0", time.Now()), nil)

	chk := checker.New(client, checker.Config{MaxAgeMonths: 12, ConcurrentRequests: 1})
	result, err := chk.Check(testHelmfile)
	require.NoError(t, err)

	require.Len(t, result.Findings, 1)
	assert.Equal(t, models.StatusOK, result.Findings[0].Status)
}

func TestCheck_Outdated(t *testing.T) {
	client := mocks.NewMockClient(t)
	client.EXPECT().
		FetchIndex("https://charts.bitnami.com/bitnami").
		Return(newIndex("17.0.0", time.Now()), nil)

	chk := checker.New(client, checker.Config{MaxAgeMonths: 12, ConcurrentRequests: 1})
	result, err := chk.Check(testHelmfile)
	require.NoError(t, err)

	require.Len(t, result.Findings, 1)
	assert.Equal(t, models.StatusOutdated, result.Findings[0].Status)
	assert.Equal(t, "17.0.0", result.Findings[0].LatestVersion)
}

func TestCheck_Unmaintained(t *testing.T) {
	client := mocks.NewMockClient(t)
	old := time.Now().AddDate(-2, 0, 0)
	client.EXPECT().
		FetchIndex("https://charts.bitnami.com/bitnami").
		Return(newIndex("16.13.0", old), nil)

	chk := checker.New(client, checker.Config{MaxAgeMonths: 12, ConcurrentRequests: 1})
	result, err := chk.Check(testHelmfile)
	require.NoError(t, err)

	require.Len(t, result.Findings, 1)
	assert.Equal(t, models.StatusUnmaintained, result.Findings[0].Status)
}

func TestCheck_Unreachable(t *testing.T) {
	client := mocks.NewMockClient(t)
	client.EXPECT().
		FetchIndex("https://charts.bitnami.com/bitnami").
		Return(nil, errors.New("connection refused"))

	chk := checker.New(client, checker.Config{MaxAgeMonths: 12, ConcurrentRequests: 1})
	result, err := chk.Check(testHelmfile)
	require.NoError(t, err)

	require.Len(t, result.Findings, 1)
	assert.Equal(t, models.StatusUnreachable, result.Findings[0].Status)
}

func TestCheck_ExcludedChart(t *testing.T) {
	client := mocks.NewMockClient(t)
	// No FetchIndex call expected — chart is excluded.

	chk := checker.New(client, checker.Config{
		MaxAgeMonths:       12,
		ConcurrentRequests: 1,
		ExcludeCharts:      []string{"bitnami/redis"},
	})
	result, err := chk.Check(testHelmfile)
	require.NoError(t, err)

	require.Len(t, result.Findings, 1)
	assert.Equal(t, models.StatusOK, result.Findings[0].Status)
}

func TestIsLocalChart_ReturnsTrue(t *testing.T) {
	localPaths := []string{
		"./charts/mychart",
		"../shared/chart",
		"/absolute/path/chart",
	}

	for _, chart := range localPaths {
		t.Run(chart, func(t *testing.T) {
			hf := &models.Helmfile{
				Releases: []models.Release{
					{Name: "local", Chart: chart, Version: "1.0.0"},
				},
			}

			client := mocks.NewMockClient(t)
			// No FetchIndex call expected — local chart is skipped.

			chk := checker.New(client, checker.Config{MaxAgeMonths: 12, ConcurrentRequests: 1})
			result, err := chk.Check(hf)
			require.NoError(t, err)

			require.Len(t, result.Findings, 1)
			assert.Equal(t, models.StatusSkipped, result.Findings[0].Status)
			assert.Contains(t, result.Findings[0].Message, "local chart reference")
		})
	}
}

func TestIsLocalChart_ReturnsFalse(t *testing.T) {
	nonLocalCharts := []struct {
		chart   string
		repoURL string
	}{
		{"bitnami/redis", "https://charts.bitnami.com/bitnami"},
		{"oci://registry/chart", ""},
		{"mychart", ""},
	}

	for _, tc := range nonLocalCharts {
		t.Run(tc.chart, func(t *testing.T) {
			repos := []models.Repository{}
			if tc.repoURL != "" {
				repoName := strings.SplitN(tc.chart, "/", 2)[0]
				repos = append(repos, models.Repository{Name: repoName, URL: tc.repoURL})
			}

			hf := &models.Helmfile{
				Repositories: repos,
				Releases: []models.Release{
					{Name: "test", Chart: tc.chart, Version: "1.0.0"},
				},
			}

			client := mocks.NewMockClient(t)
			if tc.repoURL != "" {
				client.EXPECT().
					FetchIndex(tc.repoURL).
					Return(newIndex("1.0.0", time.Now()), nil).
					Maybe()
			}

			chk := checker.New(client, checker.Config{MaxAgeMonths: 12, ConcurrentRequests: 1})
			result, err := chk.Check(hf)
			require.NoError(t, err)

			require.Len(t, result.Findings, 1)
			assert.NotEqual(t, models.StatusSkipped, result.Findings[0].Status,
				"chart %q should not be treated as a local chart", tc.chart)
		})
	}
}

func TestCheck_SkipsLocalChartWithoutCallingRepoClient(t *testing.T) {
	tests := []struct {
		name  string
		chart string
	}{
		{"relative dot-slash", "./charts/mychart"},
		{"relative parent", "../shared/chart"},
		{"absolute path", "/opt/charts/mychart"},
		{"nested relative", "./deep/nested/chart"},
		{"parent with nested", "../charts/sub/chart"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hf := &models.Helmfile{
				Repositories: []models.Repository{
					{Name: "bitnami", URL: "https://charts.bitnami.com/bitnami"},
				},
				Releases: []models.Release{
					{Name: "local-release", Chart: tt.chart, Version: "1.0.0"},
				},
			}

			client := mocks.NewMockClient(t)
			// No expectations set — any FetchIndex call will fail the test.

			chk := checker.New(client, checker.Config{MaxAgeMonths: 12, ConcurrentRequests: 1})
			result, err := chk.Check(hf)
			require.NoError(t, err)

			require.Len(t, result.Findings, 1)
			f := result.Findings[0]
			assert.Equal(t, models.StatusSkipped, f.Status)
			assert.Contains(t, f.Message, "local chart reference")
			assert.Contains(t, f.Message, tt.chart)

			// Explicitly verify the repository client was never called.
			client.AssertNotCalled(t, "FetchIndex")
		})
	}
}

func TestCheckOCIRelease_UpToDate(t *testing.T) {
	hf := &models.Helmfile{
		Repositories: []models.Repository{
			{Name: "myrepo", URL: "oci://registry.example.com/charts"},
		},
		Releases: []models.Release{
			{Name: "myapp", Chart: "myrepo/myapp", Version: "2.1.0"},
		},
	}

	idx := &repository.Index{
		Entries: map[string][]repository.ChartVersion{
			"myapp": {{Version: "2.1.0"}, {Version: "2.0.0"}},
		},
	}

	client := mocks.NewMockClient(t)
	client.EXPECT().
		FetchOCITags("oci://registry.example.com/charts/myapp").
		Return(idx, nil)

	chk := checker.New(client, checker.Config{MaxAgeMonths: 12, ConcurrentRequests: 1})
	result, err := chk.Check(hf)
	require.NoError(t, err)

	require.Len(t, result.Findings, 1)
	f := result.Findings[0]
	assert.Equal(t, models.StatusOK, f.Status)
	assert.Equal(t, "2.1.0", f.LatestVersion)
	assert.Equal(t, "2.1.0", f.CurrentVersion)
	assert.Equal(t, "up-to-date", f.Message)
	client.AssertNotCalled(t, "FetchIndex")
}

func TestCheckOCIRelease_Outdated(t *testing.T) {
	hf := &models.Helmfile{
		Repositories: []models.Repository{
			{Name: "myrepo", URL: "oci://registry.example.com/charts"},
		},
		Releases: []models.Release{
			{Name: "myapp", Chart: "myrepo/myapp", Version: "1.0.0"},
		},
	}

	idx := &repository.Index{
		Entries: map[string][]repository.ChartVersion{
			"myapp": {{Version: "1.0.0"}, {Version: "2.0.0"}, {Version: "1.5.0"}},
		},
	}

	client := mocks.NewMockClient(t)
	client.EXPECT().
		FetchOCITags("oci://registry.example.com/charts/myapp").
		Return(idx, nil)

	chk := checker.New(client, checker.Config{MaxAgeMonths: 12, ConcurrentRequests: 1})
	result, err := chk.Check(hf)
	require.NoError(t, err)

	require.Len(t, result.Findings, 1)
	f := result.Findings[0]
	assert.Equal(t, models.StatusOutdated, f.Status)
	assert.Equal(t, "2.0.0", f.LatestVersion)
	assert.Equal(t, "1.0.0", f.CurrentVersion)
	assert.Contains(t, f.Message, "newer version")
	client.AssertNotCalled(t, "FetchIndex")
}

func TestCheckOCIRelease_Unreachable(t *testing.T) {
	hf := &models.Helmfile{
		Repositories: []models.Repository{
			{Name: "myrepo", URL: "oci://registry.example.com/charts"},
		},
		Releases: []models.Release{
			{Name: "myapp", Chart: "myrepo/myapp", Version: "1.0.0"},
		},
	}

	client := mocks.NewMockClient(t)
	client.EXPECT().
		FetchOCITags("oci://registry.example.com/charts/myapp").
		Return(nil, errors.New("connection refused"))

	chk := checker.New(client, checker.Config{MaxAgeMonths: 12, ConcurrentRequests: 1})
	result, err := chk.Check(hf)
	require.NoError(t, err)

	require.Len(t, result.Findings, 1)
	f := result.Findings[0]
	assert.Equal(t, models.StatusUnreachable, f.Status)
	assert.Contains(t, f.Message, "oci://registry.example.com/charts/myapp")
	assert.Contains(t, f.Message, "connection refused")
	client.AssertNotCalled(t, "FetchIndex")
}

func TestIsOCIRepo_RoutesToFetchOCITags(t *testing.T) {
	ociURLs := []string{
		"oci://registry.example.com/charts",
		"oci://myregistry.io:5000/org/team",
		"oci://ghcr.io/helm-charts",
	}

	for _, repoURL := range ociURLs {
		t.Run(repoURL, func(t *testing.T) {
			hf := &models.Helmfile{
				Repositories: []models.Repository{
					{Name: "myrepo", URL: repoURL},
				},
				Releases: []models.Release{
					{Name: "myrelease", Chart: "myrepo/mychart", Version: "1.0.0"},
				},
			}

			idx := &repository.Index{
				Entries: map[string][]repository.ChartVersion{
					"mychart": {{Version: "1.0.0"}},
				},
			}

			client := mocks.NewMockClient(t)
			client.EXPECT().
				FetchOCITags(strings.TrimRight(repoURL, "/")+"/mychart").
				Return(idx, nil)
			// No FetchIndex call expected — OCI repos use FetchOCITags.

			chk := checker.New(client, checker.Config{MaxAgeMonths: 12, ConcurrentRequests: 1})
			result, err := chk.Check(hf)
			require.NoError(t, err)

			require.Len(t, result.Findings, 1)
			assert.Equal(t, models.StatusOK, result.Findings[0].Status)
			client.AssertNotCalled(t, "FetchIndex")
		})
	}
}

func TestIsOCIRepo_ReturnsFalseForHTTPRepos(t *testing.T) {
	httpURLs := []string{
		"https://charts.bitnami.com/bitnami",
		"http://charts.example.com/repo",
		"https://registry.example.com/charts",
	}

	for _, repoURL := range httpURLs {
		t.Run(repoURL, func(t *testing.T) {
			hf := &models.Helmfile{
				Repositories: []models.Repository{
					{Name: "myrepo", URL: repoURL},
				},
				Releases: []models.Release{
					{Name: "myrelease", Chart: "myrepo/mychart", Version: "1.0.0"},
				},
			}

			idx := &repository.Index{
				Entries: map[string][]repository.ChartVersion{
					"mychart": {{Version: "1.0.0", Created: time.Now()}},
				},
			}

			client := mocks.NewMockClient(t)
			client.EXPECT().
				FetchIndex(repoURL).
				Return(idx, nil)
			// No FetchOCITags call expected — HTTP repos use FetchIndex.

			chk := checker.New(client, checker.Config{MaxAgeMonths: 12, ConcurrentRequests: 1})
			result, err := chk.Check(hf)
			require.NoError(t, err)

			require.Len(t, result.Findings, 1)
			assert.Equal(t, models.StatusOK, result.Findings[0].Status)
			client.AssertNotCalled(t, "FetchOCITags")
		})
	}
}

func TestResult_HasIssues(t *testing.T) {
	tests := []struct {
		name     string
		findings []models.Finding
		want     bool
	}{
		{"no issues", []models.Finding{{Status: models.StatusOK}}, false},
		{"outdated", []models.Finding{{Status: models.StatusOutdated}}, true},
		{"unmaintained", []models.Finding{{Status: models.StatusUnmaintained}}, true},
		{"unreachable", []models.Finding{{Status: models.StatusUnreachable}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &models.Result{Findings: tt.findings}
			assert.Equal(t, tt.want, r.HasIssues())
		})
	}
}
