// Eval of GID-145 for /dal/repository.
package repository

import (
	"fmt"

	"github.com/pkg/errors"

	"dalsvc/dal/entity"
)

// --- Positive: declaring an error outside entity ---

var ErrConn = errors.New("conn") // want `GID-145: error "ErrConn" is declared in "dalsvc/dal/repository"\. Fix: keep this layer's errors in /dal/entity` `GID-145: creating an error via errors\.New is forbidden`

type Snapshot struct{}

// --- Positive: creating errors at runtime ---

func (s *Snapshot) bad() error {
	return fmt.Errorf("query failed") // want `GID-145: creating an error via fmt\.Errorf is forbidden\. Fix: exchange it for an error from /dal/entity`
}

// --- Negative: exchanging a connection error for an entity error ---

func (s *Snapshot) good(err error) error {
	if err != nil {
		return entity.ErrRowNotFound
	}
	return nil
}

// Negative (boundary): no exchange happened — the original connection error
// is passed through (enriched) — this is acceptable.
func (s *Snapshot) goodPassThrough(err error) error {
	return errors.Wrap(err, "select snapshot")
}
