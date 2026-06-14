package newderef

type point struct{ X, Y int }

// Positive: new() is forbidden.
func bad() *point {
	return new(point) // want `GID-005: avoid the new\(\) builtin\. Fix: use "&T\{\}" for structs or "var x T" instead of "new\(T\)"\.`
}

// Negative: &T{} instead of new(T).
func good() *point {
	return &point{}
}

// Not applicable: a user-defined function named new — not the builtin.
func boundary() int {
	newx := func(v int) int { return v }
	return newx(1)
}
