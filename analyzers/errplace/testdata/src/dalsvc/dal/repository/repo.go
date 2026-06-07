// Eval GID-145 для /dal/repository.
package repository

import (
	"fmt"

	"github.com/pkg/errors"

	"dalsvc/dal/entity"
)

// --- Позитив: объявление ошибки вне entity ---

var ErrConn = errors.New("conn") // want `GID-145: ошибка "ErrConn" объявлена в "dalsvc/dal/repository" — ошибки этого слоя живут в /dal/entity` `GID-145: создание ошибки через errors\.New запрещено`

type Snapshot struct{}

// --- Позитив: создание ошибок в рантайме ---

func (s *Snapshot) bad() error {
	return fmt.Errorf("query failed") // want `GID-145: создание ошибки через fmt\.Errorf запрещено — обменивайте на ошибку из /dal/entity`
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
