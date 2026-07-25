package patterns

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestTimeNow(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), TimeNowAnalyzer, "timenow")
}

func TestUUIDNil(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), UUIDNilAnalyzer, "uuidnil")
}

func TestUUIDVersion(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), UUIDVersionAnalyzer, "uuidversion")
}

// TestUUIDVersionMajorVersion — the same rule on the versioned import path
// (github.com/gofrs/uuid/v5): a major-version suffix is the same library, and
// comparing the bare path would silently make the rule a no-op there.
func TestUUIDVersionMajorVersion(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), UUIDVersionAnalyzer, "uuidversionv5")
}

func TestNewDeref(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewDerefAnalyzer, "newderef")
}

func TestYoda(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), YodaAnalyzer, "yoda")
}

func TestQuoteVerb(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), QuoteVerbAnalyzer, "quoteverb")
}

func TestDeepEqual(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), DeepEqualAnalyzer, "deepequal")
}
