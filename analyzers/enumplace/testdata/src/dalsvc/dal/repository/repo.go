// Eval GID-211: an enum in /dal/repository — also the DAL layer, outside /dal/entity/enum.
package repository

// --- Positive class: a string enum with const in /dal/repository — a violation ---

type Mode string // want `GID-211: enum Mode must live in /dal/entity/enum \(one file named after the entity\)\. Fix: move it there`

const (
	ModeRead  Mode = "read"
	ModeWrite Mode = "write"
)
