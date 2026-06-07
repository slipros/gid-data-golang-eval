// Неприменимость: вне service/usecase/repository приватные функции
// пакета — норма.
package util

func helper(s string) string { return s }

func Public(s string) string { return helper(s) }
