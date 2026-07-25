package errplace

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestDomainAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), DomainAnalyzer, "domainsvc/...")
}

func TestDALAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), DALAnalyzer, "dalsvc/...")
}

func TestFileAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), FileAnalyzer, "errfile/...")
}
