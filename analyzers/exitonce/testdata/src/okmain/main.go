// Класс «граничный»: defer + один os.Exit в main — это ровно один вызов, ок.
package main

import "os"

func cleanup() {}

func main() {
	defer cleanup()
	os.Exit(0)
}
