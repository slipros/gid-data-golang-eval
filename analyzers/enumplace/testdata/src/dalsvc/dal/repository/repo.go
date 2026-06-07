// Eval GID-211: enum в /dal/repository — тоже DAL-слой, вне /dal/entity/enum.
package repository

// --- Позитивный класс: string-enum с const в /dal/repository — нарушение ---

type Mode string // want `GID-211: enum Mode живёт в /dal/entity/enum \(отдельный файл по имени сущности\)`

const (
	ModeRead  Mode = "read"
	ModeWrite Mode = "write"
)
