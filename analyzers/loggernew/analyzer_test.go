package loggernew

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer,
		"svc/domain/service",
		"svc/dal/repository",
		"svc/cmd/app",
		"svc/internal/app",
		"svc/domain/usecase",
		"svc/domain/handler",
		"svc/domain/slogsvc",
	)
}
