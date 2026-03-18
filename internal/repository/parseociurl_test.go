package repository_test

import (
	"testing"

	"github.com/steffenrumpf/hdc/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOCIURL(t *testing.T) {
	tests := []struct {
		name      string
		ociURL    string
		wantHost  string
		wantRepo  string
		wantChart string
	}{
		{
			name:      "simple registry with chart",
			ociURL:    "oci://registry.example.com/charts/mychart",
			wantHost:  "registry.example.com",
			wantRepo:  "charts/mychart",
			wantChart: "mychart",
		},
		{
			name:      "deeply nested path",
			ociURL:    "oci://registry.example.com/org/team/charts/nginx",
			wantHost:  "registry.example.com",
			wantRepo:  "org/team/charts/nginx",
			wantChart: "nginx",
		},
		{
			name:      "single path segment",
			ociURL:    "oci://registry.example.com/mychart",
			wantHost:  "registry.example.com",
			wantRepo:  "mychart",
			wantChart: "mychart",
		},
		{
			name:      "registry with port",
			ociURL:    "oci://localhost:5000/helm-charts/redis",
			wantHost:  "localhost:5000",
			wantRepo:  "helm-charts/redis",
			wantChart: "redis",
		},
		{
			name:      "trailing slash stripped",
			ociURL:    "oci://registry.example.com/charts/mychart/",
			wantHost:  "registry.example.com",
			wantRepo:  "charts/mychart",
			wantChart: "mychart",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, repo, chartName, err := repository.ParseOCIURL(tt.ociURL)
			require.NoError(t, err)
			assert.Equal(t, tt.wantHost, host)
			assert.Equal(t, tt.wantRepo, repo)
			assert.Equal(t, tt.wantChart, chartName)
		})
	}
}

func TestParseOCIURL_Errors(t *testing.T) {
	tests := []struct {
		name       string
		ociURL     string
		wantErrMsg string
	}{
		{
			name:       "missing oci scheme",
			ociURL:     "https://registry.example.com/charts/mychart",
			wantErrMsg: "must start with oci://",
		},
		{
			name:       "missing host",
			ociURL:     "oci:///charts/mychart",
			wantErrMsg: "missing host",
		},
		{
			name:       "missing repo path",
			ociURL:     "oci://registry.example.com",
			wantErrMsg: "missing repository path",
		},
		{
			name:       "missing repo path with trailing slash",
			ociURL:     "oci://registry.example.com/",
			wantErrMsg: "missing repository path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := repository.ParseOCIURL(tt.ociURL)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrMsg)
		})
	}
}
