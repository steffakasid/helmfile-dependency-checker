package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
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
	FetchOCITags(ociURL string) (*Index, error)
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

// ociTagsResponse represents the JSON response from the OCI Distribution tags/list API.
type ociTagsResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// FetchOCITags fetches the list of tags from an OCI registry and returns an
// Index containing only valid semver versions. The ociURL must use the oci:// scheme.
func (c *repoClient) FetchOCITags(ociURL string) (*Index, error) {
	host, repo, chartName, err := parseOCIURL(ociURL)
	if err != nil {
		return nil, err
	}

	tagsURL := fmt.Sprintf("https://%s/v2/%s/tags/list", host, repo)
	slog.Debug("fetching OCI tags", "url", tagsURL)

	resp, err := c.http.Get(tagsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OCI tags from %s: %w", ociURL, err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OCI registry %s returned status %d", ociURL, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read OCI tags response from %s: %w", ociURL, err)
	}

	var tagsResp ociTagsResponse
	if err := json.Unmarshal(body, &tagsResp); err != nil {
		return nil, fmt.Errorf("failed to parse OCI tags response from %s: %w", ociURL, err)
	}

	versions := filterSemverTags(tagsResp.Tags)
	if len(versions) == 0 {
		return nil, fmt.Errorf("no valid semver tags found for %s", ociURL)
	}

	idx := &Index{
		Entries: map[string][]ChartVersion{
			chartName: versions,
		},
	}

	return idx, nil
}

// isValidSemver returns true if the tag is a valid semantic version (X.Y.Z with optional v prefix and pre-release suffix).
func isValidSemver(tag string) bool {
	v := strings.TrimPrefix(tag, "v")
	parts := strings.SplitN(v, ".", 3)

	if len(parts) != 3 {
		return false
	}

	for _, p := range parts {
		// strip pre-release suffix for the patch part
		p = strings.SplitN(p, "-", 2)[0]
		if p == "" {
			return false
		}

		for _, ch := range p {
			if ch < '0' || ch > '9' {
				return false
			}
		}
	}

	return true
}

// filterSemverTags filters a list of tags to only valid semver versions and
// returns them as ChartVersion entries with zero Created time (OCI tags/list
// does not provide timestamps).
func filterSemverTags(tags []string) []ChartVersion {
	var versions []ChartVersion

	for _, tag := range tags {
		if isValidSemver(tag) {
			versions = append(versions, ChartVersion{Version: tag})
		}
	}

	return versions
}

// parseOCIURL extracts host, repo path, and chart name from an oci:// URL.
// Example: oci://registry.example.com/charts/mychart
//
//	→ host: registry.example.com, repo: charts/mychart, chart: mychart
func parseOCIURL(ociURL string) (host, repo, chartName string, err error) {
	if !strings.HasPrefix(ociURL, "oci://") {
		return "", "", "", fmt.Errorf("invalid OCI URL %q: must start with oci://", ociURL)
	}

	// Replace oci:// with https:// so net/url can parse it.
	u, err := url.Parse(strings.Replace(ociURL, "oci://", "https://", 1))
	if err != nil {
		return "", "", "", fmt.Errorf("invalid OCI URL %q: %w", ociURL, err)
	}

	host = u.Host
	if host == "" {
		return "", "", "", fmt.Errorf("invalid OCI URL %q: missing host", ociURL)
	}

	trimmed := strings.Trim(u.Path, "/")
	if trimmed == "" {
		return "", "", "", fmt.Errorf("invalid OCI URL %q: missing repository path", ociURL)
	}

	repo = trimmed
	chartName = path.Base(trimmed)

	return host, repo, chartName, nil
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
