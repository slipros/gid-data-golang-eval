// Чужой пакет с именем logrus, но другим import-путём (не sirupsen).
package logrus

type Logger struct{}

// New одноимённой функции из другого пакета — НЕ logrus.New(), не флагуем.
func New() *Logger { return &Logger{} }
