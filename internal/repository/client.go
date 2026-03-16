package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// HTTPClient abstracts HTTP GET operations.
type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

// ChartVersion holds metadata for a single chart version from index.yaml.
type ChartVersion struct {
	Version string    `json:"version"`
	Created time.Time `json:"created"`
}

// Index represents the relevant parts of a Helm repository index.yaml.
type Index struct {
	Entries map[string][]ChartVersion `json:"entries"`
}

// Client fetches and parses Helm repository indexes.
type Client interface {
	FetchIndex(repoURL string) (*Index, error)
}

type repoClient struct {
	http HTTPClient
}

// New returns a Client using the provided HTTPClient.
func New(h HTTPClient) Client {
	return &repoClient{http: h}
}

// NewDefault returns a Client backed by the standard http.Client with the given timeout.
func NewDefault(timeout time.Duration) Client {
	return New(&http.Client{Timeout: timeout})
}

// FetchIndex downloads and parses the index.yaml from a Helm repository.
func (c *repoClient) FetchIndex(repoURL string) (*Index, error) {
	url := strings.TrimRight(repoURL, "/") + "/index.yaml"
	slog.Debug("fetching repository index", "url", url)

	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository index %s: %w", repoURL, err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("repository %s returned status %d", repoURL, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read index from %s: %w", repoURL, err)
	}

	index, err := parseIndex(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse index from %s: %w", repoURL, err)
	}

	return index, nil
}

// parseIndex parses a Helm index.yaml (YAML with JSON-compatible structure).
// We use a two-pass approach: unmarshal via a raw map then re-encode to JSON
// for typed unmarshalling, avoiding a yaml dependency in this package.
func parseIndex(data []byte) (*Index, error) {
	// Helm index.yaml uses YAML — we leverage encoding/json via an intermediate
	// representation decoded with the yaml package from the parent module.
	// To keep this package dependency-free we accept the raw bytes and delegate
	// to the exported ParseIndexYAML helper which lives in index.go.
	return ParseIndexYAML(data)
}

// LatestVersion returns the most recent version entry for a chart name.
func (idx *Index) LatestVersion(chartName string) (*ChartVersion, error) {
	versions, ok := idx.Entries[chartName]
	if !ok || len(versions) == 0 {
		return nil, fmt.Errorf("chart %q not found in repository index", chartName)
	}

	latest := &versions[0]
	for i := range versions[1:] {
		if versions[i+1].Created.After(latest.Created) {
			latest = &versions[i+1]
		}
	}

	return latest, nil
}

// indexJSON is used for JSON-based unmarshalling of the parsed index.
type indexJSON struct {
	Entries map[string][]struct {
		Version string `json:"version"`
		Created string `json:"created"`
	} `json:"entries"`
}

// fromJSON converts an indexJSON into an Index.
func fromJSON(raw indexJSON) (*Index, error) {
	idx := &Index{Entries: make(map[string][]ChartVersion, len(raw.Entries))}

	for chart, versions := range raw.Entries {
		cvs := make([]ChartVersion, 0, len(versions))

		for _, v := range versions {
			t, err := time.Parse(time.RFC3339, v.Created)
			if err != nil {
				// Fall back to zero time if unparseable — chart will be flagged as unmaintained.
				slog.Debug("could not parse chart created timestamp", "chart", chart, "version", v.Version, "value", v.Created)
				t = time.Time{}
			}

			cvs = append(cvs, ChartVersion{Version: v.Version, Created: t})
		}

		idx.Entries[chart] = cvs
	}

	return idx, nil
}

// reEncodeJSON re-encodes a generic map to JSON bytes for typed unmarshalling.
func reEncodeJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}

// unmarshalJSON wraps json.Unmarshal for use across files in this package.
func unmarshalJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
