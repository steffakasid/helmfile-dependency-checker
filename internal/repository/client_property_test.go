package repository_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/steffenrumpf/hdc/internal/repository"
	"pgregory.net/rapid"
)

// Feature: helmfile-dependency-checker, Property 12: OCI tag filtering and latest version selection
//
// For any list of OCI tags containing a mix of valid semver strings and
// non-semver strings (e.g. "latest", "dev"), the OCI client should filter
// to only valid semver tags and return the highest version as the latest.

func TestProperty12_FilterSemverTags_OnlyValidSemver(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		numSemver := rapid.IntRange(0, 10).Draw(t, "numSemver")
		numNonSemver := rapid.IntRange(0, 10).Draw(t, "numNonSemver")

		var tags []string
		var expectedCount int

		for i := 0; i < numSemver; i++ {
			major := rapid.IntRange(0, 99).Draw(t, "major")
			minor := rapid.IntRange(0, 99).Draw(t, "minor")
			patch := rapid.IntRange(0, 99).Draw(t, "patch")
			useVPrefix := rapid.Bool().Draw(t, "vPrefix")

			var tag string
			if useVPrefix {
				tag = fmt.Sprintf("v%d.%d.%d", major, minor, patch)
			} else {
				tag = fmt.Sprintf("%d.%d.%d", major, minor, patch)
			}

			tags = append(tags, tag)
			expectedCount++
		}

		nonSemverPatterns := []string{
			"latest", "dev", "nightly", "main", "sha-abc123",
			"rc", "test", "foo.bar", "1.2", "v1", "abc.def.ghi",
		}

		for i := 0; i < numNonSemver; i++ {
			tag := rapid.SampledFrom(nonSemverPatterns).Draw(t, "nonSemverTag")
			tags = append(tags, tag)
		}

		result := repository.FilterSemverTags(tags)

		// All returned entries must be valid semver.
		for _, cv := range result {
			if !repository.IsValidSemver(cv.Version) {
				t.Fatalf("filterSemverTags returned non-semver tag %q", cv.Version)
			}
		}

		// Count must match the number of valid semver tags generated.
		if len(result) != expectedCount {
			t.Fatalf("filterSemverTags returned %d entries, want %d", len(result), expectedCount)
		}
	})
}

func TestProperty12_FilterSemverTags_CorrectLatest(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// At least one valid semver tag so we can determine a latest.
		numSemver := rapid.IntRange(1, 10).Draw(t, "numSemver")
		numNonSemver := rapid.IntRange(0, 10).Draw(t, "numNonSemver")

		type triple struct {
			major, minor, patch int
			tag                 string
		}

		var triples []triple
		var tags []string

		for i := 0; i < numSemver; i++ {
			major := rapid.IntRange(0, 99).Draw(t, "major")
			minor := rapid.IntRange(0, 99).Draw(t, "minor")
			patch := rapid.IntRange(0, 99).Draw(t, "patch")
			useVPrefix := rapid.Bool().Draw(t, "vPrefix")

			var tag string
			if useVPrefix {
				tag = fmt.Sprintf("v%d.%d.%d", major, minor, patch)
			} else {
				tag = fmt.Sprintf("%d.%d.%d", major, minor, patch)
			}

			triples = append(triples, triple{major, minor, patch, tag})
			tags = append(tags, tag)
		}

		nonSemverPatterns := []string{"latest", "dev", "nightly", "main", "sha-abc123"}

		for i := 0; i < numNonSemver; i++ {
			tag := rapid.SampledFrom(nonSemverPatterns).Draw(t, "nonSemverTag")
			tags = append(tags, tag)
		}

		result := repository.FilterSemverTags(tags)
		if len(result) == 0 {
			t.Fatal("filterSemverTags returned empty for input with semver tags")
		}

		// Determine expected latest by sorting triples descending.
		sort.Slice(triples, func(i, j int) bool {
			if triples[i].major != triples[j].major {
				return triples[i].major > triples[j].major
			}

			if triples[i].minor != triples[j].minor {
				return triples[i].minor > triples[j].minor
			}

			return triples[i].patch > triples[j].patch
		})

		wantMajor := triples[0].major
		wantMinor := triples[0].minor
		wantPatch := triples[0].patch

		// Find latest from filtered result using semver comparison.
		latest := result[0]

		for _, cv := range result[1:] {
			if semverGT(cv.Version, latest.Version) {
				latest = cv
			}
		}

		got := parseSemverTriple(latest.Version)
		if got == nil {
			t.Fatalf("could not parse latest version %q", latest.Version)
		}

		if got[0] != wantMajor || got[1] != wantMinor || got[2] != wantPatch {
			t.Fatalf("latest=%q (%v), want %d.%d.%d",
				latest.Version, got, wantMajor, wantMinor, wantPatch)
		}
	})
}

// parseSemverTriple parses "vX.Y.Z" or "X.Y.Z" into [major, minor, patch].
func parseSemverTriple(v string) []int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)

	if len(parts) != 3 {
		return nil
	}

	nums := make([]int, 3)

	for i, p := range parts {
		p = strings.SplitN(p, "-", 2)[0]
		n := 0

		for _, ch := range p {
			if ch < '0' || ch > '9' {
				return nil
			}

			n = n*10 + int(ch-'0')
		}

		nums[i] = n
	}

	return nums
}

// semverGT returns true if a > b by semver ordering.
func semverGT(a, b string) bool {
	av := parseSemverTriple(a)
	bv := parseSemverTriple(b)

	if av == nil || bv == nil {
		return a != b
	}

	if av[0] != bv[0] {
		return av[0] > bv[0]
	}

	if av[1] != bv[1] {
		return av[1] > bv[1]
	}

	return av[2] > bv[2]
}
