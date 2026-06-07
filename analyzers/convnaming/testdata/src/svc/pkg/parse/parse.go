// Неприменимость: вне слоёв dal/domain/server/event паттерн не проверяется.
package parse

type Config struct{ Raw string }

func ConfigFromEnv(raw string) Config { return Config{Raw: raw} }
