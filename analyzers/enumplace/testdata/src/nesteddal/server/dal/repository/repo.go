// Eval GID-211: boundary — a "dal" segment nested under another layer
// (server), not anchored to the module root, is not the DAL layer. Without
// the HasLayer fix, pathseg.Contains would have falsely matched the "dal"
// segment here and flagged this enum.
package repository

// --- Boundary class: a string enum with const in server/dal/repository —
// "dal" is not the leading (root-anchored) layer segment here, so the rule
// does not apply.

type Mode string

const (
	ModeRead  Mode = "read"
	ModeWrite Mode = "write"
)
