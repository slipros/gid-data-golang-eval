// Неприменимость: в main panic разрешён (bootstrap).
package main

func main() {
	panic("bootstrap failure is fatal")
}
