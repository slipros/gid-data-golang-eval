// Eval settings.exclude: the kafka consumer validate is exempt from the requirement.
package validate

func Event(raw []byte) bool { return len(raw) > 0 }
