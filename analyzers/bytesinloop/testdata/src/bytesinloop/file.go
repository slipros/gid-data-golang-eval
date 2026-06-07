// Eval для GID-182 (avoid repeated string-to-byte conversions in loops).
package bytesinloop

const constStr = "const"

// --- Позитивные кейсы ---

// []byte("x") в for.
func byteLiteralInFor() {
	for i := 0; i < 10; i++ {
		_ = []byte("hello") // want `GID-182: конверсия \[\]byte в цикле — вычислите один раз перед циклом`
	}
}

// []byte(constStr) в range.
func byteConstInRange(items []int) {
	for range items {
		_ = []byte(constStr) // want `GID-182: конверсия \[\]byte в цикле — вычислите один раз перед циклом`
	}
}

// []rune("x") во вложенном блоке цикла.
func runeLiteralInNestedBlock() {
	for i := 0; i < 10; i++ {
		if i > 5 {
			{
				_ = []rune("world") // want `GID-182: конверсия \[\]rune в цикле — вычислите один раз перед циклом`
			}
		}
	}
}

// --- Негативные кейсы ---

// []byte("x") вне цикла — вычисляется один раз и так.
func byteLiteralOutsideLoop() {
	_ = []byte("hello")
	_ = []rune("world")
}

// []byte(variable) в цикле — конверсия переменной, не константы.
func byteVariableInLoop(s string) {
	for i := 0; i < 10; i++ {
		_ = []byte(s)
	}
}

// --- Граничные кейсы ---

// Замыкание объявлено в цикле и содержит []byte("x") — матчится
// (замыкание выполняется на каждой итерации).
func closureInLoop() {
	for i := 0; i < 10; i++ {
		fn := func() {
			_ = []byte("closure") // want `GID-182: конверсия \[\]byte в цикле — вычислите один раз перед циклом`
		}
		fn()
	}
}

// []byte(s), где s — параметр замыкания: не константа — не матчится.
func closureParamInLoop() {
	for i := 0; i < 10; i++ {
		fn := func(s string) {
			_ = []byte(s)
		}
		fn("x")
	}
}
