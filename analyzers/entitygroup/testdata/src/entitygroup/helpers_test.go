// Non-applicability: a _test.go file is not judged — a builder returning the
// entity under test belongs to the test, not to the entity's declaration file.
package entitygroup

import "testing"

func newBareUpload(id string) *Upload { return &Upload{id: id} }

func decorateID(id string) string { return "<" + id + ">" }

func (u *Upload) testOnlyName() string { return u.id }

func TestUploadID(t *testing.T) {
	if got := newBareUpload(decorateID("x")).ID(); got == "" {
		t.Fatal("empty id")
	}
}
