// Eval GID-123 in /domain/model (the model subpackage is a full-fledged model layer).
package enum

// --- Positive: alias to a basic string (a real event-collector case) ---

type ConsentEventType = string // want `GID-123: enum ConsentEventType must be a named type, not an alias`

// --- Positive: alias to a basic int ---

type Weight = int // want `GID-123: enum Weight must be a named type, not an alias`

// --- Positive: a group of untyped string constants (reported on the first) ---

const (
	RoleAdmin = "admin" // want `GID-123: a group of string constants\. Fix: declare a named string type`
	RoleUser  = "user"
)

// --- Negative: a correct enum — a named string type ---

type EventType string

const (
	EventTypeCreated EventType = "created"
	EventTypeDeleted EventType = "deleted"
)

// --- Edge case: a single untyped string const — ok ---

const DefaultRole = "guest"

// --- Edge case: a single const of a named int type — not an enum ---

type Limit int

const DefaultLimit Limit = 100
