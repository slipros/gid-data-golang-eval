package withtests

import "testing"

// TestReason_Empty_ReturnsOk — @ФТ-15: the default reason of an empty history. // want `Fix: state the constraint itself in the comment, and move the requirement id into the requirement map`
func TestReason_Empty_ReturnsOk(t *testing.T) {
	if Reason() != "ok" {
		t.Fatal("reason: want ok")
	}
}
