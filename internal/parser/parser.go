package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/steffenrumpf/hdc/internal/models"
	"gopkg.in/yaml.v3"
)

// FileReader abstracts filesystem read operations.
type FileReader interface {
	ReadFile(path string) ([]byte, error)
	ReadDir(path string) ([]os.DirEntry, error)
	Glob(pattern string) ([]string, error)
}

// Parser parses a helmfile into a models.Helmfile.
type Parser interface {
	Parse(path string) (*models.Helmfile, error)
}

type fileParser struct {
	reader FileReader
}

type osFileReader struct{}

func (osFileReader) ReadFile(path string) ([]byte, error)       { return os.ReadFile(path) }
func (osFileReader) ReadDir(path string) ([]os.DirEntry, error) { return os.ReadDir(path) }
func (osFileReader) Glob(pattern string) ([]string, error)      { return filepath.Glob(pattern) }

// New returns a new Parser backed by the real filesystem.
func New() Parser {
	return &fileParser{reader: osFileReader{}}
}

// NewWithReader returns a Parser using the provided FileReader (for testing).
func NewWithReader(r FileReader) Parser {
	return &fileParser{reader: r}
}

// templateExpr matches Go template actions: {{ ... }}.
var templateExpr = regexp.MustCompile(`\{\{-?\s*.*?-?\}\}`)

// Parse reads a helmfile (yaml or gotmpl) and returns the parsed Helmfile.
// If path points to a directory it is treated as a helmfile.d directory and
// all *.yaml / *.yaml.gotmpl files inside are merged into a single Helmfile.
func (p *fileParser) Parse(path string) (*models.Helmfile, error) {
	entries, err := p.reader.ReadDir(path)
	if err == nil {
		return p.parseDir(path, entries)
	}

	return p.parseFile(path)
}

func (p *fileParser) parseDir(dir string, entries []os.DirEntry) (*models.Helmfile, error) {
	merged := &models.Helmfile{}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		if !isHelmfileExt(name) {
			continue
		}

		hf, err := p.parseFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}

		merged.Repositories = append(merged.Repositories, hf.Repositories...)
		merged.Releases = append(merged.Releases, hf.Releases...)
	}

	return merged, nil
}

func (p *fileParser) parseFile(path string) (*models.Helmfile, error) {
	return p.parseFileWithVisited(path, make(map[string]bool))
}

func (p *fileParser) parseFileWithVisited(path string, visited map[string]bool) (*models.Helmfile, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for %s: %w", path, err)
	}

	if visited[absPath] {
		return nil, fmt.Errorf("circular sub-helmfile reference detected: %s", path)
	}
	visited[absPath] = true

	data, err := p.reader.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read helmfile %s: %w", path, err)
	}

	cleaned := templateExpr.ReplaceAll(data, []byte(""))

	var hf models.Helmfile
	if err := yaml.Unmarshal(cleaned, &hf); err != nil {
		return nil, fmt.Errorf("failed to parse helmfile %s: %w", path, err)
	}

	if err := p.resolveSubHelmfiles(&hf, filepath.Dir(path), visited); err != nil {
		return nil, err
	}

	return &hf, nil
}

// resolveSubHelmfiles normalizes helmfiles: entries, resolves globs,
// parses referenced files, and merges their repos/releases into hf.
// The visited set tracks absolute paths to detect circular references.
func (p *fileParser) resolveSubHelmfiles(hf *models.Helmfile, baseDir string, visited map[string]bool) error {
	if len(hf.Helmfiles) == 0 {
		return nil
	}

	entries, err := normalizeHelmfileEntries(hf.Helmfiles)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		paths, err := p.resolveEntry(entry, baseDir)
		if err != nil {
			return err
		}

		for _, subPath := range paths {
			sub, err := p.parseFileWithVisited(subPath, visited)
			if err != nil {
				return fmt.Errorf("failed to parse sub-helmfile %s: %w", subPath, err)
			}
			hf.Repositories = append(hf.Repositories, sub.Repositories...)
			hf.Releases = append(hf.Releases, sub.Releases...)
		}
	}

	return nil
}

// resolveEntry resolves a single SubHelmfileEntry to concrete file paths.
// String entries (stored in Path) containing glob characters are expanded;
// explicit paths must exist.
func (p *fileParser) resolveEntry(entry models.SubHelmfileEntry, baseDir string) ([]string, error) {
	resolved := filepath.Join(baseDir, entry.Path)

	if containsGlob(entry.Path) {
		matches, err := p.reader.Glob(resolved)
		if err != nil {
			return nil, fmt.Errorf("failed to glob pattern %s: %w", resolved, err)
		}
		// AC-009.6: zero matches is not an error
		return matches, nil
	}

	// Explicit path — must exist. We verify by attempting to read it;
	// parseFile will surface the actual error, so just return the path.
	return []string{resolved}, nil
}

// normalizeHelmfileEntries converts the raw []any from YAML into typed entries.
// String entries become SubHelmfileEntry with Path set to the string value.
// Map entries are re-marshalled into SubHelmfileEntry.
func normalizeHelmfileEntries(raw []any) ([]models.SubHelmfileEntry, error) {
	entries := make([]models.SubHelmfileEntry, 0, len(raw))

	for _, item := range raw {
		switch v := item.(type) {
		case string:
			entries = append(entries, models.SubHelmfileEntry{Path: v})
		case map[string]any:
			b, err := yaml.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal helmfiles entry: %w", err)
			}
			var entry models.SubHelmfileEntry
			if err := yaml.Unmarshal(b, &entry); err != nil {
				return nil, fmt.Errorf("failed to parse helmfiles entry: %w", err)
			}
			entries = append(entries, entry)
		default:
			return nil, fmt.Errorf("unsupported helmfiles entry type: %T", item)
		}
	}

	return entries, nil
}

// containsGlob reports whether s contains glob meta-characters.
func containsGlob(s string) bool {
	for _, c := range s {
		if c == '*' || c == '?' || c == '[' {
			return true
		}
	}
	return false
}

func isHelmfileExt(name string) bool {
	return filepath.Ext(name) == ".yaml" ||
		filepath.Ext(name) == ".gotmpl"
}
