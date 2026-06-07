// Eval для GID-179 (channel buffer size is one or none).
package chanbuf

const maxWorkers = 10

// --- Позитивные кейсы (буфер > 1, константа) ---

func bufLiteral() {
	_ = make(chan int, 2) // want `GID-179: channel buffer 2 is not allowed \(only 0 or 1\)\. Fix: use an unbuffered channel or buffer 1, or justify a larger buffer with //nolint:gidchanbuf\.`
}

func bufNamedConst() {
	_ = make(chan int, maxWorkers) // want `GID-179: channel buffer 10 is not allowed \(only 0 or 1\)\. Fix: use an unbuffered channel or buffer 1, or justify a larger buffer with //nolint:gidchanbuf\.`
}

func bufConstExpr() {
	_ = make(chan string, 2*3) // want `GID-179: channel buffer 6 is not allowed \(only 0 or 1\)\. Fix: use an unbuffered channel or buffer 1, or justify a larger buffer with //nolint:gidchanbuf\.`
}

// --- Негативные кейсы (буфер 0 или 1, либо без размера) ---

func bufNone() {
	_ = make(chan int)
}

func bufZero() {
	_ = make(chan int, 0)
}

func bufOne() {
	_ = make(chan int, 1)
}

// --- Граничные кейсы (не матчатся) ---

// Размер — переменная: обоснован рантаймом, решает review.
func bufVariable(n int) {
	_ = make(chan int, n)
}

// Размер — вызов функции: не константа.
func bufFuncCall() {
	_ = make(chan int, size())
}

func size() int { return 5 }

// make для слайса и мапы — не канал.
func notChannel() {
	_ = make([]int, 5)
	_ = make(map[string]int, 5)
}
