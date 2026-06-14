package upwardimport_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/upwardimport"
)

// TestAnalyzer runs GID-131 over testdata/src/app/...:
//   - positive: parent/child imports parent;
//   - negative: parent imports parent/child; child imports a sibling;
//   - boundary: parentx is NOT a child of parent (segment-wise prefix);
//   - not applicable: a package without imports from its own module.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), upwardimport.Analyzer, "app/...")
}
