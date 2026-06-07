// Eval GID-211: domain-слой не задевается — в model enum живёт в model (GID-132).
package model

// --- Класс неприменимости: string-enum в /domain/model — норма, диагностики нет ---

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)
