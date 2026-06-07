package upwardimport_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/upwardimport"
)

// TestAnalyzer прогоняет GID-131 на testdata/src/app/...:
//   - позитив: parent/child импортирует parent;
//   - негатив: parent импортирует parent/child; child импортирует соседа;
//   - граничный: parentx НЕ дочерний для parent (префикс по сегментам);
//   - неприменимость: пакет без импортов своего модуля.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), upwardimport.Analyzer, "app/...")
}
