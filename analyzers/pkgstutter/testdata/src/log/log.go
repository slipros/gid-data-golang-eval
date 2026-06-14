// Eval for GID-193: the CamelCase word-boundary case — package log.
package log

// Logger — the package name log is only a prefix of the word "Logger", not a
// separate CamelCase word: there is no word boundary, so no stutter.
type Logger struct {
	prefix string
}

// Log — an exact match of the package name with no next word: not matched.
func Log(msg string) {}
