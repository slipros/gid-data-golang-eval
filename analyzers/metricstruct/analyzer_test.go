package metricstruct_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/metricstruct"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), metricstruct.Analyzer,
		"svc/internal/metric",           // ok: prometheus.go + группы + Register wiring
		"svc/internal/ptrreg/metric",    // ok: pointer-receiver Register
		"missing/internal/metric",       // want: нет Prometheus (выбор файла по имени)
		"noregister/internal/metric",    // want: Prometheus без Register (проверка 3)
		"notstruct/internal/metric",     // want: Prometheus не struct (проверка 4)
		"named/internal/metrics",        // want: путь .../metrics (проверка 1)
		"other/internal/domain/service", // неприменимость: вне metric-пути
		"wrongfile/internal/metric",     // want: Prometheus не в prometheus.go (проверка 5)
		"wiringgroup/internal/metric",   // want: чужая группа в prometheus.go (проверка 6)
		"twogroups/internal/metric",     // want: две группы в одном файле (проверка 7)
		"notregistered/internal/metric", // want: группа-поле без Register-вызова (проверка 8)
		"embedded/internal/metric",      // want: embedded-группа без Register-вызова (проверка 8)
	)
}
