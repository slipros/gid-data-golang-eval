// The service layer names its entities under "service".
package service

import "log/slog"

type Email struct {
	logger *slog.Logger
}

// --- Positive: named under "component" in the service layer ---

func NewEmail(logger *slog.Logger) *Email { // want `GID-154: the entity is named under a key other than its layer\. Fix: use the layer as the key — logger\.With\("service", "<entity_name>"\)`
	return &Email{logger: logger.With("component", "email")}
}

// --- Negative ---

func NewEmailGood(logger *slog.Logger) *Email {
	return &Email{logger: logger.With("service", "email")}
}
