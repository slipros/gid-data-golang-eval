package importalias

import (
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer runs the default "common" prefix on the testdata/svc module
// (a go.work pulls in testdata/lib as another module):
//   - positive: an alias nothing calls for — a renamed package, the package
//     name itself, a name the path does not suggest, in a _test.go file too;
//   - negative: a collision with another import, a package-level declaration
//     or a local name at the use site; the real name goimports writes;
//   - boundary: a shadow in a function not using the import; pkg/<module>
//     with the GID-240 import and a common prefix on its own package;
//   - non-applicability: blank/dot imports, another module, generated code,
//     a nested module borrowing its parent's model under the common prefix.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, filepath.Join(analysistest.TestData(), "svc"), Analyzer, "./...")
}

// TestAnalyzerSettings — settings.prefix replaces "common" as the marker of a
// borrowed package (testdata/custom).
func TestAnalyzerSettings(t *testing.T) {
	a := NewAnalyzer(Settings{Prefix: "shared"})
	analysistest.Run(t, filepath.Join(analysistest.TestData(), "custom"), a, "./...")
}

// TestNoModule — non-applicability: without module information there is no
// telling the module's own packages apart (testdata/src/nomod, GOPATH mode).
func TestNoModule(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "nomod/...")
}
