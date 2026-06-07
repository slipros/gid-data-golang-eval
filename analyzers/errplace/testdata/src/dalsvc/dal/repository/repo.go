// Eval GID-145 для /dal/repository.
package repository

import (
	"fmt"

	"github.com/pkg/errors"

	"dalsvc/dal/entity"
)

// --- Позитив: объявление ошибки вне entity ---

var ErrConn = errors.New("conn") // want `GID-145: error "ErrConn" is declared in "dalsvc/dal/repository"\. Fix: keep this layer's errors in /dal/entity` `GID-145: creating an error via errors\.New is forbidden`

type Snapshot struct{}

// --- Позитив: создание ошибок в рантайме ---

func (s *Snapshot) bad() error {
	return fmt.Errorf("query failed") // want `GID-145: creating an error via fmt\.Errorf is forbidden\. Fix: exchange it for an error from /dal/entity`
}

// --- Негатив: обмен ошибки подключения на entity-ошибку ---

func (s *Snapshot) good(err error) error {
	if err != nil {
		return entity.ErrRowNotFound
	}
	return nil
}

// Негатив (граница): обмена не произошло — исходная ошибка подключения
// пробрасывается (с обогащением) — это допустимо.
func (s *Snapshot) goodPassThrough(err error) error {
	return errors.Wrap(err, "select snapshot")
}
