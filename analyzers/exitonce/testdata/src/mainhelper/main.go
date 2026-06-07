// Класс «позитив + негатив»: пакет main.
// Хелпер с os.Exit вне main — запрещён; единственный os.Exit в конце main — ок.
package main

import "os"

// --- Позитивный кейс: exit-вызов вне func main (в хелпере) ---

func fail() {
	os.Exit(1) // want `GID-181: os\.Exit вне func main запрещён — верните error наверх`
}

// --- Негативный кейс: ровно один os.Exit в конце main ---

func run() error {
	return nil
}

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}
