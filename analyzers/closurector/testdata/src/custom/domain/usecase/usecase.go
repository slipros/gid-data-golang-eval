// Eval for GID-273 with custom settings: settings.exclude (both forms) and
// settings.option-suffixes replacing the defaults.
package usecase

// Sender — the receiver of the excluded method.
type Sender struct{}

// Setting — a custom option type: with option-suffixes: ["Setting"] its
// builder is the options convention.
type Setting func(string)

// contactFilter — non-applicability: excluded by the "Type.Method" form.
func (s *Sender) contactFilter(scope string) func(string) bool {
	return func(candidate string) bool { return candidate == scope }
}

// statusFilter — non-applicability: excluded by the "Function" form.
func statusFilter(status string) func(string) bool {
	return func(candidate string) bool { return candidate == status }
}

// timeoutSetting — non-applicability: the custom option suffix exempts it.
func timeoutSetting(value string) Setting {
	return func(string) { _ = value }
}

// retryOption — positive: the DEFAULT suffixes are replaced, so Option no
// longer exempts.
func retryOption(value string) func(string) { // want `GID-273: function retryOption builds and returns a closure`
	return func(string) { _ = value }
}
