package codepolicy

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

func TestSQLCode(t *testing.T) {
	for _, query := range []string{
		`SELECT '$1', "$2", $$ $3 $$, $body$ $4 $body$`,
		"SELECT 1 -- $1\n /* $2 /* $3 */ $4 */",
		`SELECT 'escaped ''$1'`,
		`SELECT E'escaped\' $1'`,
		`SELECT $tag1$ $1 $tag1$`,
	} {
		if positional.MatchString(sqlCode(query)) {
			t.Errorf("quoted placeholder leaked from %q", query)
		}
	}
	if !positional.MatchString(sqlCode("SELECT id WHERE id = $12")) {
		t.Fatal("real placeholder was discarded")
	}
}
