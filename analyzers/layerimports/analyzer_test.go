package layerimports_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/layerimports"
)

// TestAnalyzer прогоняет встроенную матрицу на testdata/src/svc/...:
//   - GID-132: dal -> domain, domain/model|usecase -> dal, service -> dal/repository;
//   - GID-170: domain|dal -> event;
//   - GID-172: client -> dal;
//   - GID-224: транспорт (server/schedule/validate/event) видит только domain/model;
//   - GID-225: app и транспорт-листья никем не импортируются;
//   - GID-226: metric самостоятелен, domain/dal не импортируют metric;
//   - GID-227: domain/model — чистый словарь;
//   - GID-228: domain/dal не импортируют client;
//   - GID-229: client изолирован от слоёв сервиса.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), layerimports.Analyzer, "svc/...")
}

// TestAnalyzerSettings — settings.disable выключает встроенное правило,
// settings.rules добавляет своё (testdata/src/custom/...).
func TestAnalyzerSettings(t *testing.T) {
	a := layerimports.NewAnalyzer(layerimports.Settings{
		Disable: []string{"GID-224"},
		Rules: []layerimports.RuleSetting{{
			ID:     "SVC-1",
			Scope:  "domain/service",
			Banned: []string{"legacy"},
			Reason: "пакет legacy выпиливается",
		}},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
