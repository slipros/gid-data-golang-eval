// Позитив (settings.rules): своё правило SVC-1 банит legacy в domain/service.
package service

import "custom/legacy/store" // want `SVC-1: пакету "custom/domain/service" запрещён импорт "custom/legacy/store" — пакет legacy выпиливается`

type Snapshot struct {
	st *store.Store
}
