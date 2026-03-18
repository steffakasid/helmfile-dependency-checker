package parser_test

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/steffenrumpf/hdc/internal/parser"
	"github.com/steffenrumpf/hdc/internal/parser/mocks"
)

const yamlContent = `
repositories:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami
releases:
  - name: redis
    namespace: cache
    chart: bitnami/redis
    version: 16.13.0
`

const gotmplContent = `
repositories:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami
releases:
  - name: redis
    namespace: {{ .Values.namespace | default "cache" }}
    chart: bitnami/redis
    version: {{ .Values.redisVersion | default "16.13.0" }}
`

func TestParse_YAML(t *testing.T) {
	reader := mocks.NewMockFileReader(t)
	reader.EXPECT().ReadDir("helmfile.yaml").Return(nil, errors.New("not a dir"))
	reader.EXPECT().ReadFile("helmfile.yaml").Return([]byte(yamlContent), nil)

	hf, err := parser.NewWithReader(reader).Parse("helmfile.yaml")
	require.NoError(t, err)

	assert.Len(t, hf.Repositories, 1)
	assert.Equal(t, "bitnami", hf.Repositories[0].Name)
	assert.Len(t, hf.Releases, 1)
	assert.Equal(t, "redis", hf.Releases[0].Name)
	assert.Equal(t, "16.13.0", hf.Releases[0].Version)
}

func TestParse_GoTmpl(t *testing.T) {
	reader := mocks.NewMockFileReader(t)
	reader.EXPECT().ReadDir("helmfile.yaml.gotmpl").Return(nil, errors.New("not a dir"))
	reader.EXPECT().ReadFile("helmfile.yaml.gotmpl").Return([]byte(gotmplContent), nil)

	hf, err := parser.NewWithReader(reader).Parse("helmfile.yaml.gotmpl")
	require.NoError(t, err)

	assert.Len(t, hf.Releases, 1)
	assert.Equal(t, "redis", hf.Releases[0].Name)
	assert.Equal(t, "bitnami/redis", hf.Releases[0].Chart)
	assert.Empty(t, hf.Releases[0].Version)
	assert.Empty(t, hf.Releases[0].Namespace)
}

func TestParse_Directory(t *testing.T) {
	entries := []os.DirEntry{
		newFakeDirEntry("app.yaml", false),
		newFakeDirEntry("subdir", true),
		newFakeDirEntry("ignored.txt", false),
	}

	reader := mocks.NewMockFileReader(t)
	reader.EXPECT().ReadDir("helmfile.d").Return(entries, nil)
	reader.EXPECT().ReadFile("helmfile.d/app.yaml").Return([]byte(yamlContent), nil)

	hf, err := parser.NewWithReader(reader).Parse("helmfile.d")
	require.NoError(t, err)

	assert.Len(t, hf.Releases, 1)
	assert.Equal(t, "redis", hf.Releases[0].Name)
}

func TestParse_GlobResolvesAndMergesSubHelmfiles(t *testing.T) {
	parentYAML := `
repositories:
  - name: stable
    url: https://charts.helm.sh/stable
releases:
  - name: nginx
    namespace: web
    chart: stable/nginx
    version: 1.0.0
helmfiles:
  - "subdir/*.yaml"
`
	subYAML := `
repositories:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami
releases:
  - name: redis
    namespace: cache
    chart: bitnami/redis
    version: 16.13.0
  - name: postgresql
    namespace: db
    chart: bitnami/postgresql
    version: 12.1.0
`

	reader := mocks.NewMockFileReader(t)
	// Parent file is not a directory
	reader.EXPECT().ReadDir("helmfile.yaml").Return(nil, errors.New("not a dir"))
	reader.EXPECT().ReadFile("helmfile.yaml").Return([]byte(parentYAML), nil)
	// Glob resolves to one sub-helmfile
	reader.EXPECT().Glob("subdir/*.yaml").Return([]string{"subdir/apps.yaml"}, nil)
	// Sub-helmfile content (parseFileWithVisited reads directly, no ReadDir)
	reader.EXPECT().ReadFile("subdir/apps.yaml").Return([]byte(subYAML), nil)

	hf, err := parser.NewWithReader(reader).Parse("helmfile.yaml")
	require.NoError(t, err)

	// Parent repo + sub-helmfile repo
	assert.Len(t, hf.Repositories, 2)
	assert.Equal(t, "stable", hf.Repositories[0].Name)
	assert.Equal(t, "bitnami", hf.Repositories[1].Name)

	// Parent release + 2 sub-helmfile releases
	assert.Len(t, hf.Releases, 3)
	assert.Equal(t, "nginx", hf.Releases[0].Name)
	assert.Equal(t, "redis", hf.Releases[1].Name)
	assert.Equal(t, "postgresql", hf.Releases[2].Name)
}

