// Eval: repository работает только с entity.
package repository

import (
	"svc/client/billing" // want `GID-228: пакету "svc/dal/repository" запрещён импорт "svc/client/billing" — dal не вызывает внешние API напрямую — клиента wiring'ует app`
	"svc/dal/entity"

	"svc/domain/model" // want `GID-132: пакету "svc/dal/repository" запрещён импорт "svc/domain/model" — dal-слой работает только с entity, domain-типы ему недоступны`
)

type Snapshot struct{}

// Негатив: entity в repo — норма.
func (s *Snapshot) Snapshot(id string) (entity.Snapshot, error) {
	return entity.Snapshot{ID: id}, nil
}

// Позитив выше: model в dal-слое запрещён.
func (s *Snapshot) leak(in *model.Snapshot) {}

// Позитив выше (GID-228): внешние API дёргает client, его wiring'ует app.
func (s *Snapshot) leakClient(c *billing.Client) {}
