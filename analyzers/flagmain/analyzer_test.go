package flagmain_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/flagmain"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), flagmain.Analyzer,
		"libflag",     // позитив: flag.String в библиотечном пакете
		"libparse",    // позитив: flag.Parse в библиотеке
		"mainbad",     // позитив: имена флагов не snake_case (maxRetries, max-retries)
		"maingood",    // негатив: snake_case имя + flag.Parse в main
		"ownflag",     // граничный: свой пакет с именем flag — не матчится
		"dynamicname", // граничный: имя не константа (main — часть 2 пропуск; в либе часть 1 ловит)
		"testskip",    // граничный: *_test.go с flag пропускается; обычный файл чист
		"cleanlib",    // неприменимость: библиотечный пакет без flag
	)
}
