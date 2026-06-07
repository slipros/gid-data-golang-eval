// Позитив (GID-172): у клиента свои типы, dal-слой ему недоступен.
package snapshot

import (
	"svc/dal/entity"     // want `GID-172: пакету "svc/client/snapshot" запрещён импорт "svc/dal/entity" — у клиента свои типы, он ничего не знает о entity/repository из dal-слоя`
	"svc/dal/repository" // want `GID-172: пакету "svc/client/snapshot" запрещён импорт "svc/dal/repository" — у клиента свои типы, он ничего не знает о entity/repository из dal-слоя`

	"svc/domain/model"
)

type Client struct {
	repo *repository.Snapshot
}

// Негатив: model клиенту разрешён (он отдаёт наружу свои/доменные типы).
func (c *Client) Snapshot() model.Snapshot {
	return model.Snapshot{}
}

// Позитив выше: entity в client-слое запрещён.
func (c *Client) leak(in entity.Snapshot) {}
