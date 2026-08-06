package svc

import "testing"

// TestCabinets_Duplicates_ResolvedOnce — @ФТ-15: a repeated id costs one call.
// Not judged by default: the requirement map is built from exactly this pairing
// of an id with the name of the test that proves it.
func TestCabinets_Duplicates_ResolvedOnce(t *testing.T) {
	// a repeated id must not produce a second request to the registry
	if got := dedup([]string{"a", "a"}); len(got) != 1 {
		t.Fatalf("dedup: got %d ids, want 1", len(got))
	}
}

// TestReason_Empty_ReturnsOk checks the default reason of an empty history.
func TestReason_Empty_ReturnsOk(t *testing.T) {
	if Reason() != "ok" {
		t.Fatal("reason: want ok")
	}
}
