// Package httprouter implements rules GID-264 and GID-265: the shape of the
// HTTP server composition built on the gid.team http library
// (gitlab.gid.team/gid-data/tech/golang/libs/http.git/server).
//
// The library splits the routes in two. NewSystemRouter* mounts /version,
// /metrics, /ready, /health — plain handlers, no middleware. NewApplicationRouter
// mounts everything the application itself serves, and it is the only one that
// wraps the routes it is given:
//
//	rWithMetrics := r.With(
//		panicRecovery,
//		sentryhttp…Handle,
//		middleware.OpenTelemetry(pureRouter),
//		middleware.Metrics(metrics.IncrementServerRequest),
//		accessLogMiddleware,
//	)
//
// So an application route that reaches the server by any other path is served
// with no panic recovery, no sentry, no tracing, no metrics and no access log —
// and nothing about it looks broken: the endpoint answers, it is simply
// invisible and unprotected. That is what the two diagnostics of GID-264 are
// about, both seen in generated code (incident 2026-08-09):
//
//	// the application router is missing altogether
//	NewServer(&opts.HTTPServer, log,
//		NewSystemRouterWithConnectionsPings(debug, &v, log, pings),
//		newBookmarksRouter(handler, factory))          // ← bare, unwrapped
//
//	// the application router is nested in the system one
//	NewServer(&opts.HTTPServer, log,
//		NewSystemRouterWithConnectionsPings(debug, &v, log, pings,
//			NewApplicationRouter(nil, log, nil, nil, bookmarksRouter)))
//
// The canonical form registers the two routers side by side, as separate
// arguments of NewServer (lk-api internal/app/lk-api/service.go,
// govorun-server internal/app/api/application.go):
//
//	NewServer(&opts.HTTPServer, log,
//		NewSystemRouterWithConnectionsPings(debug, &v, log, pings),
//		NewApplicationRouter(nil, log, metrics.HTTP, nil, bookmarksRouter))
//
// GID-265 is the other half of the same incident: NewApplicationRouter accepts
// a nil metrics and silently substitutes dummyMetrics{}, whose
// IncrementServerRequest does nothing. The server then runs with a metrics
// middleware that measures into the void — a shape indistinguishable from a
// working one until somebody opens a dashboard. The parameter is located by
// NAME, not by position: the library has moved it (v1.7 takes
// (options, log, metrics, panicRecoveryHandler, routers…), the version
// govorun-server pins takes (log, metrics, routers…)).
//
// What counts as an application router, in order: the library's own
// NewApplicationRouter; a constructor named in settings.application-routers
// ("<pkg>.<Func>" or a bare "<Func>"); and — so that a service wrapping the
// library needs no configuration at all — any call handed a metrics increment
// function, i.e. a value of type func(string, string, string, int,
// time.Duration) (lk-api passes prometheus.HTTP.IncrementServerRequest into its
// own router.NewApplication). A wrapper that fits none of the three reads as a
// bare route and is reported; //nolint:gidhttprouter is the way out.
package httprouter

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
)

const (
	ruleComposition = "GID-264"
	ruleMetrics     = "GID-265"
)

// What a router argument is (the kind type is declared below).
const (
	// kindBare — anything else: the application's own router, handed over raw.
	kindBare kind = iota
	// kindSystem — a NewSystemRouter* call.
	kindSystem
	// kindApplication — a call that wraps the routes in the middleware chain.
	kindApplication
)

// defaultServerPackages — the http server library this rule knows.
var defaultServerPackages = []string{"gitlab.gid.team/gid-data/tech/golang/libs/http.git/server"}

// Analyzer — GID-264/GID-265 with default settings.
var Analyzer = NewAnalyzer(Settings{})

// kind — what a router argument is.
type kind int

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Packages — import path prefixes of the http server library. Replaces the
	// built-in list (a fork of the library, a vendored copy).
	Packages []string `json:"packages"`
	// ApplicationRouters — constructors that wrap application routes the way
	// NewApplicationRouter does: "<pkg>.<Func>" or a bare "<Func>". Additive.
	ApplicationRouters []string `json:"application-routers"`
}

// fromServerPackage reports whether the callee belongs to the http server
// library.
func (s Settings) fromServerPackage(fn *types.Func) bool {
	pkg := fn.Pkg()
	if pkg == nil {
		return false
	}

	for _, prefix := range s.Packages {
		if pkg.Path() == prefix || strings.HasSuffix(pkg.Path(), "/"+strings.TrimPrefix(prefix, "/")) {
			return true
		}
	}

	return false
}

