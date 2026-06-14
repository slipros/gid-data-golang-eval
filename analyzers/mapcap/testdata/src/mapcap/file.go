// Eval for GID-183 (map capacity hint when filling from range).
package mapcap

// --- Class 1: positive (make(map) without cap + unconditional fill in range) ---

// Fill from range over a slice.
func fillFromSlice(src []int) map[int]int {
	m := make(map[int]int) // want `GID-183: make without capacity while filling from range\. Fix: make\(map\[K\]V, len\(src\)\)`
	for _, v := range src {
		m[v] = v
	}
	return m
}

// var form of declaration.
func fillFromSliceVar(src []string) map[string]bool {
	var m = make(map[string]bool) // want `GID-183: make without capacity while filling from range\. Fix: make\(map\[K\]V, len\(src\)\)`
	for _, v := range src {
		m[v] = true
	}
	return m
}

// Fill from range over a map.
func fillFromMap(src map[string]int) map[string]int {
	m := make(map[string]int) // want `GID-183: make without capacity while filling from range\. Fix: make\(map\[K\]V, len\(src\)\)`
	for k, v := range src {
		m[k] = v
	}
	return m
}

// Fill from range over a string.
func fillFromString(src string) map[rune]int {
	m := make(map[rune]int) // want `GID-183: make without capacity while filling from range\. Fix: make\(map\[K\]V, len\(src\)\)`
	for _, r := range src {
		m[r] = 1
	}
	return m
}

// --- Class 2: negative ---

// make with capacity already specified — correct.
func withCapacity(src []int) map[int]int {
	m := make(map[int]int, len(src))
	for _, v := range src {
		m[v] = v
	}
	return m
}

// Fill without range — the size is not inferred from a collection.
func fillWithoutRange() map[int]int {
	m := make(map[int]int)
	m[1] = 1
	m[2] = 2
	return m
}

// range over a channel — the length is unknown, the size cannot be hinted.
func fillFromChan(src chan int) map[int]int {
	m := make(map[int]int)
	for v := range src {
		m[v] = v
	}
	return m
}

// --- Class 3: boundary (not matched) ---

// Conditional fill inside an if in the loop body — real size < len(src).
func conditionalFill(src []int) map[int]int {
	m := make(map[int]int)
	for _, v := range src {
		if v > 0 {
			m[v] = v
		}
	}
	return m
}

// m is used between make and the loop (fill outside the loop) — cancels the diagnostic.
func usedBeforeLoop(src []int) map[int]int {
	m := make(map[int]int)
	m[0] = 0
	for _, v := range src {
		m[v] = v
	}
	return m
}

// m is passed into a call between make and the loop — cancels the diagnostic.
func passedBeforeLoop(src []int) map[int]int {
	m := make(map[int]int)
	consume(m)
	for _, v := range src {
		m[v] = v
	}
	return m
}

func consume(m map[int]int) { _ = m }

// --- Class 4: non-applicability (no make(map)) ---

// make of a slice — not a map.
func makeSlice(src []int) []int {
	s := make([]int, 0)
	for _, v := range src {
		s = append(s, v)
	}
	return s
}
