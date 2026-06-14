// Eval for GID-179 (channel buffer size is one or none).
package chanbuf

const maxWorkers = 10

// --- Positive cases (buffer > 1, a constant) ---

func bufLiteral() {
	_ = make(chan int, 2) // want `GID-179: channel buffer 2 is not allowed \(only 0 or 1\)\. Fix: use an unbuffered channel or buffer 1, or justify a larger buffer with //nolint:gidchanbuf\.`
}

func bufNamedConst() {
	_ = make(chan int, maxWorkers) // want `GID-179: channel buffer 10 is not allowed \(only 0 or 1\)\. Fix: use an unbuffered channel or buffer 1, or justify a larger buffer with //nolint:gidchanbuf\.`
}

func bufConstExpr() {
	_ = make(chan string, 2*3) // want `GID-179: channel buffer 6 is not allowed \(only 0 or 1\)\. Fix: use an unbuffered channel or buffer 1, or justify a larger buffer with //nolint:gidchanbuf\.`
}

// --- Negative cases (buffer 0 or 1, or no size at all) ---

func bufNone() {
	_ = make(chan int)
}

func bufZero() {
	_ = make(chan int, 0)
}

func bufOne() {
	_ = make(chan int, 1)
}

// --- Edge cases (not matched) ---

// The size is a variable: justified at runtime, review decides.
func bufVariable(n int) {
	_ = make(chan int, n)
}

// The size is a function call: not a constant.
func bufFuncCall() {
	_ = make(chan int, size())
}

func size() int { return 5 }

// make for a slice and a map — not a channel.
func notChannel() {
	_ = make([]int, 5)
	_ = make(map[string]int, 5)
}
