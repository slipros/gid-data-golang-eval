// Package othernew содержит локальную функцию New, не связанную с pkg/errors.
package othernew

// New — одноимённая функция другого пакета; правило GID-136 её не задевает.
func New(message string) error { return nil }
