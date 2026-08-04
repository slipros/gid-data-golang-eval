package svc

import "testing"

// Non-applicability: _test.go is skipped — a test legitimately assembles
// expected SQL and fixture strings from constants. This file carries no
// expectation comment: were it in scope, analysistest would fail here on an
// unexpected diagnostic.
func TestQuery(t *testing.T) {
	expected := "SELECT " + columns + " FROM " + table
	if expected == "" {
		t.Fatal("empty")
	}
}
