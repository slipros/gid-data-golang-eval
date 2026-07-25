// Eval of the naming pair: the key is the layer the entity lives in, not a
// free-form word — /client/** names its entities under "client".
package debugsink

import "log/slog"

type Sink struct {
	logger *slog.Logger
}

// --- Positive: a free-form key instead of the layer ---

func NewSink(logger *slog.Logger) *Sink { // want `GID-154: the entity is named under a key other than its layer\. Fix: use the layer as the key — logger\.With\("client", "<entity_name>"\); a free-form key \("component"\) is not filterable in the logs`
	return &Sink{logger: logger.With("component", "sink")}
}

// --- Positive: the layer key, but a CamelCase entity name ---

func NewSinkCamel(logger *slog.Logger) *Sink {
	return &Sink{logger: logger.With("client", "DebugSink")} // want `GID-154: the entity name "DebugSink" is not lower snake_case\. Fix: spell it as the log fields do`
}

// --- Negative: the layer key and a snake_case entity name ---

func NewSinkGood(logger *slog.Logger) *Sink {
	return &Sink{logger: logger.With("client", "debug_sink")}
}
