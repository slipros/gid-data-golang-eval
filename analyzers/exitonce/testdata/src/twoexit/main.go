// Класс «позитив»: два os.Exit в func main — второй (повторный) запрещён.
package main

import "os"

func main() {
	if len(someArgs()) == 0 {
		os.Exit(2) // первый вызов — допустим
	}
	os.Exit(0) // want `GID-181: повторный os\.Exit в main — выходите из программы в одном месте`
}

func someArgs() []string {
	return nil
}
