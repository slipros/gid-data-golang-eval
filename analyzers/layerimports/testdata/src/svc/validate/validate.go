// Позитив (GID-224): валидатор лезет в dal — запрещено;
// негатив: model валидатору разрешён.
package validate

import (
	"svc/dal/entity" // want `GID-224: пакету "svc/validate" запрещён импорт "svc/dal/entity" — валидаторы работают только с domain/model и типами запросов`

	"svc/domain/model"
)

type Snapshot struct{}

// Негатив: model — норма.
func (v *Snapshot) Validate(in model.Snapshot) error {
	return nil
}

// Позитив выше: entity валидатору недоступен.
func (v *Snapshot) leak(in entity.Snapshot) {}
