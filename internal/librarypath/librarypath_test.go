package librarypath

import (
	"go/token"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
)

const (
	analyzerName   = "example"
	originalDoc    = "Use corp.example/lib."
	forkSubpackage = "fork.example/lib/sub"
	sourceLibrary  = "corp.example/lib"
	forkLibrary    = "fork.example/lib"
)

func TestConfiguredPaths(t *testing.T) {
	original := &analysis.Analyzer{
		Name: analyzerName,
		Doc:  originalDoc,
		Run: func(pass *analysis.Pass) (any, error) {
			if got := Canonical(pass, forkSubpackage); got != "corp.example/lib/sub" {
				t.Errorf("canonical = %q", got)
			}

			if got := Canonical(pass, "fork.example/library"); got != "fork.example/library" {
				t.Errorf("prefix boundary = %q", got)
			}

			if got := Replacement(pass, "corp.example/lib/sub"); got != forkSubpackage {
				t.Errorf("replacement = %q", got)
			}

			pass.Reportf(token.Pos(1), "Use corp.example/lib/sub")

			return nil, nil
		},
	}

	configured, err := Configure(original, map[string]string{sourceLibrary: forkLibrary})
	if err != nil {
		t.Fatal(err)
	}

	if len(original.Requires) != 0 || original.Doc != originalDoc {
		t.Fatal("configuration mutated the shared analyzer")
	}

	resolver := configured.Requires[0]
	result, err := resolver.Run(&analysis.Pass{})
	if err != nil {
		t.Fatal(err)
	}

	var messages []string
	pass := &analysis.Pass{
		ResultOf: map[*analysis.Analyzer]any{resolver: result},
		Report: func(diagnostic analysis.Diagnostic) {
			messages = append(messages, diagnostic.Message)
		},
	}

	if _, err := configured.Run(pass); err != nil {
		t.Fatal(err)
	}

	if len(messages) != 1 || messages[0] != "Use fork.example/lib/sub" {
		t.Fatalf("diagnostics = %v", messages)
	}

	if got := Canonical(&analysis.Pass{}, forkLibrary); got != forkLibrary {
		t.Fatalf("unconfigured analyzer inherited mapping: %q", got)
	}
}

func TestInvalidMappings(t *testing.T) {
	for _, libraries := range []map[string]string{
		{"": forkLibrary},
		{sourceLibrary: "invalid path"},
		{"corp.example/one": forkLibrary, "corp.example/two": forkLibrary},
	} {
		if _, err := Configure(&analysis.Analyzer{Name: analyzerName}, libraries); err == nil {
			t.Errorf("accepted invalid libraries %v", libraries)
		}
	}
}

func TestLongestPrefix(t *testing.T) {
	paths := mapping{
		{original: sourceLibrary, fork: forkLibrary},
		{original: "corp.example/special", fork: forkSubpackage},
	}

	pass := &analysis.Pass{ResultOf: map[*analysis.Analyzer]any{nil: paths}}
	if got := Canonical(pass, "fork.example/lib/sub/type"); got != "corp.example/special/type" {
		t.Fatalf("canonical = %q", got)
	}

	if got := paths.rewrite("corp.example/special corp.example/lib"); !strings.Contains(got, forkSubpackage) {
		t.Fatalf("rewritten = %q", got)
	}
}