func TestParse_ExplicitPathWithSelectorsIgnored(t *testing.T) {
	parentYAML := `
repositories:
  - name: stable
    url: https://charts.helm.sh/stable
releases:
  - name: nginx
    namespace: web
    chart: stable/nginx
    version: 1.0.0
helmfiles:
  - path: envs/production.yaml
    selectors:
      - name=prometheus
      - name=grafana
    selectorsInherited: true
`
	subYAML := `
repositories:
  - name: prometheus
    url: https://prometheus-community.github.io/helm-charts
releases:
  - name: prometheus
    namespace: monitoring
    chart: prometheus/prometheus
    version: 25.0.0
  - name: grafana
    namespace: monitoring
    chart: prometheus/grafana
    version: 7.0.0
`

	reader := mocks.NewMockFileReader(t)
	reader.EXPECT().ReadDir("helmfile.yaml").Return(nil, errors.New("not a dir"))
	reader.EXPECT().ReadFile("helmfile.yaml").Return([]byte(parentYAML), nil)
	// Explicit path: no glob, parser reads the file directly
	reader.EXPECT().ReadFile("envs/production.yaml").Return([]byte(subYAML), nil)

	hf, err := parser.NewWithReader(reader).Parse("helmfile.yaml")
	require.NoError(t, err)

	// Selectors are ignored — all repos and releases from sub-helmfile are merged
	assert.Len(t, hf.Repositories, 2)
	assert.Equal(t, "stable", hf.Repositories[0].Name)
	assert.Equal(t, "prometheus", hf.Repositories[1].Name)

	assert.Len(t, hf.Releases, 3)
	assert.Equal(t, "nginx", hf.Releases[0].Name)
	assert.Equal(t, "prometheus", hf.Releases[1].Name)
	assert.Equal(t, "grafana", hf.Releases[2].Name)
}

