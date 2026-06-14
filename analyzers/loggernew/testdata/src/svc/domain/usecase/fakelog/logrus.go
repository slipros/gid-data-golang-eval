// A foreign package named logrus but with a different import path (not sirupsen).
package logrus

type Logger struct{}

// New is a same-named function from a different package — NOT logrus.New(), not flagged.
func New() *Logger { return &Logger{} }
