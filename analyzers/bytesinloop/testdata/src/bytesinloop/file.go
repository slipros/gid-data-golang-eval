// Eval for GID-182 (avoid repeated string-to-byte conversions in loops).
package bytesinloop

const constStr = "const"

// --- Positive cases ---

// []byte("x") in a for loop.
func byteLiteralInFor() {
	for i := 0; i < 10; i++ {
		_ = []byte("hello") // want `GID-182: converting to \[\]byte inside a loop repeats the allocation\. Fix: compute it once before the loop\.`
	}
}

// []byte(constStr) in a range loop.
func byteConstInRange(items []int) {
	for range items {
		_ = []byte(constStr) // want `GID-182: converting to \[\]byte inside a loop repeats the allocation\. Fix: compute it once before the loop\.`
	}
}

// []rune("x") in a nested block of a loop.
func runeLiteralInNestedBlock() {
	for i := 0; i < 10; i++ {
		if i > 5 {
			{
				_ = []rune("world") // want `GID-182: converting to \[\]rune inside a loop repeats the allocation\. Fix: compute it once before the loop\.`
			}
		}
	}
}

// --- Negative cases ---

// []byte("x") outside a loop — it is computed once anyway.
func byteLiteralOutsideLoop() {
	_ = []byte("hello")
	_ = []rune("world")
}

// []byte(variable) in a loop — a conversion of a variable, not a constant.
func byteVariableInLoop(s string) {
	for i := 0; i < 10; i++ {
		_ = []byte(s)
	}
}

// --- Edge cases ---

// A closure is declared in the loop and contains []byte("x") — matched
// (the closure runs on every iteration).
func closureInLoop() {
	for i := 0; i < 10; i++ {
		fn := func() {
			_ = []byte("closure") // want `GID-182: converting to \[\]byte inside a loop repeats the allocation\. Fix: compute it once before the loop\.`
		}
		fn()
	}
}

// []byte(s), where s is a closure parameter: not a constant — not matched.
func closureParamInLoop() {
	for i := 0; i < 10; i++ {
		fn := func(s string) {
			_ = []byte(s)
		}
		fn("x")
	}
}
