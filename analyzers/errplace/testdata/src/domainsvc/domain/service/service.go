// Eval GID-144 для /domain/service.
package service

import (
	"github.com/pkg/errors"

	"domainsvc/domain/model"
)

// --- Позитив: объявление ошибки вне model (даже через pkg/errors) ---

var ErrLocal = errors.New("local") // want `GID-144: error "ErrLocal" is declared in "domainsvc/domain/service"\. Fix: keep this layer's errors in /domain/model` `GID-144: creating an error via errors\.New is forbidden`

type Snapshot struct{}

// --- Позитив: создание ошибок в рантайме ---

func (s *Snapshot) bad() error {
	return errors.New("ad-hoc") // want `GID-144: creating an error via errors\.New is forbidden\. Fix: exchange it for an error from /domain/model \(Wrap/WithStack are allowed\)`
}

func (s *Snapshot) badErrorf(id string) error {
	return errors.Errorf("snapshot %s", id) // want `GID-144: creating an error via errors\.Errorf is forbidden`
}

// --- Негатив: обмен на ошибку из model и обогащение стеком ---

func (s *Snapshot) good(err error) error {
	if err != nil {
		return model.ErrSnapshotNotFound
	}
	return nil
}

func (s *Snapshot) goodWrap(err error) error {
	return errors.WithStack(err)
}
