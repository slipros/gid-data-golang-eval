package layerimports_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/layerimports"
)

// TestAnalyzer прогоняет все правила пакета на testdata/src/svc/...:
//   - GID-132: dal -> domain, domain/model|usecase -> dal, service -> dal/repository;
//   - GID-170: domain|dal -> event;
//   - GID-172: client -> dal.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), layerimports.Analyzer, "svc/...")
}
