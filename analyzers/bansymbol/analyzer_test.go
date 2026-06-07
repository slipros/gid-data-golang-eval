package bansymbol_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/bansymbol"
)

// TestAnalyzer — дефолтные настройки: запрет gdpostgres.TQuery.
// Покрывает позитив (вызов и generic-инстанциация), негатив (Select,
// NamedStruct, одноимённый TQuery другого пакета) и неприменимость
// (сгенерированный файл svc/generated.go).
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), bansymbol.Analyzer, "svc")
}

// TestInapplicable — пакет без импорта забаненной библиотеки: чисто.
func TestInapplicable(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), bansymbol.Analyzer, "clean")
}

// TestCustomMsg — граничный: символ задан настройками с кастомным Msg —
// флагается этим Msg.
func TestCustomMsg(t *testing.T) {
	a := bansymbol.NewAnalyzer(bansymbol.Settings{Symbols: []bansymbol.Symbol{
		{
			Pkg:  "example.com/otherdb",
			Name: "TQuery",
			Msg:  "otherdb.TQuery под запретом проекта",
		},
	}})
	analysistest.Run(t, analysistest.TestData(), a, "custom")
}

// TestDefaultMsg — символ без Msg использует общую формулировку; Pkg задан
// суффиксом сегментов пути (postgres.git) — проверяет суффикс-матч.
func TestDefaultMsg(t *testing.T) {
	a := bansymbol.NewAnalyzer(bansymbol.Settings{Symbols: []bansymbol.Symbol{
		{
			Pkg:  "libs/postgres.git",
			Name: "Select",
		},
	}})
	analysistest.Run(t, analysistest.TestData(), a, "nomsg")
}
