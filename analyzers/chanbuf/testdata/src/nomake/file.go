// Not applicable: a file without make — no diagnostic is emitted.
package nomake

func work() int {
	x := 0
	for i := 0; i < 10; i++ {
		x += i
	}
	return x
}
