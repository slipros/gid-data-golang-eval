package buildsig

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer,
		"dalsvc/dal/repository/build",
		"dalsvc/dal/repository",
		"domainsvc/domain/service",
		"dalsvc/client/x/dal/repository/build")
}
