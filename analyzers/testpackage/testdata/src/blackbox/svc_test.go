// Positive: an external test package keeps the test away from the unexported
// code of the package under test.
package blackbox_test // want `GID-250: an external test package "blackbox_test" keeps the test away from the unexported code it tests\. Fix: declare "blackbox" in this file`

import (
	"testing"

	"blackbox"
)

func TestValue(t *testing.T) {
	if blackbox.Value() != 42 {
		t.Fatal("unexpected")
	}
}
