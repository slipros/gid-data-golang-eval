package custom

import (
	gdhttpserver "gitlab.gid.team/gid-data/tech/golang/libs/http.git/server"
)

// routes — the application's own routes.
func routes(r gdhttpserver.Router) error {
	_ = r

	return nil
}

// NewApplication — a service's own application router that takes no metrics
// function of its own (it reads them from a package-level registry), so the
// fingerprint heuristic cannot see it. settings.application-routers names it.
func NewApplication(routers ...gdhttpserver.RouterFunc) gdhttpserver.RouterFunc {
	_ = routers

	return routes
}

// Build — clean once NewApplication is on settings.application-routers.
func Build(opts *gdhttpserver.Options, log gdhttpserver.Logger) *gdhttpserver.Server {
	return gdhttpserver.NewServer(
		opts,
		log,
		gdhttpserver.NewSystemRouterWithConnectionsPings(true, nil, log, nil),
		NewApplication(routes),
	)
}
