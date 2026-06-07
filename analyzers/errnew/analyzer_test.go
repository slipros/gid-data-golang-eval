package errnew_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/errnew"
)

// TestAnalyzer покрывает позитив, негатив и граничные кейсы.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), errnew.Analyzer, "svc/...")
}

// TestInapplicable — пакет без github.com/pkg/errors не репортится.
func TestInapplicable(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), errnew.Analyzer, "nopkgerrors/...")
}
