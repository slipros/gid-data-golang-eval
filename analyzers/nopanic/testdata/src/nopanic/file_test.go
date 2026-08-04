// Non-applicability: a _test.go file is not judged. A must-helper feeding a
// package-level fixture has no *testing.T to fail through — panic is the only
// way to report at that point.
package nopanic

import "testing"

var answer = mustParse("42")

func mustParse(s string) string {
	if s == "" {
		panic("empty")
	}

	return s
}

func TestAnswer(t *testing.T) {
	if answer == "" {
		t.Fatal("empty")
	}
}
