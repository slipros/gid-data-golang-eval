// Eval for GID-215 (boundary): a package with segments dal and entity nested
// under a DIFFERENT layer (server/api) — not the module's entity layer.
package entity

// Misc — a type that looks like an entity by path segments alone, but the
// package is nested under server/api, not the module's /dal/entity.
type Misc struct {
	Value string
}
