package repository_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/steffenrumpf/hdc/internal/repository"
	"github.com/steffenrumpf/hdc/internal/repository/mocks"
)

const indexYAML = `
apiVersion: v1
entries:
  redis:
    - version: "17.0.0"
      created: "2023-06-01T00:00:00Z"
    - version: "16.13.0"
      created: "2022-01-01T00:00:00Z"
`

func TestFetchIndex_Success(t *testing.T) {
	httpClient := mocks.NewMockHTTPClient(t)
	httpClient.EXPECT().
		Get("https://charts.bitnami.com/bitnami/index.yaml").
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(indexYAML)),
		}, nil)

	client := repository.New(httpClient)
	idx, err := client.FetchIndex("https://charts.bitnami.com/bitnami")
	require.NoError(t, err)

	assert.Contains(t, idx.Entries, "redis")
	assert.Len(t, idx.Entries["redis"], 2)
}

func TestFetchIndex_HTTPError(t *testing.T) {
	httpClient := mocks.NewMockHTTPClient(t)
	httpClient.EXPECT().
		Get("https://charts.bitnami.com/bitnami/index.yaml").
		Return(nil, errors.New("connection refused"))

	client := repository.New(httpClient)
	_, err := client.FetchIndex("https://charts.bitnami.com/bitnami")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch repository index")
}

func TestFetchIndex_Non200(t *testing.T) {
	httpClient := mocks.NewMockHTTPClient(t)
	httpClient.EXPECT().
		Get("https://charts.bitnami.com/bitnami/index.yaml").
		Return(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(bytes.NewBufferString("")),
		}, nil)

	client := repository.New(httpClient)
	_, err := client.FetchIndex("https://charts.bitnami.com/bitnami")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestLatestVersion(t *testing.T) {
	idx, err := repository.ParseIndexYAML([]byte(indexYAML))
	require.NoError(t, err)

	latest, err := idx.LatestVersion("redis")
	require.NoError(t, err)
	assert.Equal(t, "17.0.0", latest.Version)
}

func TestLatestVersion_NotFound(t *testing.T) {
	idx, err := repository.ParseIndexYAML([]byte(indexYAML))
	require.NoError(t, err)

	_, err = idx.LatestVersion("nonexistent")
	require.Error(t, err)
}

func TestFetchOCITags_Success(t *testing.T) {
	tagsJSON := `{"name":"charts/mychart","tags":["1.0.0","2.0.0","1.5.0","latest","dev"]}`

	httpClient := mocks.NewMockHTTPClient(t)
	httpClient.EXPECT().
		Get("https://registry.example.com/v2/charts/mychart/tags/list").
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(tagsJSON)),
		}, nil)

	client := repository.New(httpClient)
	idx, err := client.FetchOCITags("oci://registry.example.com/charts/mychart")
	require.NoError(t, err)

	assert.Contains(t, idx.Entries, "mychart")
	// Only valid semver tags should be included (latest and dev are filtered out)
	assert.Len(t, idx.Entries["mychart"], 3)

	versions := make([]string, 0, len(idx.Entries["mychart"]))
	for _, cv := range idx.Entries["mychart"] {
		versions = append(versions, cv.Version)
	}
	assert.ElementsMatch(t, []string{"1.0.0", "2.0.0", "1.5.0"}, versions)
}

func TestFetchOCITags_ConstructsCorrectURL(t *testing.T) {
	tagsJSON := `{"name":"org/team/nginx","tags":["3.2.1"]}`

	httpClient := mocks.NewMockHTTPClient(t)
	// Verify the exact URL constructed from the OCI reference
	httpClient.EXPECT().
		Get("https://myregistry.io:5000/v2/org/team/nginx/tags/list").
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(tagsJSON)),
		}, nil)

	client := repository.New(httpClient)
	idx, err := client.FetchOCITags("oci://myregistry.io:5000/org/team/nginx")
	require.NoError(t, err)

	assert.Contains(t, idx.Entries, "nginx")
	assert.Len(t, idx.Entries["nginx"], 1)
	assert.Equal(t, "3.2.1", idx.Entries["nginx"][0].Version)
}

