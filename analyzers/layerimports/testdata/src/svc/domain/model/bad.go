// Позитив: model не зависит от dal.
package model

import "svc/dal/entity" // want `GID-132: пакету "svc/domain/model" запрещён импорт "svc/dal/entity" — model не зависит от dal-слоя`

type Legacy struct {
	Raw entity.Snapshot
}