// classify decides what a router argument is.
func (s Settings) classify(pass *analysis.Pass, arg ast.Expr) kind {
	call, ok := arg.(*ast.CallExpr)
	if !ok {
		return kindBare
	}

	fn := typeutil.StaticCallee(pass.TypesInfo, call)
	if fn == nil {
		return kindBare
	}

	if s.fromServerPackage(fn) {
		switch name := fn.Name(); {
		case strings.HasPrefix(name, "NewSystemRouter"):
			return kindSystem
		case name == "NewApplicationRouter":
			return kindApplication
		}
		// NewRouterGroup, NewProxyRouter and friends wrap nothing: they are
		// application routes like any other.
		return kindBare
	}

	if s.namedApplicationRouter(fn) || takesMetricsFunc(pass, call) {
		return kindApplication
	}

	return kindBare
}

// namedApplicationRouter reports whether the callee is listed in
// settings.application-routers.
func (s Settings) namedApplicationRouter(fn *types.Func) bool {
	for _, entry := range s.ApplicationRouters {
		if entry == fn.Name() {
			return true
		}
		if pkg := fn.Pkg(); pkg != nil && entry == pkg.Name()+"."+fn.Name() {
			return true
		}
	}

	return false
}

// NewAnalyzer builds the GID-264/GID-265 analyzer from linter settings.
func NewAnalyzer(s Settings) *analysis.Analyzer {
	if len(s.Packages) == 0 {
		s.Packages = defaultServerPackages
	}

	return &analysis.Analyzer{
		Name: "gidhttprouter",
		Doc: ruleComposition + ": application routes reach the http server through the application router, " +
			"registered as a separate argument of NewServer next to the system router; " +
			ruleMetrics + ": that router is given real metrics, not nil. " +
			"Fix: NewServer(opts, log, NewSystemRouter…(…), NewApplicationRouter(nil, log, metrics.HTTP, nil, routes))",
		Requires: astwalk.Requires,
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s)
		},
	}
}

func run(pass *analysis.Pass, s Settings) (any, error) {
	astwalk.NodesOf(pass, nil, func(_ *ast.File, call *ast.CallExpr) {
		fn := typeutil.StaticCallee(pass.TypesInfo, call)
		if fn == nil || !s.fromServerPackage(fn) {
			return
		}

		switch name := fn.Name(); {
		case name == "NewServer":
			checkServer(pass, call, fn, s)
		case strings.HasPrefix(name, "NewSystemRouter"):
			checkSystemRouter(pass, call, fn, s)
		case name == "NewApplicationRouter":
			checkMetrics(pass, call, fn)
		}
	})

	return nil, nil
}

// checkServer reports every router argument of NewServer that is neither the
// system router nor an application router — a route registered past the
// middleware chain.
func checkServer(pass *analysis.Pass, call *ast.CallExpr, fn *types.Func, s Settings) {
	for _, arg := range variadicArgs(call, fn) {
		if s.classify(pass, arg) != kindBare {
			continue
		}

		pass.Reportf(arg.Pos(),
			"%s: this router is registered on the http server directly, so its routes get no panic recovery, "+
				"no tracing, no metrics and no access log — those live in the application router. "+
				"Fix: wrap it, gdhttpserver.NewApplicationRouter(nil, log, metrics.HTTP, nil, %s)",
			ruleComposition, exprText(pass, arg))
	}
}

// checkSystemRouter reports routers nested in the system router. The system
// router mounts /version, /metrics, /ready and /health and applies no
// middleware to what it is given, so a route nested there is as exposed as one
// passed to NewServer raw — and an application router nested there merely hides
// that it belongs next to it.
func checkSystemRouter(pass *analysis.Pass, call *ast.CallExpr, fn *types.Func, s Settings) {
	for _, arg := range variadicArgs(call, fn) {
		if s.classify(pass, arg) == kindApplication {
			pass.Reportf(arg.Pos(),
				"%s: the application router is nested in the system router, which serves /version, /ready and "+
					"/health. Fix: register it as a separate argument of NewServer, "+
					"NewServer(opts, log, NewSystemRouter…(…), %s)",
				ruleComposition, exprText(pass, arg))

			continue
		}

		pass.Reportf(arg.Pos(),
			"%s: this router is nested in the system router, so its routes get no panic recovery, no tracing, "+
				"no metrics and no access log. Fix: move it into the application router, "+
				"NewServer(opts, log, NewSystemRouter…(…), gdhttpserver.NewApplicationRouter(nil, log, "+
				"metrics.HTTP, nil, %s))",
			ruleComposition, exprText(pass, arg))
	}
}

