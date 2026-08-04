// Non-applicability: a _test.go file is not judged — the diagnostic is about
// the package's placement and is reported once per file, so a test file would
// only duplicate what the production file of the same package already says.
package redis

import "testing"

func TestCache(t *testing.T) {
	if (Cache{}) != (Cache{}) {
		t.Fatal("unexpected cache")
	}
}
