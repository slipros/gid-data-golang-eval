package svc

import (
	"context"
	"testing"
)

func TestSegment_Create(t *testing.T) {
	var repo Repo = &MemRepo{}
	if err := repo.Save(context.Background(), "seed"); err != nil {
		t.Fatal(err)
	}

	s := NewSegment(repo)
	if err := s.Create(context.Background(), "name"); err != nil {
		t.Fatal(err)
	}

	drive(t, s.Handed)

	if got := normalize([]string{"", "a"}); len(got) != 1 {
		t.Fatalf("normalize: %v", got)
	}

	if DefaultLimits.Min > DefaultLimits.Max {
		t.Fatalf("limits: %+v", DefaultLimits)
	}
}

// NewFixture — an exported helper of the suite: scaffolding declared in a
// _test.go file is not a candidate, so it needs no test of its own.
func NewFixture(t *testing.T) *Segment {
	t.Helper()

	if testing.Short() {
		t.Skip("short mode")
	}

	return NewSegment(&MemRepo{})
}

func drive(t *testing.T, fn func(context.Context, string) error) {
	t.Helper()

	if err := fn(context.Background(), "x"); err != nil {
		t.Fatal(err)
	}
}
