// Eval for GID-126: non-applicability and boundary "similar but not Options" cases.
package model

// --- Not applicable: a package without Options types ---

type Job struct {
	ID   int
	Name string
}

var DefaultJob = Job{Name: "default"}

// --- Boundary: non-struct types named Options are not affected ---
// (an alias to an entity type and an interface — not a bare struct Options)

type entOptions struct {
	Retries int
}

type Options = entOptions // an alias — not affected

type OptionsProvider interface { // an interface, not a struct — not affected
	Opts() entOptions
}
