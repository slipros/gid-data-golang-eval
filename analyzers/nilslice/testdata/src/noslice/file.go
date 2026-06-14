// Class 4: not applicable — a file without any slice literals at all.
package noslice

type Point struct {
	X, Y int
}

func add(a, b int) int {
	return a + b
}

func makePoint() Point {
	return Point{X: 1, Y: 2}
}

var counter = 0