func TestParse_MissingExplicitPathReturnsError(t *testing.T) {
	parentYAML := `
repositories:
  - name: stable
    url: https://charts.helm.sh/stable
releases:
  - name: nginx
    namespace: web
    chart: stable/nginx
    version: 1.0.0
helmfiles:
  - path: envs/missing.yaml
`

	reader := mocks.NewMockFileReader(t)
	reader.EXPECT().ReadDir("helmfile.yaml").Return(nil, errors.New("not a dir"))
	reader.EXPECT().ReadFile("helmfile.yaml").Return([]byte(parentYAML), nil)
	reader.EXPECT().ReadFile("envs/missing.yaml").Return(nil, os.ErrNotExist)

	_, err := parser.NewWithReader(reader).Parse("helmfile.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "envs/missing.yaml")
	assert.Contains(t, err.Error(), "failed to parse sub-helmfile")
}

func TestParse_GlobMatchingZeroFilesContinuesWithoutError(t *testing.T) {
	parentYAML := `
repositories:
  - name: stable
    url: https://charts.helm.sh/stable
releases:
  - name: nginx
    namespace: web
    chart: stable/nginx
    version: 1.0.0
helmfiles:
  - "nonexistent/*.yaml"
`

	reader := mocks.NewMockFileReader(t)
	reader.EXPECT().ReadDir("helmfile.yaml").Return(nil, errors.New("not a dir"))
	reader.EXPECT().ReadFile("helmfile.yaml").Return([]byte(parentYAML), nil)
	// Glob returns zero matches
	reader.EXPECT().Glob("nonexistent/*.yaml").Return([]string{}, nil)

	hf, err := parser.NewWithReader(reader).Parse("helmfile.yaml")
	require.NoError(t, err)

	// Only parent repos and releases are present — no error from empty glob
	assert.Len(t, hf.Repositories, 1)
	assert.Equal(t, "stable", hf.Repositories[0].Name)
	assert.Len(t, hf.Releases, 1)
	assert.Equal(t, "nginx", hf.Releases[0].Name)
}

func TestParse_RecursiveSubHelmfileReferencesAreFollowed(t *testing.T) {
	rootYAML := `
repositories:
  - name: stable
    url: https://charts.helm.sh/stable
releases:
  - name: nginx
    namespace: web
    chart: stable/nginx
    version: 1.0.0
helmfiles:
  - path: child/helmfile.yaml
`
	childYAML := `
repositories:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami
releases:
  - name: redis
    namespace: cache
    chart: bitnami/redis
    version: 16.13.0
helmfiles:
  - path: grandchild/helmfile.yaml
`
	grandchildYAML := `
repositories:
  - name: grafana
    url: https://grafana.github.io/helm-charts
releases:
  - name: grafana
    namespace: monitoring
    chart: grafana/grafana
    version: 7.0.0
`

	reader := mocks.NewMockFileReader(t)
	reader.EXPECT().ReadDir("helmfile.yaml").Return(nil, errors.New("not a dir"))
	reader.EXPECT().ReadFile("helmfile.yaml").Return([]byte(rootYAML), nil)
	reader.EXPECT().ReadFile("child/helmfile.yaml").Return([]byte(childYAML), nil)
	reader.EXPECT().ReadFile("child/grandchild/helmfile.yaml").Return([]byte(grandchildYAML), nil)

	hf, err := parser.NewWithReader(reader).Parse("helmfile.yaml")
	require.NoError(t, err)

	// All three levels of repositories merged
	assert.Len(t, hf.Repositories, 3)
	assert.Equal(t, "stable", hf.Repositories[0].Name)
	assert.Equal(t, "bitnami", hf.Repositories[1].Name)
	assert.Equal(t, "grafana", hf.Repositories[2].Name)

	// All three levels of releases merged
	assert.Len(t, hf.Releases, 3)
	assert.Equal(t, "nginx", hf.Releases[0].Name)
	assert.Equal(t, "redis", hf.Releases[1].Name)
	assert.Equal(t, "grafana", hf.Releases[2].Name)
}

func TestParse_MixedStringAndMapEntriesInHelmfilesList(t *testing.T) {
	parentYAML := `
repositories:
  - name: stable
    url: https://charts.helm.sh/stable
releases:
  - name: nginx
    namespace: web
    chart: stable/nginx
    version: 1.0.0
helmfiles:
  - "subdir/*.yaml"
  - path: envs/production.yaml
    selectors:
      - name=prometheus
`
	globSubYAML := `
repositories:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami
releases:
  - name: redis
    namespace: cache
    chart: bitnami/redis
    version: 16.13.0
`
	explicitSubYAML := `
repositories:
  - name: prometheus
    url: https://prometheus-community.github.io/helm-charts
releases:
  - name: prometheus
    namespace: monitoring
    chart: prometheus/prometheus
    version: 25.0.0
`

	reader := mocks.NewMockFileReader(t)
	reader.EXPECT().ReadDir("helmfile.yaml").Return(nil, errors.New("not a dir"))
	reader.EXPECT().ReadFile("helmfile.yaml").Return([]byte(parentYAML), nil)
	// String entry resolved via glob
	reader.EXPECT().Glob("subdir/*.yaml").Return([]string{"subdir/apps.yaml"}, nil)
	reader.EXPECT().ReadFile("subdir/apps.yaml").Return([]byte(globSubYAML), nil)
	// Map entry resolved via explicit path
	reader.EXPECT().ReadFile("envs/production.yaml").Return([]byte(explicitSubYAML), nil)

	hf, err := parser.NewWithReader(reader).Parse("helmfile.yaml")
	require.NoError(t, err)

	// Parent + glob sub + explicit sub repositories
	assert.Len(t, hf.Repositories, 3)
	assert.Equal(t, "stable", hf.Repositories[0].Name)
	assert.Equal(t, "bitnami", hf.Repositories[1].Name)
	assert.Equal(t, "prometheus", hf.Repositories[2].Name)

	// Parent + glob sub + explicit sub releases
	assert.Len(t, hf.Releases, 3)
	assert.Equal(t, "nginx", hf.Releases[0].Name)
	assert.Equal(t, "redis", hf.Releases[1].Name)
	assert.Equal(t, "prometheus", hf.Releases[2].Name)
}

func TestParse_ReadError(t *testing.T) {
	reader := mocks.NewMockFileReader(t)
	reader.EXPECT().ReadDir("missing.yaml").Return(nil, errors.New("not a dir"))
	reader.EXPECT().ReadFile("missing.yaml").Return(nil, errors.New("file not found"))

	_, err := parser.NewWithReader(reader).Parse("missing.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read helmfile")
}

// fakeDirEntry is a minimal os.DirEntry implementation for testing.
type fakeDirEntry struct {
	name  string
	isDir bool
}

func newFakeDirEntry(name string, isDir bool) os.DirEntry {
	return fakeDirEntry{name: name, isDir: isDir}
}

func (f fakeDirEntry) Name() string               { return f.name }
func (f fakeDirEntry) IsDir() bool                { return f.isDir }
func (f fakeDirEntry) Type() os.FileMode          { return 0 }
func (f fakeDirEntry) Info() (os.FileInfo, error) { return nil, nil }

func TestParse_Integration_SubHelmfiles(t *testing.T) {
	p := parser.New()
	hf, err := p.Parse("../../testdata/helmfiles/helmfile-integration.yaml")
	require.NoError(t, err)

	// Parent defines 4 repositories; sub-helmfiles define none.
	assert.Len(t, hf.Repositories, 4)

	repoNames := make([]string, len(hf.Repositories))
	for i, r := range hf.Repositories {
		repoNames[i] = r.Name
	}
	assert.Contains(t, repoNames, "bitnami")
	assert.Contains(t, repoNames, "ingress-nginx")
	assert.Contains(t, repoNames, "prometheus-community")
	assert.Contains(t, repoNames, "grafana")

	// Glob apps/*/helmfile.yaml resolves to 3 sub-helmfiles:
	//   databases (postgresql, redis)
	//   ingress   (nginx-ingress, nginx-internal)
	//   monitoring (kube-prometheus-stack, grafana)
	assert.Len(t, hf.Releases, 6)

	releaseNames := make(map[string]int)
	for _, r := range hf.Releases {
		releaseNames[r.Name]++
	}
	assert.Equal(t, 1, releaseNames["postgresql"])
	assert.Equal(t, 1, releaseNames["redis"])
	assert.Equal(t, 1, releaseNames["nginx-ingress"])
	assert.Equal(t, 1, releaseNames["nginx-internal"])
	assert.Equal(t, 1, releaseNames["kube-prometheus-stack"])
	assert.Equal(t, 1, releaseNames["grafana"])

	// Verify specific release details.
	for _, r := range hf.Releases {
		if r.Name == "postgresql" {
			assert.Equal(t, "database", r.Namespace)
			assert.Equal(t, "bitnami/postgresql", r.Chart)
			assert.Equal(t, "11.6.0", r.Version)
		}
	}
}

func TestParse_MultiDocumentYAML(t *testing.T) {
	multiDocContent := `---
environments:
  local: {}
  dev: {}
  prod: {}
---
missingFileHandler: Error

repositories:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami

helmfiles:
  - path: "releases/oauth.yaml"
  - path: "releases/logging.yaml"
`
	oauthYAML := `
releases:
  - name: oauth-proxy
    namespace: auth
    chart: bitnami/oauth2-proxy
    version: 4.0.0
`
	loggingYAML := `
releases:
  - name: fluentbit
    namespace: logging
    chart: bitnami/fluent-bit
    version: 1.0.0
`

	reader := mocks.NewMockFileReader(t)
	reader.EXPECT().ReadDir("helmfile.yaml").Return(nil, errors.New("not a dir"))
	reader.EXPECT().ReadFile("helmfile.yaml").Return([]byte(multiDocContent), nil)
	reader.EXPECT().ReadFile("releases/oauth.yaml").Return([]byte(oauthYAML), nil)
	reader.EXPECT().ReadFile("releases/logging.yaml").Return([]byte(loggingYAML), nil)

	hf, err := parser.NewWithReader(reader).Parse("helmfile.yaml")
	require.NoError(t, err)

	assert.Len(t, hf.Repositories, 1)
	assert.Equal(t, "bitnami", hf.Repositories[0].Name)

	assert.Len(t, hf.Releases, 2)
	assert.Equal(t, "oauth-proxy", hf.Releases[0].Name)
	assert.Equal(t, "fluentbit", hf.Releases[1].Name)
}

func TestParse_Integration_MultiDocumentYAML(t *testing.T) {
	p := parser.New()
	hf, err := p.Parse("../../testdata/helmfiles/helmfile-multidoc.yaml")
	require.NoError(t, err)

	// The multi-doc file references databases and ingress sub-helmfiles
	assert.Len(t, hf.Releases, 4)

	releaseNames := make(map[string]bool)
	for _, r := range hf.Releases {
		releaseNames[r.Name] = true
	}
	assert.True(t, releaseNames["postgresql"])
	assert.True(t, releaseNames["redis"])
	assert.True(t, releaseNames["nginx-ingress"])
	assert.True(t, releaseNames["nginx-internal"])
}

func TestParse_Integration_OverlappingEntriesDetectedAsCircular(t *testing.T) {
	// The original testdata helmfile has a glob AND explicit paths that
	// reference the same files. The visited-path cycle detection treats
	// re-visiting an already-parsed file as a circular reference.
	p := parser.New()
	_, err := p.Parse("../../testdata/helmfiles/helmfile.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular sub-helmfile reference detected")
}
