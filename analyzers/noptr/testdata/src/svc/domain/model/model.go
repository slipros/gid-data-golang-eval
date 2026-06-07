// Eval GID-120/121 (указатели в model).
package model

import (
	"time"

	"github.com/gofrs/uuid"
)

type SnapshotStatus string

// --- Позитивные кейсы ---

type Snapshot struct {
	ParentID    *uuid.UUID     // want `GID-120: \*uuid\.UUID is forbidden\. Fix: use uuid\.UUID and check emptiness with IsNil\(\)`
	CompletedAt *time.Time     // want `GID-121: \*time\.Time is unnecessary in model\. Fix: use time\.Time and check absence with t\.IsZero\(\)`
	Description *string        // want `GID-121: a pointer to a string type is unnecessary in model\. Fix: use the value and check len\(s\) == 0`
	Status      *SnapshotStatus // want `GID-121: a pointer to a string type is unnecessary in model`
}

// Граничный кейс: *uuid.UUID в сигнатуре — тоже нарушение GID-120.
func Lookup(id *uuid.UUID) bool { // want `GID-120: \*uuid\.UUID is forbidden\. Fix: use uuid\.UUID and check emptiness with IsNil\(\)`
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
