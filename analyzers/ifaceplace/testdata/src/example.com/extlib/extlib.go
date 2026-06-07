// Package extlib — внешняя библиотека (путь без слой-сегментов сервиса).
package extlib

// Encoder — интерфейс внешней библиотеки. Использование разрешено.
type Encoder interface {
	Encode(v any) error
}
