// Non-applicability: a _test.go file is not judged. Tests live in the same
// package (GID-250), and a capture double named after the repository it fakes
// is scaffolding, not a service reaching for a foreign entity.
package service

import (
	"context"
	"testing"
)

// captureJobRepo records the calls the service under test makes: its own name
// has nothing to do with the entity of the repository it holds.
type captureJobRepo struct {
	repo JobRepository
	ids  []string
}

func (c *captureJobRepo) Job(ctx context.Context, id string) (string, error) {
	c.ids = append(c.ids, id)

	return c.repo.Job(ctx, id)
}

func TestCaptureJobRepo(t *testing.T) {
	c := captureJobRepo{}
	if len(c.ids) != 0 {
		t.Fatal("unexpected calls")
	}
}
