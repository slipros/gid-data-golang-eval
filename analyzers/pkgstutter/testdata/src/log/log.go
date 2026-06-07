// Eval для GID-193: граничный кейс границы CamelCase-слова — пакет log.
package log

// Logger — имя пакета log является лишь префиксом слова "Logger", а не
// отдельным CamelCase-словом: границы слова нет, заикания нет.
type Logger struct {
	prefix string
}

// Log — точное совпадение с именем пакета без следующего слова: не матчится.
func Log(msg string) {}
