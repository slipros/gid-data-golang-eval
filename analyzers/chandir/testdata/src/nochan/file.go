// Неприменимость: файл без каналов в сигнатурах — диагностика не выводится.
package nochan

func add(a, b int) int {
	return a + b
}

func greet(name string) string {
	return "hi, " + name
}
