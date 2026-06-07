// Eval для GID-126: неприменимость и граничные «похожие, но не Options» кейсы.
package model

// --- Неприменимость: пакет без Options-типов ---

type Job struct {
	ID   int
	Name string
}

var DefaultJob = Job{Name: "default"}

// --- Граничный: не-struct типы с именем Options не задеваются ---
// (alias на сущностный тип и interface — не голый struct Options)

type entOptions struct {
	Retries int
}

type Options = entOptions // alias — не задеваем

type OptionsProvider interface { // interface, не struct — не задеваем
	Opts() entOptions
}
