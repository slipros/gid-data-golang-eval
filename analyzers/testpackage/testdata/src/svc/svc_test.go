// Negative: the test lives in the same package and reaches the unexported
// helper directly.
package svc

import "testing"

func TestHelper(t *testing.T) {
	if helper() != 42 {
		t.Fatal("unexpected")
	}
}
