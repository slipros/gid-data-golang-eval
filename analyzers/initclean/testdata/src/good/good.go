// Eval GID-180: негативные и граничные кейсы — диагностика не выводится.
package good

import "os"

var cfg = map[string]string{}

// Негатив: конструирование мапы в init — детерминированно, ок.
func init() {
	cfg["a"] = "1"
	cfg["b"] = "2"
}

// Негатив: чтение env в init разрешено (это не I/O).
func init() {
	cfg["host"] = os.Getenv("HOST")
	if v, ok := os.LookupEnv("PORT"); ok {
		cfg["port"] = v
	}
}

// Негатив: go-statement в обычной функции — правило к не-init не применяется.
func StartWorker() {
	go func() {}()
}

// Граничный (ограничение): хелпер вне init с os.Open, вызванный из init,
// НЕ матчится — анализ внутрипроцедурный, тело хелпера не обходится как init.
func loadFile() {
	_, _ = os.Open("/etc/hosts")
}

func init() {
	loadFile()
}
