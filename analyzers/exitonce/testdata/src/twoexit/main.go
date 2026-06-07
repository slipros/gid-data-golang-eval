// Класс «позитив»: два os.Exit в func main — второй (повторный) запрещён.
package main

import "os"

func main() {
	if len(someArgs()) == 0 {
		os.Exit(2) // первый вызов — допустим
	}
	os.Exit(0) // want `GID-181: duplicate os\.Exit in main\. Fix: exit the program in a single place`
}

func someArgs() []string {
	return nil
}