// checkMetrics reports NewApplicationRouter called with a nil metrics. The
// parameter is found by name: the library has moved it between versions.
func checkMetrics(pass *analysis.Pass, call *ast.CallExpr, fn *types.Func) {
	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		return
	}

	idx := paramIndex(sig, "metrics")
	if idx < 0 || idx >= len(call.Args) {
		return
	}
	if !isNil(pass, call.Args[idx]) {
		return
	}

	pass.Reportf(call.Args[idx].Pos(),
		"%s: the application router is given no metrics, so it falls back to a dummy that counts nothing and "+
			"every http request goes unmeasured. Fix: pass the service metrics, e.g. metrics.HTTP",
		ruleMetrics)
}

// takesMetricsFunc reports whether the call is handed a metrics increment
// function — the fingerprint of a service's own application router (lk-api
// passes prometheus.HTTP.IncrementServerRequest into router.NewApplication).
func takesMetricsFunc(pass *analysis.Pass, call *ast.CallExpr) bool {
	for _, arg := range call.Args {
		if tv, ok := pass.TypesInfo.Types[arg]; ok && isMetricsFunc(tv.Type) {
			return true
		}
	}

	return false
}

// isMetricsFunc matches the signature of Metrics.IncrementServerRequest:
// func(method, host, route string, statusCode int, duration time.Duration).
func isMetricsFunc(t types.Type) bool {
	// The package time.Duration comes from.
	const timePkgPath = "time"

	sig, ok := t.Underlying().(*types.Signature)
	if !ok {
		return false
	}

	params, results := sig.Params(), sig.Results()
	if results.Len() != 0 || params.Len() != 5 {
		return false
	}

	for i := range 3 {
		if !isBasicKind(params.At(i), types.String) {
			return false
		}
	}
	if !isBasicKind(params.At(3), types.Int) {
		return false
	}

	duration := params.At(4)

	named, isNamed := duration.Type().(*types.Named)
	if !isNamed {
		return false
	}

	obj := named.Obj()
	pkg := obj.Pkg()

	return obj.Name() == "Duration" && pkg != nil && pkg.Path() == timePkgPath
}

// isBasicKind reports whether the variable's type is the given basic kind.
func isBasicKind(v *types.Var, want types.BasicKind) bool {
	t := v.Type()

	basic, ok := t.Underlying().(*types.Basic)

	return ok && basic.Kind() == want
}

// variadicArgs returns the arguments that landed in the variadic router
// parameter of the call.
func variadicArgs(call *ast.CallExpr, fn *types.Func) []ast.Expr {
	sig, ok := fn.Type().(*types.Signature)
	if !ok || !sig.Variadic() {
		return nil
	}

	params := sig.Params()

	fixed := params.Len() - 1
	if len(call.Args) <= fixed || call.Ellipsis.IsValid() {
		// Nothing variadic, or the caller spread a slice — there are no
		// individual routers to judge.
		return nil
	}

	return call.Args[fixed:]
}

// paramIndex is the position of the parameter called name, or -1.
func paramIndex(sig *types.Signature, name string) int {
	params := sig.Params()
	for i := range params.Len() {
		param := params.At(i)
		if param.Name() == name {
			return i
		}
	}

	return -1
}

// isNil reports whether the expression is the predeclared nil.
func isNil(pass *analysis.Pass, expr ast.Expr) bool {
	id, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}

	obj, isNilObj := pass.TypesInfo.Uses[id].(*types.Nil)

	return isNilObj && obj != nil
}

// exprText renders the expression for the fix example: the callee's name for a
// call, the identifier itself otherwise, so the diagnostic reads like code.
func exprText(pass *analysis.Pass, expr ast.Expr) string {
	switch node := expr.(type) {
	case *ast.CallExpr:
		if fn := typeutil.StaticCallee(pass.TypesInfo, node); fn != nil {
			return fn.Name() + "(…)"
		}
	case *ast.Ident:
		return node.Name
	case *ast.SelectorExpr:
		return node.Sel.Name
	}

	return "routes"
}
