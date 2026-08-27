// Eval for GID-125: the driver is hidden behind an in-house wrapper.
package repository

import (
	"gid.team/libs/pgstore"

	"wrapsvc/dal/entity"
)

// Clients reads clients through the wrapper.
type Clients struct {
	store *pgstore.Store
}

// ByID returns one client.
func (c *Clients) ByID(id string) entity.Client {
	return entity.Client{ID: id}
}
