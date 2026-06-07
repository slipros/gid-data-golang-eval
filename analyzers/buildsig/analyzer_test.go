package buildsig_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/buildsig"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), buildsig.Analyzer,
		"dalsvc/dal/repository/build",
		"dalsvc/dal/repository",
		"domainsvc/domain/service")
}
