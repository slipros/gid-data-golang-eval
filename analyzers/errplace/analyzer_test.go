package errplace_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/errplace"
)

func TestDomainAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), errplace.DomainAnalyzer, "domainsvc/...")
}

func TestDALAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), errplace.DALAnalyzer, "dalsvc/...")
}

func TestFileAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), errplace.FileAnalyzer, "errfile/...")
}
