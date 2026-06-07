// Позитив (GID-227): model импортирует транспорт — запрещено,
// model — чистый словарь сервиса.
package model

import "svc/server/middleware" // want `GID-227: пакету "svc/domain/model" запрещён импорт "svc/server/middleware" — domain/model — чистый словарь сервиса, слои в него не текут`

// Legacy2 тянет транспорт в словарь — нарушение.
type Legacy2 struct{}

func (l Legacy2) touch() {
	middleware.Noop()
}
