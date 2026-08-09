package exclcase

import (
	"context"
	"testing"
)

func TestSegment_Create(t *testing.T) {
	s := &Segment{}
	if err := s.Create(context.Background(), "id"); err != nil {
		t.Fatal(err)
	}
}
