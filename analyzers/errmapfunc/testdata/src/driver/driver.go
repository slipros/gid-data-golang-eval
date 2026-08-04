// Package driver is a stand-in for a storage driver library (gdpostgres,
// pgx, go-redis): it publishes its error classification as bool-predicates
// over an error, not as sentinels compared with errors.Is. Shape (b) of
// GID-242 exists for exactly this — a mapper built on these predicates never
// mentions errors.Is and used to slip through.
package driver

// IsUniqueViolation reports whether err is a unique-index violation.
func IsUniqueViolation(err error) bool { return err != nil }

// IsNoResult reports whether err is an empty-result error.
func IsNoResult(err error) bool { return err != nil }

// IsTemporary reports whether err is a transient failure worth a retry.
func IsTemporary(err error) bool { return err != nil }
