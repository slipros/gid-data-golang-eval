package httprouter

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer — default settings, against a stand-in of the library sitting at
// its real import path.
//
//   - positive: a bare router (a call and a value) passed to NewServer, an
//     application router nested in the system one, a bare router nested there,
//     and NewApplicationRouter given a nil metrics (GID-265);
//   - negative: the canonical shape — system and application router side by
//     side, metrics passed in;
//   - boundary: NewRouterGroup wraps nothing and is judged as a bare route; a
//     service's own wrapper is recognised by the metrics increment function it
//     is handed;
//   - non-applicability: a service with system routes only, and a same-named
//     NewServer from a foreign package.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "app", "foreign")
}

// TestAnalyzerApplicationRouters — non-applicability: a service's own
// application router that the fingerprint cannot see is named in
// settings.application-routers.
func TestAnalyzerApplicationRouters(t *testing.T) {
	a := NewAnalyzer(Settings{ApplicationRouters: []string{"custom.NewApplication"}})
	analysistest.Run(t, analysistest.TestData(), a, "custom")
}

// TestAnalyzerWithoutSetting — the same package without the setting: the
// wrapper reads as a bare route, which is what makes the setting necessary.
func TestAnalyzerWithoutSetting(t *testing.T) {
	dir := analysistest.TestData()
	results := analysistest.Run(t, dir, Analyzer, "custombare")
	if len(results) == 0 {
		t.Fatal("no package was analysed")
	}
}
