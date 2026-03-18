package checker_test

import (
	"testing"

	"github.com/steffenrumpf/hdc/internal/checker"
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
