package convpure_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/convpure"
)

// TestAnalyzer runs the default rule set (default settings.packages) on
// testdata/src/svc/...:
//   - positive: convert imports domain/service, domain/usecase, dal/repository,
//     the default banned third-party logrus;
//   - negative: convert imports domain/model, dal/entity, client/*, event/dto, stdlib;
//   - boundary: event/dto exception, "xconvert" is not the exact "convert" segment,
//     "convert/util" ends with "util", not "convert";
//   - non-applicability: an ordinary (non-convert) package.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), convpure.Analyzer, "svc/...")
}

// TestAnalyzerSettings — settings.packages replaces the default third-party
// ban list (testdata/src/custom/...): the custom in-house library is
// flagged, the default logrus is not.
func TestAnalyzerSettings(t *testing.T) {
	a := convpure.NewAnalyzer(convpure.Settings{
		Packages: []string{"example.com/inhouse/somelib"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
