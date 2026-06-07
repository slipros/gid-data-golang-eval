package ruleguard_test

import (
	"testing"

	"github.com/quasilyte/go-ruleguard/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

// TestRules — eval ruleguard-правил: тот же analysistest-формат (// want),
// что и у go/analysis-анализаторов, поверх нашего rules.go.
func TestRules(t *testing.T) {
	if err := analyzer.Analyzer.Flags.Set("rules", "rules.go"); err != nil {
		t.Fatalf("set rules flag: %v", err)
	}
	analysistest.Run(t, analysistest.TestData(), analyzer.Analyzer,
		"notimenow", "nouuidcompare", "uuidonlyv7",
		"newderef", "yodaconditions", "quoteverb", "nodeepequal")
}
