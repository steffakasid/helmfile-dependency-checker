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
