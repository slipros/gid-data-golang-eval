package exclcase

import "context"

// Segment — a service whose Rebuild is on settings.exclude.
type Segment struct {
	ids []string
}

// Create — non-trivial and covered by the test.
func (s *Segment) Create(ctx context.Context, id string) error {
	if id == "" {
		return context.Canceled
	}
	s.ids = append(s.ids, id)

	return ctx.Err()
}

// Rebuild — non-trivial and mentioned by no test, but excluded by name through
// settings.exclude ("Segment.Rebuild").
func (s *Segment) Rebuild(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if id == "" {
			return context.Canceled
		}
		s.ids = append(s.ids, id)
	}

	return ctx.Err()
}
