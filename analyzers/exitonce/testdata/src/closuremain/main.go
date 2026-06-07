// Класс «граничный»: os.Exit внутри замыкания в main считается вызовом в main
// (а не «вне main»). Единственный такой вызов — диагностики нет.
package main

import "os"

func main() {
	done := func(code int) {
		os.Exit(code)
	}
	done(0)
}
