// Неприменимость: файл без циклов и конверсий — диагностика не выводится.
package noloop

func greet(name string) string {
	return "hello, " + name
}

func sum(a, b int) int {
	return a + b
}
