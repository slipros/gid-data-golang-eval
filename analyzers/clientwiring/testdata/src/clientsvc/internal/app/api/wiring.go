// Eval of GID-255 non-applicability: the very same connection factory, in the
// place it belongs — the composition root. /app/** is not the client layer,
// so a function-only package here is the norm.
package api

// Options configures the connection.
type Options struct {
	Addr string
}

type connection struct {
	addr string
}

// NewRegistryConnection assembles the registry connection during wiring.
func NewRegistryConnection(opts *Options) (*connection, error) {
	return &connection{addr: opts.Addr}, nil
}

// NewLoggingDecider builds the body-logging policy of the connection.
func NewLoggingDecider(logBody bool) func(string) bool {
	return func(string) bool { return logBody }
}
