package metricstruct_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/metricstruct"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), metricstruct.Analyzer,
		"svc/internal/metric",           // ok: prometheus.go + groups + Register wiring
		"svc/internal/ptrreg/metric",    // ok: pointer-receiver Register
		"missing/internal/metric",       // want: no Prometheus (file picked by name)
		"noregister/internal/metric",    // want: Prometheus without Register (check 3)
		"notstruct/internal/metric",     // want: Prometheus is not a struct (check 4)
		"named/internal/metrics",        // want: path .../metrics (check 1)
		"other/internal/domain/service", // not applicable: outside the metric path
		"wrongfile/internal/metric",     // want: Prometheus outside prometheus.go (check 5)
		"wiringgroup/internal/metric",   // want: foreign group in prometheus.go (check 6)
		"twogroups/internal/metric",     // want: two groups in one file (check 7)
		"notregistered/internal/metric", // want: group field without a Register call (check 8)
		"embedded/internal/metric",      // want: embedded group without a Register call (check 8)
	)
}
