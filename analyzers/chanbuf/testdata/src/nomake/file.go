// Неприменимость: файл без make — диагностика не выводится.
package nomake

func work() int {
	x := 0
	for i := 0; i < 10; i++ {
		x += i
	}
	return x
}
