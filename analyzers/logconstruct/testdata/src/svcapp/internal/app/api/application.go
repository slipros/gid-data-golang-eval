// Non-applicability: the composition root wires the service together and hands
// the logger to the components — each of those names itself (GID-154 fires
// there), so the root itself is exempt, as it is in GID-104 and GID-214.
package api

import "log/slog"

type Application struct {
	logger *slog.Logger
}

// New builds the application: it takes the logger and passes it on as is.
func New(logger *slog.Logger) (*Application, error) {
	return &Application{logger: logger}, nil
}
