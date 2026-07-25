package patterns_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/patterns"
)

func TestTimeNow(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), patterns.TimeNowAnalyzer, "timenow")
}

func TestUUIDNil(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), patterns.UUIDNilAnalyzer, "uuidnil")
}

func TestUUIDVersion(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), patterns.UUIDVersionAnalyzer, "uuidversion")
}

// TestUUIDVersionMajorVersion — the same rule on the versioned import path
// (github.com/gofrs/uuid/v5): a major-version suffix is the same library, and
// comparing the bare path would silently make the rule a no-op there.
func TestUUIDVersionMajorVersion(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), patterns.UUIDVersionAnalyzer, "uuidversionv5")
}

func TestNewDeref(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), patterns.NewDerefAnalyzer, "newderef")
}

func TestYoda(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), patterns.YodaAnalyzer, "yoda")
}

func TestQuoteVerb(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), patterns.QuoteVerbAnalyzer, "quoteverb")
}

func TestDeepEqual(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), patterns.DeepEqualAnalyzer, "deepequal")
}
