// Eval GID-211: enum в /dal/repository — тоже DAL-слой, вне /dal/entity/enum.
package repository

// --- Позитивный класс: string-enum с const в /dal/repository — нарушение ---

type Mode string // want `GID-211: enum Mode must live in /dal/entity/enum \(one file named after the entity\)\. Fix: move it there`

const (
	ModeRead  Mode = "read"
	ModeWrite Mode = "write"
)
