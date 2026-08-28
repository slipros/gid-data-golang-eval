package mappedfields

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestNormalizeFieldPath(t *testing.T) {
	tests := map[string]string{
		"id":           "id",
		"User.Profile": "user.profile",
		"page_size":    "page_size",
		"PAGE_SIZE":    "page_size",
		"URLValue":     "urlValue",
	}

	for input, want := range tests {
		if got := normalizeFieldPath(input); got != want {
			t.Errorf("normalizeFieldPath(%q) = %q, want %q", input, got, want)
		}
	}

	if normalizeFieldPath("page_size") == normalizeFieldPath("pageSize") {
		t.Fatal("snake_case and lowerCamel paths must remain distinct")
	}
}

// TestAnalyzer — the BFF module: positive, negative and boundary cases, plus
// the two non-applicability fixtures living inside it (a _test.go file and the
// transport layer).
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "bff/...")
}

// TestDataLayer — non-applicability: a module owning a data layer is not a BFF.
// It reaches another service through a repository (GID-160), and the fixture
// calls a gRPC client with no mapping and carries no // want.
func TestDataLayer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestLibraryModule — non-applicability: a module that is not laid out as a
// service (no composition root, no data layer) answers to no frontend.
func TestLibraryModule(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "lib/...")
}

// TestExclude — settings.exclude names the RPC being CALLED: a bare method name
// and a "Client.Method" pair. The sibling call outside the list is still
// reported.
func TestExclude(t *testing.T) {
	a := NewAnalyzer(Settings{
		Exclude: []string{"DeleteOrder", "OrderServiceClient.Order"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "bffexclude/...")
}
