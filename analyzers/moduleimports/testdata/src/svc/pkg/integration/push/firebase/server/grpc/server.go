// Non-applicability: a transport layer of the module is not judged by default
// (settings.layers = ["domain", "dal"]) — what it takes from the core there is
// shared infrastructure, not core data.
package grpc

import (
	commonrouter "svc/internal/server/router"
)

// Server — the module's transport.
type Server struct {
	router commonrouter.Router
}
