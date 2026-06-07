// Eval GID-144 для /domain/service.
package service

import (
	"github.com/pkg/errors"

	"domainsvc/domain/model"
)

// --- Позитив: объявление ошибки вне model (даже через pkg/errors) ---

var ErrLocal = errors.New("local") // want `GID-144: ошибка "ErrLocal" объявлена в "domainsvc/domain/service" — ошибки этого слоя живут в /domain/model` `GID-144: создание ошибки через errors\.New запрещено`

type Snapshot struct{}

// --- Позитив: создание ошибок в рантайме ---

func (s *Snapshot) bad() error {
	return errors.New("ad-hoc") // want `GID-144: создание ошибки через errors\.New запрещено — обменивайте на ошибку из /domain/model \(Wrap/WithStack — допустимы\)`
}

func (s *Snapshot) badErrorf(id string) error {
	return errors.Errorf("snapshot %s", id) // want `GID-144: создание ошибки через errors\.Errorf запрещено`
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
