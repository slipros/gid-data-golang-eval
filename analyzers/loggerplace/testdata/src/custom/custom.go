// Non-applicability of the default suffixes: settings.suffixes replaces them,
// so only the configured suffix is checked.
package custom

import "log/slog"

// Checked: the configured "Params" suffix.
type SinkParams struct {
	Logger *slog.Logger // want `GID-251: options struct "SinkParams" holds a logger`
}

// Not checked any more: "Options" is not in the configured list.
type SinkOptions struct {
	Logger *slog.Logger
}
