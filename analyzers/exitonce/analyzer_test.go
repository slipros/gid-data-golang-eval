package exitonce_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/exitonce"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), exitonce.Analyzer,
		"mainhelper",  // позитив (os.Exit в хелпере main-пакета) + негатив (один os.Exit в main, возврат error)
		"libpkg",      // позитив (log.Fatal/logrus.Fatal* в не-main пакете) + негатив (возврат error)
		"twoexit",     // позитив (два os.Exit в main — повторный)
		"okmain",      // граничный (defer + один os.Exit — ок)
		"closuremain", // граничный (os.Exit в замыкании в main — считается вызовом в main)
		"cleanlib",    // неприменимость (библиотечный пакет без exit-вызовов)
	)
}
