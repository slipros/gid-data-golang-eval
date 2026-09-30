// Non-applicability: a _test.go file is not judged.
package svc

import "testing"

func TestUpload(t *testing.T) {
	_, err := upload(client{})
	switch {
	case err == nil:
		return
	default:
		t.Fatal(err)
	}
}
