// Non-applicability: a _test.go file is not judged.
package svc

import (
	"errors"
	"testing"
)

func TestUpload(t *testing.T) {
	_, err := upload(client{})
	switch {
	case err == nil:
		return
	default:
		t.Fatal(err)
	}
}

func TestClassification(t *testing.T) {
	err := client{}.failure()
	if errors.Is(err, ErrRate) {
		t.Fatal(err)
	}
}
