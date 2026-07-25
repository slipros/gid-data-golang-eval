// GID-003 on the versioned import path: a service on github.com/gofrs/uuid/v5
// is on the same library, so the rule fires exactly as it does on the bare
// path (pathseg.SameLibrary).
package uuidversionv5

import "github.com/gofrs/uuid/v5"

// Positive: NewV7 with a hand-handled error, on the /v5 path.
func badMust() (uuid.UUID, error) {
	return uuid.NewV7() // want `GID-003: UUIDs must be generated via uuid\.Must\. Fix: use uuid\.Must\(uuid\.NewV7\(\)\) instead of handling the error\.`
}

// Positive: a banned version on the /v5 path.
func badVersion() {
	_, _ = uuid.NewV4() // want `GID-003: .* instead of uuid\.NewV4\(\)\.`
}

// Negative: the canonical form.
func good() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}
