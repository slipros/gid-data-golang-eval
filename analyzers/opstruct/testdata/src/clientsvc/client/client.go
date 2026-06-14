// Eval GID-210: not applicable — a Create struct outside model/entity is not flagged.
package client

import "time"

// /client is neither model nor entity, the rule does not apply even with ID/CreatedAt/UpdatedAt.
type CreateJob struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt time.Time
}
