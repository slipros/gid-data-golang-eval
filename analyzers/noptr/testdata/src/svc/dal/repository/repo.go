// Eval: GID-120 действует везде, GID-121 — только в model.
package repository

import (
	"time"

	"github.com/gofrs/uuid"
)

type cursor struct {
	id *uuid.UUID // want `GID-120: \*uuid\.UUID is forbidden\. Fix: use uuid\.UUID and check emptiness with IsNil\(\)`

	// Неприменимость GID-121: вне model указатель на time допустим
	// (в entity это закрывает GID-122 через sql.NullTime).
	at *time.Time
}
