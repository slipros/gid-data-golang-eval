// Класс 4: неприменимость — файл без слайсовых литералов вообще.
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
