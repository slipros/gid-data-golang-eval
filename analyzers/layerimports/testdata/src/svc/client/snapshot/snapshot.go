// Позитив (GID-172): у клиента свои типы, dal-слой ему недоступен.
package snapshot

import (
	"svc/dal/entity"     // want `GID-172: package "svc/client/snapshot" must not import "svc/dal/entity"\. Fix: the client has its own types and knows nothing about entity/repository from the dal layer`
	"svc/dal/repository" // want `GID-172: package "svc/client/snapshot" must not import "svc/dal/repository"\. Fix: the client has its own types and knows nothing about entity/repository from the dal layer`

	"svc/domain/model" // want `GID-229: package "svc/client/snapshot" must not import "svc/domain/model"\. Fix: the client has its own types; model <-> client DTO conversion lives at the consumer`
)

type Client struct {
	repo *repository.Snapshot
}

// Позитив (GID-229): domain клиенту недоступен — у него свои типы.
func (c *Client) Snapshot() model.Snapshot {
	return model.Snapshot{}
}

// Позитив выше: entity в client-слое запрещён.
func (c *Client) leak(in entity.Snapshot) {}
