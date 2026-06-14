// Eval of GID-144 for /domain/service.
package service

import (
	"github.com/pkg/errors"

	"domainsvc/domain/model"
)

// --- Positive: declaring an error outside model (even via pkg/errors) ---

var ErrLocal = errors.New("local") // want `GID-144: error "ErrLocal" is declared in "domainsvc/domain/service"\. Fix: keep this layer's errors in /domain/model` `GID-144: creating an error via errors\.New is forbidden`

type Snapshot struct{}

// --- Positive: creating errors at runtime ---

func (s *Snapshot) bad() error {
	return errors.New("ad-hoc") // want `GID-144: creating an error via errors\.New is forbidden\. Fix: exchange it for an error from /domain/model \(Wrap/WithStack are allowed\)`
}

func (s *Snapshot) badErrorf(id string) error {
	return errors.Errorf("snapshot %s", id) // want `GID-144: creating an error via errors\.Errorf is forbidden`
}

// --- Negative: exchange for an error from model and stack enrichment ---

func (s *Snapshot) good(err error) error {
	if err != nil {
		return model.ErrSnapshotNotFound
	}
	return nil
}

func (s *Snapshot) goodWrap(err error) error {
	return errors.WithStack(err)
}
