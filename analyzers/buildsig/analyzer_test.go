package buildsig

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer,
		"dalsvc/dal/repository/build",
		"dalsvc/dal/repository",
		"domainsvc/domain/service",
		"dalsvc/client/x/dal/repository/build")
}

// TestAllowResults — settings.allow-results extends the contract for non-SQL
// builders (search-engine DSL): the listed signatures pass, anything else is
// still flagged. The entries deliberately vary in spelling (whitespace, missing
// parentheses around a single result) — normalization must handle it.
func TestAllowResults(t *testing.T) {
	a := NewAnalyzer(Settings{AllowResults: []string{
		"(string, error)",
		"string",
		"[]string",
		"( *omd.SearchParams,error )",
	}})
	analysistest.Run(t, analysistest.TestData(), a, "dslsvc/dal/repository/build")
}

// TestNormalizeResults — the setting is spelled the way the signature is written
// in Go; comparison ignores whitespace, outer parentheses and interface{}/any.
// The strict default (an empty list allows nothing extra) is covered by
// TestAnalyzer: BuildBad returning (string, error) is still flagged there.
func TestNormalizeResults(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "pair with spaces", in: "(string, error)", want: "string,error"},
		{name: "pair without parentheses", in: "string,error", want: "string,error"},
		{name: "inner spaces", in: "( *omd.SearchParams,  error )", want: "*omd.SearchParams,error"},
		{name: "single result", in: "[]string", want: "[]string"},
		{name: "single result in parentheses", in: "(string)", want: "string"},
		{name: "empty interface spelled long", in: "(string, []interface{}, error)", want: "string,[]any,error"},
	}
	for _, tt := range tests { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeResults(tt.in); got != tt.want {
				t.Fatalf("normalizeResults(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