func TestFetchOCITags_HTTPError(t *testing.T) {
	httpClient := mocks.NewMockHTTPClient(t)
	httpClient.EXPECT().
		Get("https://registry.example.com/v2/charts/mychart/tags/list").
		Return(nil, errors.New("connection refused"))

	client := repository.New(httpClient)
	_, err := client.FetchOCITags("oci://registry.example.com/charts/mychart")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch OCI tags")
}

func TestFetchOCITags_Non200(t *testing.T) {
	httpClient := mocks.NewMockHTTPClient(t)
	httpClient.EXPECT().
		Get("https://registry.example.com/v2/charts/mychart/tags/list").
		Return(&http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       io.NopCloser(bytes.NewBufferString("")),
		}, nil)

	client := repository.New(httpClient)
	_, err := client.FetchOCITags("oci://registry.example.com/charts/mychart")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestFetchOCITags_InvalidJSON(t *testing.T) {
	httpClient := mocks.NewMockHTTPClient(t)
	httpClient.EXPECT().
		Get("https://registry.example.com/v2/charts/mychart/tags/list").
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString("not json")),
		}, nil)

	client := repository.New(httpClient)
	_, err := client.FetchOCITags("oci://registry.example.com/charts/mychart")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse OCI tags response")
}

func TestFetchOCITags_InvalidOCIURL(t *testing.T) {
	httpClient := mocks.NewMockHTTPClient(t)

	client := repository.New(httpClient)
	_, err := client.FetchOCITags("https://not-oci.example.com/charts/mychart")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must start with oci://")
}

func TestFetchOCITags_FiltersNonSemverTags(t *testing.T) {
	tests := []struct {
		name             string
		tags             string
		expectedVersions []string
	}{
		{
			name:             "filters common non-semver tags",
			tags:             `{"name":"charts/app","tags":["1.0.0","latest","dev","2.3.4","nightly","main"]}`,
			expectedVersions: []string{"1.0.0", "2.3.4"},
		},
		{
			name:             "filters sha-like and date-based tags",
			tags:             `{"name":"charts/app","tags":["0.1.0","abc123","20240101","v1.2.3","sha-deadbeef"]}`,
			expectedVersions: []string{"0.1.0", "v1.2.3"},
		},
		{
			name:             "keeps pre-release semver tags",
			tags:             `{"name":"charts/app","tags":["1.0.0-rc1","2.0.0-beta.1","3.0.0","stable"]}`,
			expectedVersions: []string{"1.0.0-rc1", "2.0.0-beta.1", "3.0.0"},
		},
		{
			name:             "filters single-segment and two-segment versions",
			tags:             `{"name":"charts/app","tags":["1","1.0","1.0.0","v2","v2.1","v2.1.0"]}`,
			expectedVersions: []string{"1.0.0", "v2.1.0"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			httpClient := mocks.NewMockHTTPClient(t)
			httpClient.EXPECT().
				Get("https://registry.example.com/v2/charts/app/tags/list").
				Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(tc.tags)),
				}, nil)

			client := repository.New(httpClient)
			idx, err := client.FetchOCITags("oci://registry.example.com/charts/app")
			require.NoError(t, err)

			require.Contains(t, idx.Entries, "app")
			assert.Len(t, idx.Entries["app"], len(tc.expectedVersions))

			versions := make([]string, 0, len(idx.Entries["app"]))
			for _, cv := range idx.Entries["app"] {
				versions = append(versions, cv.Version)
			}
			assert.ElementsMatch(t, tc.expectedVersions, versions)
		})
	}
}

func TestFetchOCITags_NoValidSemverTags(t *testing.T) {
	tagsJSON := `{"name":"charts/app","tags":["latest","dev","nightly","main"]}`

	httpClient := mocks.NewMockHTTPClient(t)
	httpClient.EXPECT().
		Get("https://registry.example.com/v2/charts/app/tags/list").
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(tagsJSON)),
		}, nil)

	client := repository.New(httpClient)
	_, err := client.FetchOCITags("oci://registry.example.com/charts/app")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid semver tags found")
}
