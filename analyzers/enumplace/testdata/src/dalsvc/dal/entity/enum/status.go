// Eval GID-211: каноническое место enum — /dal/entity/enum, диагностики нет.
package enum

// --- Негативный класс: string-enum в своём месте — ок ---

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)
