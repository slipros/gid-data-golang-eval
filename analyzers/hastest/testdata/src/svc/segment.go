package svc

import "context"

// Repo — the dependency the segment saves through.
type Repo interface {
	Save(ctx context.Context, id string) error
}

// Segment — the service under test.
type Segment struct {
	repo Repo
	id   string
}

// NewSegment — a one-line constructor: trivial, never a candidate.
func NewSegment(repo Repo) *Segment {
	return &Segment{repo: repo}
}

// ID — a getter: trivial, never a candidate.
func (s *Segment) ID() string {
	return s.id
}

// Delete — a one-line delegation: trivial, never a candidate even though no
// test mentions it.
func (s *Segment) Delete(ctx context.Context, id string) error {
	return s.repo.Save(ctx, id)
}

// Create — non-trivial and called by TestSegment_Create.
func (s *Segment) Create(ctx context.Context, name string) error {
	if name == "" {
		return context.Canceled
	}

	return s.repo.Save(ctx, name)
}

// Handed — non-trivial and never called by a test, but handed to a helper as a
// value: a mention is a mention.
func (s *Segment) Handed(ctx context.Context, id string) error {
	if id == "" {
		return context.Canceled
	}

	return s.repo.Save(ctx, id)
}

// Rebuild — non-trivial and mentioned by no test.
// Reported by TestTestVariant: analysistest cannot drive this package
// (it also hands over the base variant, which the rule leaves alone).
func (s *Segment) Rebuild(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if err := s.repo.Save(ctx, id); err != nil {
			return err
		}
	}

	return nil
}

// DefaultLimits — a package-level var the test uses; the function building it
// therefore runs at package initialization.
var DefaultLimits = NewLimits(10, 20)

// Limits — the settings DefaultLimits holds.
type Limits struct {
	Min int
	Max int
}

// NewLimits — non-trivial and called by no test directly, but the test names
// DefaultLimits, whose initializer runs it: reporting it as untouched would be
// a false statement.
func NewLimits(min, max int) Limits {
	if min > max {
		min, max = max, min
	}

	return Limits{Min: min, Max: max}
}

// MemRepo — an implementation the test only ever drives through the Repo
// interface: the object resolved at that call site is Repo.Save, so Save counts
// as covered by name.
type MemRepo struct {
	saved []string
}

// Save — non-trivial, reached in the test through the interface.
func (m *MemRepo) Save(_ context.Context, id string) error {
	if id == "" {
		return context.Canceled
	}
	m.saved = append(m.saved, id)

	return nil
}

// normalize — unexported: not a candidate, it is reached through the exported
// surface the rule does judge.
func normalize(in []string) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		if v != "" {
			out = append(out, v)
		}
	}

	return out
}
