// Eval GID-120/121 (указатели в model).
package model

import (
	"time"

	"github.com/gofrs/uuid"
)

type SnapshotStatus string

// --- Позитивные кейсы ---

type Snapshot struct {
	ParentID    *uuid.UUID     // want `GID-120: \*uuid\.UUID запрещён — пустой UUID проверяется через IsNil\(\)`
	CompletedAt *time.Time     // want `GID-121: \*time\.Time в model не нужен — отсутствие проверяется t\.IsZero\(\)`
	Description *string        // want `GID-121: указатель на string-тип в model не нужен — пустота проверяется len\(s\) == 0`
	Status      *SnapshotStatus // want `GID-121: указатель на string-тип в model не нужен`
}

// Граничный кейс: *uuid.UUID в сигнатуре — тоже нарушение GID-120.
func Lookup(id *uuid.UUID) bool { // want `GID-120: \*uuid\.UUID запрещён — пустой UUID проверяется через IsNil\(\)`
	return id != nil
}

// --- Негативные кейсы ---

type Job struct {
	ID          uuid.UUID
	CompletedAt time.Time
	Description string
	Status      SnapshotStatus
	Enabled     *bool // указатель оправдан: false — валидное значение
	Parent      *Job  // вложенная структура — указатель допустим
}

// Неприменимость: разыменование — не тип.
func deref(id *bool) bool { return *id }
