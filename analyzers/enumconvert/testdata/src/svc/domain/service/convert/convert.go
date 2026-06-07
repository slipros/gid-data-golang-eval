// Eval GID-143: map-конвертация enum обрабатывает отсутствующий ключ
// через gderror.NewUnhandledValueError. Действует в convert-пакете.
package convert

import gderror "gitlab.gid.team/gid-data/tech/golang/libs/helper.git/errors"

// Enum по GID-123: именованный тип на основе string.
type (
	EntityStatus string
	ModelStatus  string
)

// statusMap — конвертация enum→enum (оба ключ и значение именованные).
var statusMap = map[EntityStatus]ModelStatus{
	"active": "active",
}

// --- Класс 1: позитив ---

// Одиночное присваивание (не comma-ok): отсутствующий ключ молча даёт zero-value.
func ModelStatusFromEntity(s EntityStatus) ModelStatus {
	v := statusMap[s] // want `GID-143: enum conversion via map without comma-ok\. Fix: a missing key must return gderror.NewUnhandledValueError`
	return v
}

// Использование выражением (не comma-ok) — тоже без обработки.
func ModelStatusExpr(s EntityStatus) ModelStatus {
	return statusMap[s] // want `GID-143: enum conversion via map without comma-ok\. Fix: a missing key must return gderror.NewUnhandledValueError`
}

// comma-ok есть, но в функции НЕТ вызова NewUnhandledValueError.
func ModelStatusNoHandler(s EntityStatus) ModelStatus {
	v, ok := statusMap[s] // want `GID-143: a missing enum-conversion key must be handled with gderror.NewUnhandledValueError`
	if !ok {
		return ""
	}
	return v
}

// --- Класс 2: негатив ---

// comma-ok + обработка отсутствующего ключа через gderror.NewUnhandledValueError.
func ModelStatusFromEntityOK(s EntityStatus) (ModelStatus, error) {
	v, ok := statusMap[s]
	if !ok {
		return "", gderror.NewUnhandledValueError(s)
	}
	return v, nil
}

// --- Класс 3: граничный ---

// Ключ мапы — базовый string (не enum) → не матчится.
var titleMap = map[string]ModelStatus{"active": "active"}

func ModelTitleFromString(s string) ModelStatus {
	return titleMap[s]
}

// Значение мапы — базовый тип (не именованный) → не матчится (это не enum-конвертация).
var weightMap = map[EntityStatus]int{"active": 1}

func WeightFromStatus(s EntityStatus) int {
	return weightMap[s]
}
