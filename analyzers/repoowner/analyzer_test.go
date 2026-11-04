package repoowner

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}
func TestAnalyzerExclude(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(Settings{Exclude: []string{"svc/dal/repository.Campaign"}}), "excluded")
}
