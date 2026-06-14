// Not applicable: the same struct in /dal/entity — db tags are mandatory here
// (GID-125), GID-168 does not apply here.
package entity

import "time"

type Snapshot struct {
	ID        string    `db:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at"`
}
