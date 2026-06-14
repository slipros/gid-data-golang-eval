// Positive (GID-172): the client has its own types, the dal layer is not available to it.
package snapshot

import (
	"svc/dal/entity"     // want `GID-172: package "svc/client/snapshot" must not import "svc/dal/entity"\. Fix: the client has its own types and knows nothing about entity/repository from the dal layer`
	"svc/dal/repository" // want `GID-172: package "svc/client/snapshot" must not import "svc/dal/repository"\. Fix: the client has its own types and knows nothing about entity/repository from the dal layer`

	"svc/domain/model" // want `GID-229: package "svc/client/snapshot" must not import "svc/domain/model"\. Fix: the client has its own types; model <-> client DTO conversion lives at the consumer`
)

type Client struct {
	repo *repository.Snapshot
}

// Positive (GID-229): domain is not available to the client — it has its own types.
func (c *Client) Snapshot() model.Snapshot {
	return model.Snapshot{}
}

// Positive above: entity is forbidden in the client layer.
func (c *Client) leak(in entity.Snapshot) {}
