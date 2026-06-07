// Eval GID-210: неприменимость — Create-структура в /event/dto не флагается.
package dto

import "time"

// /event/dto — не model и не entity, правило не применяется.
type CreateJob struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt time.Time
}
