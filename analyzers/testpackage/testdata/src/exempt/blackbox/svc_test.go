// Non-applicability: a deliberate black-box suite in an exempted directory.
package blackbox_test

import (
	"testing"

	"exempt/blackbox"
)

func TestValue(t *testing.T) {
	if blackbox.Value() != 42 {
		t.Fatal("unexpected")
	}
}
