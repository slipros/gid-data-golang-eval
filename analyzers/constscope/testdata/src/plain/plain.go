// Eval for GID-194: edge cases in an ordinary package outside the layers.
package plain

import "fmt"

// --- Edge: an iota group used entirely by a single function ---

const ( // want `GID-194: this constant group is used only in "stateName"\. Fix: declare it inside that function`
	stateIdle = iota
	stateBusy
)

func stateName(s int) string {
	switch s {
	case stateIdle:
		return "idle"
	case stateBusy:
		return "busy"
	}
	return fmt.Sprintf("unknown:%d", s)
}

// --- Edge: an iota group used by different functions — fine ---

const (
	colorRed = iota
	colorBlue
)

func isRed(c int) bool { return c == colorRed }

func isBlue(c int) bool { return c == colorBlue }

// --- Edge: an iota group with an exported constant — localization is not
// suggested, the diagnostic is only about the export ---

const (
	ModePrimary = iota // want `GID-194: exported constant "ModePrimary" is declared outside model/entity\. Fix: keep shared constants in /domain/model or /dal/entity, and declare local ones where they are used`
	modeSecondary
)

func modeLabel() int { return modeSecondary }

// --- Edge: a use in a package-level var — the constant is immovable ---

const defaultLabel = "default"

var currentLabel = defaultLabel

// --- Edge: a use in a signature (an array length) — immovable ---

const bufSize = 8

func fill(buf [bufSize]byte) byte { return buf[0] }

// --- Edge: an unused constant is the domain of unused, not GID-194 ---

const orphan = "unused"

// --- Edge: used only by a generated file — immovable ---

const genLabel = "gen"

// --- Edge: used only by a test — immovable ---

const testLabel = "test"

func use() (string, bool, bool, int, byte) {
	return stateName(0), isRed(1), isBlue(1), modeLabel(), fill([bufSize]byte{})
}

var _ = use

// --- Edge: a named-type string enum whose values are read by separate
// predicates — a definitional group, localized only as a whole, so no
// diagnostic: the enum stays grouped at the top (GID-123) ---

type taskStatus string

const (
	taskStarted taskStatus = "started"
	taskOK      taskStatus = "ok"
)

func isStarted(s taskStatus) bool { return s == taskStarted }

func isOK(s taskStatus) bool { return s == taskOK }

// --- Edge: a named-type enum used entirely by one function — moved as a whole,
// a single diagnostic on the block ---

type phase string

const ( // want `GID-194: this constant group is used only in "phaseName"\. Fix: declare it inside that function`
	phaseInit phase = "init"
	phaseDone phase = "done"
)

func phaseName(p phase) string {
	switch p {
	case phaseInit:
		return "init"
	case phaseDone:
		return "done"
	}
	return "unknown"
}
