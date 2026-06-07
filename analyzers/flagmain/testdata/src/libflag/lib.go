// Класс «позитив»: регистрация флага в библиотечном (не main) пакете запрещена.
package libflag

import "flag"

var maxRetries = flag.Int("max_retries", 3, "retries") // want `GID-192: registering a flag outside package main is forbidden\. Fix: declare flags in the binary, let libraries take parameters`

func register() {
	flag.String("addr", ":8080", "listen addr") // want `GID-192: registering a flag outside package main is forbidden\. Fix: declare flags in the binary, let libraries take parameters`
}

// Граничный: имя флага динамическое — часть 2 не считаем, но часть 1
// (регистрация вне main) всё равно срабатывает.
func registerDynamic(name string) {
	flag.String(name, "", "addr") // want `GID-192: registering a flag outside package main is forbidden\. Fix: declare flags in the binary, let libraries take parameters`
}
