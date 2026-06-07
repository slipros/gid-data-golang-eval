// Класс «граничный»: имя флага задано не строковой константой —
// часть 2 (snake_case) не проверяется (имя неизвестно статически).
// Пакет main, поэтому часть 1 не применяется.
package main

import "flag"

func name() string { return "maxRetries" }

func main() {
	dyn := name()
	flag.String(dyn, "3", "retries") // имя динамическое — диагностики нет
	flag.Parse()
}
