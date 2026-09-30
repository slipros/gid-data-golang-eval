// Eval of GID-194 in a library module: no internal/app and no domain+dal pair,
// so the module is not laid out as a service (modlayout.IsServiceModule is false).
package lib

// --- Negative: an exported constant is public API — there is no model/entity to move it to ---

const ContextKeyQueryName = "trino.query-name"

// --- Negative: an exported enum block (iota) ---

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
)

// --- Negative: an exported constant used by exactly one function ---

const DefaultTimeout = 30

func Timeout() int { return DefaultTimeout }

func QueryName() string { return ContextKeyQueryName }

func Levels() Level { return LevelDebug + LevelInfo }

// --- Positive: an UNEXPORTED constant used by exactly one function is still advice about the code ---

const onlyHere = "x" // want `GID-194: constant "onlyHere" is used only in "Only"\. Fix: declare it inside that function`

func Only() string { return onlyHere }

// --- Negative: an unexported constant shared by two functions stays package-level ---

const shared = "s"

func A() string { return shared }
func B() string { return shared + shared }
