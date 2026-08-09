package app

import (
	"time"

	gdhttpserver "gitlab.gid.team/gid-data/tech/golang/libs/http.git/server"
)

// metrics — a service metrics holder; IncrementServerRequest is the method the
// application router measures into.
type metrics struct{}

func (m *metrics) IncrementServerRequest(method, host, route string, statusCode int, duration time.Duration) {
	_, _, _, _, _ = method, host, route, statusCode, duration
}

// bookmarksRouter — the application's own routes.
func bookmarksRouter(r gdhttpserver.Router) error {
	_ = r

	return nil
}

// newBookmarksRouter — the same routes behind a constructor.
func newBookmarksRouter(handler, factory any) gdhttpserver.RouterFunc {
	_, _ = handler, factory

	return bookmarksRouter
}

// ownApplicationRouter — a service's own wrapper: it is handed the metrics
// increment function, which is how the rule recognises it without any setting.
func ownApplicationRouter(inc func(method, host, route string, statusCode int, duration time.Duration),
	routers ...gdhttpserver.RouterFunc) gdhttpserver.RouterFunc {
	_, _ = inc, routers

	return bookmarksRouter
}

// Canonical — the shape the rule asks for: system router and application router
// side by side, metrics passed in.
func Canonical(opts *gdhttpserver.Options, log gdhttpserver.Logger, m *metrics) *gdhttpserver.Server {
	return gdhttpserver.NewServer(
		opts,
		log,
		gdhttpserver.NewSystemRouterWithConnectionsPings(true, nil, log, nil),
		gdhttpserver.NewApplicationRouter(nil, log, m, nil, bookmarksRouter),
	)
}

// BareRouter — the application router is missing altogether.
func BareRouter(opts *gdhttpserver.Options, log gdhttpserver.Logger) *gdhttpserver.Server {
	return gdhttpserver.NewServer(
		opts,
		log,
		gdhttpserver.NewSystemRouterWithConnectionsPings(true, nil, log, nil),
		newBookmarksRouter(nil, nil), // want `GID-264: this router is registered on the http server directly`
	)
}

// BareRouterValue — the same, handed over as a value rather than a call.
func BareRouterValue(opts *gdhttpserver.Options, log gdhttpserver.Logger) *gdhttpserver.Server {
	return gdhttpserver.NewServer(
		opts,
		log,
		gdhttpserver.NewSystemRouterWithConnectionsPings(true, nil, log, nil),
		bookmarksRouter, // want `GID-264: this router is registered on the http server directly`
	)
}

// NestedApplicationRouter — the application router is nested in the system one.
func NestedApplicationRouter(opts *gdhttpserver.Options, log gdhttpserver.Logger, m *metrics) *gdhttpserver.Server {
	return gdhttpserver.NewServer(
		opts,
		log,
		gdhttpserver.NewSystemRouterWithConnectionsPings(true, nil, log, nil,
			gdhttpserver.NewApplicationRouter(nil, log, m, nil, bookmarksRouter), // want `GID-264: the application router is nested in the system router`
		),
	)
}

// NestedBareRouter — a bare route nested in the system router.
func NestedBareRouter(opts *gdhttpserver.Options, log gdhttpserver.Logger) *gdhttpserver.Server {
	return gdhttpserver.NewServer(
		opts,
		log,
		gdhttpserver.NewSystemRouter(true, nil, nil, nil,
			bookmarksRouter, // want `GID-264: this router is nested in the system router`
		),
	)
}

// NilMetrics — the application router falls back to a dummy that counts
// nothing.
func NilMetrics(opts *gdhttpserver.Options, log gdhttpserver.Logger) *gdhttpserver.Server {
	return gdhttpserver.NewServer(
		opts,
		log,
		gdhttpserver.NewSystemRouterWithConnectionsPings(true, nil, log, nil),
		gdhttpserver.NewApplicationRouter(nil, log, nil, nil, bookmarksRouter), // want `GID-265: the application router is given no metrics`
	)
}

// RouterGroup — NewRouterGroup wraps nothing, so it is an application route
// like any other.
func RouterGroup(opts *gdhttpserver.Options, log gdhttpserver.Logger) *gdhttpserver.Server {
	return gdhttpserver.NewServer(
		opts,
		log,
		gdhttpserver.NewSystemRouterWithConnectionsPings(true, nil, log, nil),
		gdhttpserver.NewRouterGroup("/v1", bookmarksRouter), // want `GID-264: this router is registered on the http server directly`
	)
}

// OwnWrapper — a service's own application router, recognised by the metrics
// increment function it is handed.
func OwnWrapper(opts *gdhttpserver.Options, log gdhttpserver.Logger, m *metrics) *gdhttpserver.Server {
	return gdhttpserver.NewServer(
		opts,
		log,
		gdhttpserver.NewSystemRouterWithConnectionsPings(true, nil, log, nil),
		ownApplicationRouter(m.IncrementServerRequest, bookmarksRouter),
	)
}

// SystemOnly — a service with no public routes at all.
func SystemOnly(opts *gdhttpserver.Options, log gdhttpserver.Logger) *gdhttpserver.Server {
	return gdhttpserver.NewServer(
		opts,
		log,
		gdhttpserver.NewSystemRouterWithConnectionsPings(true, nil, log, nil),
	)
}
