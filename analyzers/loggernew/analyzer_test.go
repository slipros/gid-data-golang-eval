package loggernew_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/loggernew"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), loggernew.Analyzer,
		"svc/domain/service",
		"svc/dal/repository",
		"svc/cmd/app",
		"svc/internal/app",
		"svc/domain/usecase",
		"svc/domain/handler",
	)
}
