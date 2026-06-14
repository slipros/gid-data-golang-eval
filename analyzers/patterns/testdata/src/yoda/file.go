package yoda

// Positive: a literal on the left — a yoda condition.
func bad(x int) bool {
	return 0 == x // want `GID-006: yoda condition — the literal must be on the right\. Fix: write "x == 0" instead of "0 == x"\.`
}

// Positive: same for !=.
func bad2(x int) bool {
	return 5 != x // want `GID-006: yoda condition`
}

// Negative: a variable on the left, a literal on the right.
func good(x int) bool {
	return x == 0
}

// Not applicable: both sides are non-constants.
func boundary(a, b int) bool {
	return a == b
}

// Not applicable: both sides are constants.
func boundaryConst() bool {
	return 1 == 1
}
