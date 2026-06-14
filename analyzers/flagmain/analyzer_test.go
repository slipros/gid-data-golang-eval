package flagmain_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/flagmain"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), flagmain.Analyzer,
		"libflag",     // positive: flag.String in a library package
		"libparse",    // positive: flag.Parse in a library
		"mainbad",     // positive: flag names not in snake_case (maxRetries, max-retries)
		"maingood",    // negative: a snake_case name + flag.Parse in main
		"ownflag",     // boundary: a local package named flag — not matched
		"dynamicname", // boundary: the name is not a constant (main — part 2 skips; in a library part 1 catches)
		"testskip",    // boundary: *_test.go with flag is skipped; the regular file is clean
		"cleanlib",    // inapplicability: a library package without flag
	)
}
