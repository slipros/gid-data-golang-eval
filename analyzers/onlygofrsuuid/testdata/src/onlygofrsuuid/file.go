// Eval for GID-137 (only-gofrs-uuid).
package onlygofrsuuid

import (
	"github.com/gofrs/uuid"

	googleuuid "github.com/google/uuid" // want `GID-137: importing "github.com/google/uuid" is forbidden\. Fix: use github.com/gofrs/uuid for UUID`

	satori "github.com/satori/go.uuid" // want `GID-137: importing "github.com/satori/go.uuid" is forbidden\. Fix: use github.com/gofrs/uuid for UUID`

	"example.com/uuidutil"
)

// --- Positive: the forbidden imports are caught above (including the alias boundary case) ---

func bad() googleuuid.UUID { return googleuuid.New() }

func badSatori() satori.UUID { return satori.NewV4() }

// --- Negative: the allowed library passes ---

func good() uuid.UUID { return uuid.Must(uuid.NewV7()) }

// --- Not applicable: a package with "uuid" in the name, but not a uuid library ---

func notApplicable() string { return uuidutil.Normalize("x") }
