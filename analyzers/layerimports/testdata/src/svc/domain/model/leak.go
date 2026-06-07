// Позитив (GID-227): model импортирует транспорт — запрещено,
// model — чистый словарь сервиса.
package model

import "svc/server/middleware" // want `GID-227: package "svc/domain/model" must not import "svc/server/middleware"\. Fix: domain/model is the pure vocabulary of the service; layers do not flow into it`

// Legacy2 тянет транспорт в словарь — нарушение.
type Legacy2 struct{}

func (l Legacy2) touch() {
	middleware.Noop()
}
