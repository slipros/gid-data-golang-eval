package custombare

import (
	gdhttpserver "gitlab.gid.team/gid-data/tech/golang/libs/http.git/server"
)

// routes — the application's own routes.
func routes(r gdhttpserver.Router) error {
	_ = r

	return nil
}

// NewApplication — the same wrapper as in package custom, judged without
// settings.application-routers: nothing tells it apart from a bare route, so it
// is reported. That is the price of the setting existing.
func NewApplication(routers ...gdhttpserver.RouterFunc) gdhttpserver.RouterFunc {
	_ = routers

	return routes
}

// Build — reported until the wrapper is declared in the settings.
func Build(opts *gdhttpserver.Options, log gdhttpserver.Logger) *gdhttpserver.Server {
	return gdhttpserver.NewServer(
		opts,
		log,
		gdhttpserver.NewSystemRouterWithConnectionsPings(true, nil, log, nil),
		NewApplication(routes), // want `GID-264: this router is registered on the http server directly`
	)
}
