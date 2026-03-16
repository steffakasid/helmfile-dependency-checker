package checker_test

import (
	"errors"
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
