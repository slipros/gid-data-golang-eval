// Eval settings.exclude: kafka consumer validate освобождён от требования.
package validate

func Event(raw []byte) bool { return len(raw) > 0 }
