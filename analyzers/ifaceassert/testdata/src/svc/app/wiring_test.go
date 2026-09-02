package app

import (
	"testing"

	"svc/service"
	"svc/usecase"
)

// The double is asserted and wired in the same test file: the assertion is
// redundant here for the same reason it is redundant in production code.
var _ usecase.LatestPageStore = testStore{} // want `GID\-274:\ redundant\ compile\-time\ assertion:\ the\ package\ already\ passes\ this\ value\ as\ usecase\.LatestPageStore\ at\ wiring_test\.go:27,\ so\ the\ compiler\ checks\ the\ contract\ there\.\ Fix:\ delete\ the\ "var\ _\ usecase\.LatestPageStore\ =\ testStore\{\}"\ line`

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
