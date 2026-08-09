// Package server is a stand-in for the gid.team http library
// (gitlab.gid.team/gid-data/tech/golang/libs/http.git/server): same import
// path, same constructor names and parameter names, bodies stripped. The rule
// matches the library by import path and finds the metrics parameter by name,
// so both have to be exactly what the real library has.
package server

import "time"

// Router — chi.Router in the real library.
type Router any

// RouterFunc — a route registrar. An alias in the library too.
type RouterFunc = func(r Router) error

// Logger — logger.Logger in the real library.
type Logger any

// Metrics — what the application router measures requests into.
type Metrics interface {
	IncrementServerRequest(method, host, route string, statusCode int, duration time.Duration)
}

// PanicRecoveryHandler — middleware.PanicRecoveryHandler in the library.
type PanicRecoveryHandler = func(v any)

// ConnectionPingFunc — a dependency health probe.
type ConnectionPingFunc = func() error

// Options — server options.
type Options struct {
	Addr string
}

// Version — the version served at /version.
type Version struct {
	Value string
}

// ApplicationRouterOptions — options of the application router.
type ApplicationRouterOptions struct {
	AccessLog any
}

// Server — the http server.
type Server struct{}

// NewServer returns a server serving the given routers.
func NewServer(options *Options, log Logger, routers ...RouterFunc) *Server {
	_, _, _ = options, log, routers

	return &Server{}
}

// NewSystemRouter mounts /version, /metrics, /ready and /health.
func NewSystemRouter(
	debug bool,
	version *Version,
	readinessService any,
	livelinessService any,
	routers ...RouterFunc,
) RouterFunc {
	_, _, _, _, _ = debug, version, readinessService, livelinessService, routers

	return func(r Router) error { return nil }
}

// NewSystemRouterWithConnectionsPings is NewSystemRouter with dependency pings.
func NewSystemRouterWithConnectionsPings(
	debug bool,
	version *Version,
	logger Logger,
	connectionsPings map[string]ConnectionPingFunc,
	routers ...RouterFunc,
) RouterFunc {
	_, _, _, _, _ = debug, version, logger, connectionsPings, routers

	return func(r Router) error { return nil }
}

// NewApplicationRouter wraps the application routes in the middleware chain:
// panic recovery, sentry, OpenTelemetry, metrics, access log.
func NewApplicationRouter(
	options *ApplicationRouterOptions,
	log Logger,
	metrics Metrics,
	panicRecoveryHandler PanicRecoveryHandler,
	routers ...RouterFunc,
) RouterFunc {
	_, _, _, _, _ = options, log, metrics, panicRecoveryHandler, routers

	return func(r Router) error { return nil }
}

// NewRouterGroup mounts the routers under a prefix. It wraps nothing.
func NewRouterGroup(prefix string, routers ...RouterFunc) RouterFunc {
	_, _ = prefix, routers

	return func(r Router) error { return nil }
}
