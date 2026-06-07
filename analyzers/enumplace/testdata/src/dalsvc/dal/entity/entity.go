// Eval GID-211: enum в корне /dal/entity вместо /dal/entity/enum.
package entity

// --- Позитивный класс: string-enum с const в /dal/entity — нарушение ---

type Status string // want `GID-211: enum Status must live in /dal/entity/enum \(one file named after the entity\)\. Fix: move it there`

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

// --- Негативный класс: string-тип без const — не enum, не флагается ---

type RawJSON string

// --- Граничный класс: alias на string с const — зона GID-123, не GID-211 ---

type Code = string

const (
	CodeA Code = "a"
	CodeB Code = "b"
)

// --- Граничный класс: именованный int-тип с const — не string-enum ---

type Priority int

const (
	PriorityLow  Priority = 1
	PriorityHigh Priority = 2
)

// --- Обычная сущность без enum — не трогаем ---

type Job struct {
	ID string
}
