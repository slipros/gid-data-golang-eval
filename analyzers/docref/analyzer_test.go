package docref

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer — default markers: a document name (ARD/PRD/BACKLOG), a
// requirement id (@ФТ-11), a section (§12), a task number, a traceability
// marker (VERIFY) and a commit reference are all reported, one diagnostic per
// comment. The package also holds svc_test.go: a test file is not judged by
// default — the id paired with the name of the test proving it is the trace
// the requirement map is built from.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc")
}

// TestClean — negative: comments that state the constraint itself, plus the
// text that only looks like a reference (RFC3339, UTF-8, HTTP/2, SkipVerify)
// and the directive comments (//nolint with its justification, //go:generate).
func TestClean(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "clean")
}

// TestGenerated — non-applicability: generated code is skipped whatever its
// comments carry.
func TestGenerated(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "gen")
}

// TestCustomPatterns — boundary: settings.patterns replaces the built-in list
// (so ARD/задача stop being markers) and settings.extra adds the project's
// tracker key on top.
func TestCustomPatterns(t *testing.T) {
	a := NewAnalyzer(Settings{
		Patterns: []string{`(?i)\bwiki\b`},
		Extra:    []string{`\bUDMP-\d+\b`},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom")
}

// TestIncludeTests — boundary: settings.include-tests judges _test.go too, for
// a repository that has already extracted its requirement map into a file.
func TestIncludeTests(t *testing.T) {
	a := NewAnalyzer(Settings{IncludeTests: true})
	analysistest.Run(t, analysistest.TestData(), a, "withtests")
}

// TestFirstMarker — the quoted marker of the diagnostic: the leftmost
// reference of the comment, with the delimiter the pattern had to capture
// trimmed off, and its offset inside the comment text. The analysistest
// fixtures cannot assert this: the marker inside a `// want` comment would be
// matched as a reference of its own.
func TestFirstMarker(t *testing.T) {
	markers, err := compile(defaultPatterns)
	if err != nil {
		t.Fatalf("compile default patterns: %v", err)
	}

	tests := []struct {
		name        string
		text        string
		marker      string
		requirement bool // the finding carries the requirement-map fix
		found       bool
	}{
		{name: "document name", text: "// resolves cabinets (ARD Р-11): one call", marker: "ARD", found: true},
		{name: "requirement id", text: "// unique ids before the call (@ФТ-11)", marker: "@ФТ-11", requirement: true, found: true},
		{name: "requirement id without dash", text: "// не попадает в ответ (ФТ33)", marker: "ФТ33", requirement: true, found: true},
		{name: "document code", text: "// the same guard (B-48) as the sibling", marker: "B-48", found: true},
		{name: "task number", text: "// один вызов на страницу (задача 29)", marker: "задача 29", found: true},
		{name: "section", text: "// terminal code of the record, PRD §12", marker: "PRD", found: true},
		{name: "leftmost wins", text: "// (BACKLOG B-48) — задача 43", marker: "BACKLOG", found: true},
		{name: "bdd tag", text: "// повторное выключение (@negative)", marker: "@negative", requirement: true, found: true},
		{name: "commit", text: "// порядок шагов — коммит 34640e6", marker: "коммит 34640e6", found: true},
		{name: "explanation", text: "// a per-item resolve would cost N requests", found: false},
		{name: "rfc3339", text: "// the registry answers in RFC3339", found: false},
		{name: "utf-8", text: "// the payload stays UTF-8", found: false},
		{name: "env var", text: "// TRACE_ENABLED turns the exporter on", found: false},
	}

	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, ok := firstMarker(tt.text, markers)
			if ok != tt.found {
				t.Fatalf("firstMarker(%q): got found=%v, want %v (marker %q)", tt.text, ok, tt.found, found.text)
			}
			if !tt.found {
				return
			}
			if found.text != tt.marker {
				t.Errorf("firstMarker(%q): got marker %q, want %q", tt.text, found.text, tt.marker)
			}
			if got := tt.text[found.offset : found.offset+len(tt.marker)]; got != tt.marker {
				t.Errorf("firstMarker(%q): offset %d points at %q, want %q", tt.text, found.offset, got, tt.marker)
			}
			if gotRequirement := found.fix == fixRequirement; gotRequirement != tt.requirement {
				t.Errorf("firstMarker(%q): requirement fix=%v, want %v", tt.text, gotRequirement, tt.requirement)
			}
		})
	}
}

// TestInvalidPattern — a broken regexp in settings surfaces as an analysis
// error instead of silently disabling the marker.
func TestInvalidPattern(t *testing.T) {
	a := NewAnalyzer(Settings{Patterns: []string{`(unclosed`}})
	if _, err := a.Run(nil); err == nil {
		t.Fatal("run with an invalid pattern: got nil error, want a compile error")
	}
}
