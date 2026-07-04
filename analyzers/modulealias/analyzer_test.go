package modulealias_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/modulealias"
)

// TestAnalyzer runs the default "common" prefix on testdata/src/repo/...:
//   - positive: a shared internal/** import with no alias or with an
//     alias lacking the common prefix, including a dot-import;
//   - negative: a commonservice alias, and a blank import (side-effect-only);
//   - boundary: outside pkg/<module>, GID-240 does not apply.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), modulealias.Analyzer, "repo/...")
}

// TestAnalyzerSettings — settings.prefix replaces the default "common"
// prefix (testdata/src/custom/...): a "commonservice" alias no longer
// satisfies the rule, while "sharedservice" does.
func TestAnalyzerSettings(t *testing.T) {
	a := modulealias.NewAnalyzer(modulealias.Settings{Prefix: "shared"})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
