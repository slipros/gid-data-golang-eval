// Eval GID-123 в /domain/model (подпакет model — полноправный model-слой).
package enum

// --- Позитив: alias на basic string (реальный кейс event-collector) ---

type ConsentEventType = string // want `GID-123: enum ConsentEventType must be a named type, not an alias`

// --- Позитив: alias на basic int ---

type Weight = int // want `GID-123: enum Weight must be a named type, not an alias`

// --- Позитив: группа нетипизированных string-констант (репорт на первой) ---

const (
	RoleAdmin = "admin" // want `GID-123: a group of string constants\. Fix: declare a named string type`
	RoleUser  = "user"
)

// --- Негатив: правильный enum — именованный string-тип ---

type EventType string

const (
	EventTypeCreated EventType = "created"
	EventTypeDeleted EventType = "deleted"
)

// --- Граничный: одиночная нетипизированная string-const — ок ---

const DefaultRole = "guest"

// --- Граничный: одиночная const именованного int-типа — не enum ---

type Limit int

const DefaultLimit Limit = 100
