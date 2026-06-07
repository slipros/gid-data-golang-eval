// Eval для GID-161 (no panic вне main).
package nopanic

import "errors"

// --- Позитивные кейсы ---

func bad() {
	panic("boom") // want `GID-161: panic используется только в пакете main — верните error`
}

// Граничный кейс: panic с error-аргументом.
func badErr(err error) {
	panic(err) // want `GID-161: panic используется только в пакете main — верните error`
}

// --- Негативные кейсы ---

func good() error {
	return errors.New("boom") //nolint // (GID-146 проверяется другим линтером)
}

// Граничный кейс: локальная функция с именем panic — не встроенный panic.
func shadowed() {
	panic := func(s string) {}
	panic("ok")
}
