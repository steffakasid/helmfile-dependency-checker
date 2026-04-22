package checker_test

import (
	"testing"

	"github.com/steffenrumpf/hdc/internal/checker"
	"github.com/steffenrumpf/hdc/internal/models"
	"pgregory.net/rapid"
)

// Feature: helmfile-dependency-checker, Property 10: Local chart detection
//
// For any string starting with `./`, `../`, or `/`, isLocalChart returns true.
// For any string without these prefixes, isLocalChart returns false.

func TestProperty10_IsLocalChart_WithPathPrefix(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		suffix := rapid.StringMatching(`[a-zA-Z0-9/_\-\.]{1,50}`).Draw(t, "suffix")

		prefix := rapid.SampledFrom([]string{"./", "../", "/"}).Draw(t, "prefix")
		chart := prefix + suffix

		if !checker.IsLocalChart(chart) {
			t.Fatalf("isLocalChart(%q) = false, want true", chart)
		}
	})
}

func TestProperty10_IsLocalChart_WithoutPathPrefix(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate strings that do NOT start with './' , '../', or '/'.
		// Start with an alphanumeric char to guarantee no path prefix.
		chart := rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9/_\-\.]{0,50}`).Draw(t, "chart")

		if checker.IsLocalChart(chart) {
			t.Fatalf("isLocalChart(%q) = true, want false", chart)
		}
	})
}

// Feature: helmfile-dependency-checker, Property 13: OCI repository detection
//
// For any URL with `oci://` prefix, isOCIRepo returns true.
// For any URL with another scheme, isOCIRepo returns false.

func TestProperty13_IsOCIRepo_WithOCIPrefix(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		host := rapid.StringMatching(`[a-z0-9][a-z0-9\-]{0,20}\.[a-z]{2,5}`).Draw(t, "host")
		path := rapid.StringMatching(`[a-z0-9][a-z0-9/\-]{0,30}`).Draw(t, "path")
		url := "oci://" + host + "/" + path

		if !checker.IsOCIRepo(url) {
			t.Fatalf("isOCIRepo(%q) = false, want true", url)
		}
	})
}

func TestProperty13_IsOCIRepo_WithoutOCIPrefix(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		scheme := rapid.SampledFrom([]string{"http://", "https://", "ftp://", "ssh://"}).Draw(t, "scheme")
		host := rapid.StringMatching(`[a-z0-9][a-z0-9\-]{0,20}\.[a-z]{2,5}`).Draw(t, "host")
		path := rapid.StringMatching(`[a-z0-9][a-z0-9/\-]{0,30}`).Draw(t, "path")
		url := scheme + host + "/" + path

		if checker.IsOCIRepo(url) {
			t.Fatalf("isOCIRepo(%q) = true, want false", url)
		}
	})
}

// Feature: helmfile-dependency-checker, Property 15: OCI repository flag detection
//
// For any Repository struct with `OCI: true`, IsOCIFromFlag returns true.
// For any Repository struct with `OCI: false` or unset, IsOCIFromFlag returns false.

func TestProperty15_IsOCIFromFlag_WithFlagSet(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		name := rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9\-]{0,20}`).Draw(t, "name")
		url := rapid.StringMatching(`[a-z0-9][a-z0-9\-]{0,20}\.[a-z]{2,5}[a-z0-9/\-]{0,30}`).Draw(t, "url")

		repo := models.Repository{Name: name, URL: url, OCI: true}

		if !checker.IsOCIFromFlag(repo) {
			t.Fatalf("IsOCIFromFlag(%+v) = false, want true", repo)
		}
	})
}

func TestProperty15_IsOCIFromFlag_WithFlagUnset(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		name := rapid.StringMatching(`[a-zA-Z][a-zA-Z0-9\-]{0,20}`).Draw(t, "name")
		url := rapid.StringMatching(`[a-z0-9][a-z0-9\-]{0,20}\.[a-z]{2,5}[a-z0-9/\-]{0,30}`).Draw(t, "url")
		ociFlag := rapid.SampledFrom([]bool{false}).Draw(t, "ociFlag") // Only false since unset defaults to false

		repo := models.Repository{Name: name, URL: url, OCI: ociFlag}

		if checker.IsOCIFromFlag(repo) {
			t.Fatalf("IsOCIFromFlag(%+v) = true, want false", repo)
		}
	})
}

// Feature: helmfile-dependency-checker, Property 16: OCI repositories with flag use same fetching logic
//
// For any release whose repository has `oci: true` set, the checker should use
// the same OCI fetching and version comparison logic as releases with `oci://` prefixed repository URLs.

func TestProperty16_OCIFromFlag_UsesSameLogic(t *testing.T) {
	// This property is tested indirectly through the integration tests
	// and unit tests that verify both oci:// URLs and oci: true flags
	// result in the same OCI fetching behavior.
	// Since this is a behavioral property that requires mocking,
	// it's covered by the unit tests TestCheck_OCI_Repository_With_Flag
	// and the existing OCI tests.
	t.Skip("Property 16 is covered by unit tests TestCheck_OCI_Repository_With_Flag")
}
