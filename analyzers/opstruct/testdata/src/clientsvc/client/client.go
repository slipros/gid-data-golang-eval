// Eval GID-210: неприменимость — Create-структура вне model/entity не флагается.
package client

import "time"

// /client — не model и не entity, правило не применяется даже с ID/CreatedAt/UpdatedAt.
type CreateJob struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt time.Time
}
