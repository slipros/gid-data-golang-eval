// Eval of GID-255 positive (the rule's origin, ad-cabinet-connector
// internal/client/resourceregistry): the package sits in the client layer but
// declares no client — a connection factory plus its logging policy, no type,
// no method. Assembling a transport object out of options, a logger and
// metrics once at start-up is the composition root's job.
package resourceregistry // want `GID-255: package "resourceregistry" is in the client layer but declares no client`

// Options configures the connection.
type Options struct {
	Addr      string
	AuthToken string
}

// Connection stands in for the grpc library's connection type — note that it
// comes from ANOTHER package: the constructor below builds a foreign object,
// which is what a factory does and a client never does.
type connection struct {
	addr string
}

// NewConnection assembles the connection: address and service token.
func NewConnection(opts *Options) (*connection, error) {
	return &connection{addr: opts.Addr}, nil
}

// NewLoggingDecider builds the body-logging policy of the connection.
func NewLoggingDecider(logBody bool) func(string) bool {
	return func(string) bool { return logBody }
}
