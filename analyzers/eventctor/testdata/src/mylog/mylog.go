// Stub of a project-specific logger, used to prove settings.loggerTypes drives
// which parameter type counts as a logger.
package mylog

type Logger struct{}

func (l *Logger) With(_ ...any) *Logger { return l }
