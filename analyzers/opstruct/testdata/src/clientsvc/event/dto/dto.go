// Eval GID-210: not applicable — a Create struct in /event/dto is not flagged.
package dto

import "time"

// /event/dto is neither model nor entity, the rule does not apply.
type CreateJob struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt time.Time
}
