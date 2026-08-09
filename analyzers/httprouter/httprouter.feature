# language: en

Feature: GID-264/GID-265 — application routes go through the application router (gidhttprouter)
  As a developer
  I want the public routes of the service to be registered through
  gdhttpserver.NewApplicationRouter, next to the system router and with real
  metrics
  So that every application endpoint is served with panic recovery, sentry,
  tracing, metrics and an access log, instead of answering while being invisible
  and unprotected

  # The library (gitlab.gid.team/gid-data/tech/golang/libs/http.git/server)
  # splits the routes in two. NewSystemRouter* mounts /version, /metrics, /ready
  # and /health and applies NO middleware to what it is given.
  # NewApplicationRouter is the only one that wraps its routers:
  #   r.With(panicRecovery, sentry, OpenTelemetry, Metrics, accessLog)
  # So a route that arrives any other way is served bare — and looks perfectly
  # healthy, because the endpoint answers.
  #
  # Canonical shape (lk-api internal/app/lk-api/service.go, govorun-server
  # internal/app/api/application.go):
  #   NewServer(opts, log,
  #       NewSystemRouterWithConnectionsPings(debug, &v, log, pings),
  #       NewApplicationRouter(nil, log, metrics.HTTP, nil, routes))
  #
  # GID-264 — composition: every router argument of NewServer must be the system
  # router or an application router; a router nested in the system router is
  # reported too (a bare route stays bare there, and an application router only
  # hides that it belongs beside it).
  # GID-265 — the metrics of NewApplicationRouter must not be nil: the library
  # substitutes dummyMetrics{}, whose IncrementServerRequest does nothing, so
  # the metrics middleware measures into the void. The parameter is found by
  # NAME, not by position — the library moved it between versions
  # (v1.7: options, log, metrics, panicRecoveryHandler, routers…;
  # the version govorun-server pins: log, metrics, routers…).
  #
  # What counts as an application router, in order:
  #   1. the library's own NewApplicationRouter;
  #   2. a constructor named in settings.application-routers ("<pkg>.<Func>" or
  #      a bare "<Func>");
  #   3. any call handed a metrics increment function — a value of type
  #      func(string, string, string, int, time.Duration) — so a service that
  #      wraps the library needs no configuration (lk-api passes
  #      prometheus.HTTP.IncrementServerRequest into its own
  #      router.NewApplication).
  #
  # The library is matched by import path (settings.packages replaces the
  # built-in list), so a same-named NewServer from another package is not
  # judged. Escape hatch: //nolint:gidhttprouter.

  # --- Class 1: positive ---

  Scenario: positive — a bare router passed straight to NewServer
    Given "NewServer(opts, log, NewSystemRouterWithConnectionsPings(...), newBookmarksRouter(handler, factory))"
    When the gidhttprouter analyzer checks the file
    Then the diagnostic "GID-264: this router is registered on the http server directly, so its routes get no panic recovery, no tracing, no metrics and no access log" is reported on the argument

  Scenario: positive — the application router nested in the system router
    Given "NewSystemRouterWithConnectionsPings(debug, &v, log, pings, NewApplicationRouter(nil, log, m, nil, routes))"
    When the gidhttprouter analyzer checks the file
    Then the diagnostic "GID-264: the application router is nested in the system router" is reported on the nested call

  Scenario: positive — a bare route nested in the system router
    Given "NewSystemRouter(debug, &v, nil, nil, bookmarksRouter)"
    When the gidhttprouter analyzer checks the file
    Then the diagnostic "GID-264: this router is nested in the system router" is reported on the argument

  Scenario: positive — the application router with nil metrics
    Given "NewApplicationRouter(nil, log, nil, nil, bookmarksRouter)"
    When the gidhttprouter analyzer checks the file
    Then the diagnostic "GID-265: the application router is given no metrics, so it falls back to a dummy that counts nothing" is reported on the nil

  # --- Class 2: negative ---

  Scenario: negative — the canonical shape
    Given "NewServer(opts, log, NewSystemRouterWithConnectionsPings(...), NewApplicationRouter(nil, log, m, nil, routes))"
    When the gidhttprouter analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a service's own wrapper handed the metrics function
    Given "ownApplicationRouter(m.IncrementServerRequest, bookmarksRouter)" passed to NewServer
    When the gidhttprouter analyzer checks the file
    Then no diagnostic is reported
    # The metrics increment signature is the fingerprint of a real wrapper.

  # --- Class 3: boundary ---

  Scenario: boundary — a router handed over as a value, not a call
    Given "NewServer(opts, log, NewSystemRouter...(...), bookmarksRouter)"
    When the gidhttprouter analyzer checks the file
    Then the diagnostic "GID-264: this router is registered on the http server directly" is reported
    # A value is as bare as a call.

  Scenario: boundary — NewRouterGroup is not a wrapper
    Given "NewServer(opts, log, NewSystemRouter...(...), NewRouterGroup(\"/v1\", bookmarksRouter))"
    When the gidhttprouter analyzer checks the file
    Then the diagnostic "GID-264: this router is registered on the http server directly" is reported
    # It mounts routes under a prefix and applies no middleware.

  Scenario: boundary — a service with system routes only
    Given "NewServer(opts, log, NewSystemRouterWithConnectionsPings(...))" and no public routes
    When the gidhttprouter analyzer checks the file
    Then no diagnostic is reported

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a same-named NewServer from another package
    Given "foreign.NewServer(\":8080\", routes)" from a library that is not the gid.team one
    When the gidhttprouter analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — the wrapper is named in settings.application-routers
    Given settings.application-routers holding "custom.NewApplication" and "NewServer(opts, log, NewSystemRouter...(...), NewApplication(routes))"
    When the gidhttprouter analyzer checks the file
    Then no diagnostic is reported
    # Without the setting the same call is reported — the price of the setting
    # existing, covered by its own fixture.
