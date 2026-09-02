package app

import (
	"testing"

	"svc/service"
	"svc/usecase"
)

// A test file is not judged: the double satisfies an interface it does not own,
// and the assertion is how the test states which one. The wiring below does not
// make it redundant — and it proves nothing about the production assertions
// either, being checked by go test rather than by go build.
var _ usecase.LatestPageStore = testStore{}

type testStore struct{}

func (testStore) Object(key string) []byte { return []byte(key) }

func TestWire(t *testing.T) {
	if Wire() == nil {
		t.Fatal("no page")
	}

	// The only conversion of the catalog service lives here, in a test file:
	// the production assertion of it stays the check `go build` performs.
	_ = usecase.UseCatalog(&service.DatasetCatalog{})

	_ = usecase.NewLatestPage(nil, nil, testStore{}, nil)
}

// The same holds for an assertion of a production type written in a test file:
// the file it lives in is not judged, whatever the production code does with
// the type — here Wire() converts latestPageStore two lines into wiring.go.
var _ usecase.LatestPageStore = latestPageStore{}
