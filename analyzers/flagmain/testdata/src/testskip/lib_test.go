// A *_test.go file: flag is legitimate in tests — the analyzer skips such
// files, no diagnostic even in a non-main package and with a non-snake_case name.
package testskip

import (
	"flag"
	"testing"
)

var update = flag.Bool("updateGolden", false, "update golden files")

func TestAdd(t *testing.T) {
	flag.Parse()
	if Add(1, 2) != 3 {
		t.Fatal("bad")
	}
	_ = update
}
