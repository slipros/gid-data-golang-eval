// Eval for GID-267, non-applicability: a test lives in the same package as the
// code (GID-250), and a double of the client the service is wired to has to
// speak that client's types. The import mirrors what the production wiring
// already declares, so a _test.go file is not judged and expects no diagnostic.
package service

import (
	"testing"

	"svc/client/billing"
)

func TestSnapshotStub(t *testing.T) {
	var stub *billing.Client
	if stub != nil {
		t.Fatal("stub is not wired in this fixture")
	}
}
