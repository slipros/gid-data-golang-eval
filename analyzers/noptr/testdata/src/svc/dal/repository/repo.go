// Eval: GID-120 applies everywhere, GID-121 — only in model.
package repository

import (
	"time"

	"github.com/gofrs/uuid"
)

type cursor struct {
	id *uuid.UUID // want `GID-120: \*uuid\.UUID is forbidden\. Fix: use uuid\.UUID and check emptiness with IsNil\(\)`

	// GID-121 not applicable: outside model a pointer to time is allowed
	// (in entity this is covered by GID-122 via sql.NullTime).
	at *time.Time
}
