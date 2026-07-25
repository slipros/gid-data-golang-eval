package bansymbol

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer — default settings: ban on gdpostgres.TQuery.
// Covers the positive case (a call and a generic instantiation), the negative
// case (Select, NamedStruct, a same-named TQuery from another package) and
// non-applicability (the generated file svc/generated.go).
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc")
}

// TestLogrusInterfaces — the other half of the built-in list: the logrus
// logger interfaces (FieldLogger/StdLogger/Ext1FieldLogger) are banned because
// they cannot carry a context, while the concrete *logrus.Entry is fine.
func TestLogrusInterfaces(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "logrusiface")
}

// TestInapplicable — a package without an import of the banned library: clean.
func TestInapplicable(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "clean")
}

// TestCustomMsg — edge case: a symbol configured via settings with a custom Msg —
// it is flagged with that Msg.
func TestCustomMsg(t *testing.T) {
	a := NewAnalyzer(Settings{Symbols: []Symbol{
		{
			Pkg:  "example.com/otherdb",
			Name: "TQuery",
			Msg:  "otherdb.TQuery is banned by the project",
		},
	}})
	analysistest.Run(t, analysistest.TestData(), a, "custom")
}

// TestDefaultMsg — a symbol without Msg uses the generic wording; Pkg is given
// as a suffix of path segments (postgres.git) — verifies the suffix match.
func TestDefaultMsg(t *testing.T) {
	a := NewAnalyzer(Settings{Symbols: []Symbol{
		{
			Pkg:  "libs/postgres.git",
			Name: "Select",
		},
	}})
	analysistest.Run(t, analysistest.TestData(), a, "nomsg")
}
