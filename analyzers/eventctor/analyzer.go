// Package eventctor implements rule GID-216 (event-ctor-deps): dependencies
// of event-layer constructors (source: event.md).
//
//   - GID-216 (gideventctor): in the consumer layer the constructor must take
//     a logger — the consumer enriches it with broker/consumer fields; in the
//     producer layer the constructor takes no logger — errors are propagated to
//     the caller.
//
// Which parameter types count as a logger is configurable via
// settings.loggerTypes (<package>.<Type>) — the default set covers both the
// logrus era (logrus.Logger, logrus.Entry) and slog (slog.Logger), so the rule
// is not pinned to a single logging stack.
//
// Scope is determined by the import path: the package must be in the event
// layer (anchored to the module root — pathseg.HasLayer), and its own name
// (leaf segment) must be consumer or producer (pathseg.EndsWith). Leaf
// packages named validate or convert are excluded: they hold validators and
// converters, not consumers/producers.
//
// A constructor is an exported function ^New[A-Z] returning a pointer to a
// struct type declared IN THE SAME PACKAGE. This automatically excludes
// schema functions like New<X>Schema that return *registry.Schema of a
// foreign package.
//
// Exclusions: constructor names in settings.exclude (.golangci.yml) or
// pointwise //nolint:gideventctor.
package eventctor

import (
	"go/ast"
	"go/types"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-216"

// Check modes for an event-layer package (the scope type is declared below).
const (
	scopeNone scope = iota
	scopeConsumer
	scopeProducer
)

// defaultLoggerTypes — the logger parameter types accepted in a consumer
// constructor, as <package>.<Type>. Covers the logrus era (Logger/Entry) and
// slog so the rule is not tied to one logging stack.
var defaultLoggerTypes = []string{"logrus.Logger", "logrus.Entry", "slog.Logger"}

// ctorName — a constructor: an exported name of the form New + a capital letter.
var ctorName = regexp.MustCompile(`^New[A-Z]`)

// Analyzer — variant with default settings (no exclusions).
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — names of excluded constructors (for example "NewOrderConsumer").
	Exclude []string `json:"exclude"`
	// LoggerTypes — the parameter types a consumer constructor may take as its
	// logger, each as <package>.<Type> (e.g. "logrus.Logger", "slog.Logger").
	// Empty → defaultLoggerTypes.
	LoggerTypes []string `json:"loggerTypes"`
}

// NewAnalyzer builds the GID-216 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(cfg Settings) *analysis.Analyzer {
	loggerTypes := cfg.LoggerTypes
	if len(loggerTypes) == 0 {
		loggerTypes = defaultLoggerTypes
	}
	loggers := make(map[string]struct{}, len(loggerTypes))
	for _, t := range loggerTypes {
		loggers[t] = struct{}{}
	}
	return &analysis.Analyzer{
		Name: "gideventctor",
		Doc: ruleID + ": consumer constructors take a logger (" + strings.Join(loggerTypes, ", ") +
			"), producer constructors do not. Fix: add a logger to consumer constructors, remove it from producers",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, cfg, loggers)
		},
	}
}

// scope — the check mode for an event-layer package.
type scope int

func pkgScope(pkgPath string) scope {
	// The event layer is anchored to the module root: a package nested under a
	// different layer (e.g. .../server/grpc/event/consumer) is NOT the event
	// layer, so pathseg.HasLayer (not Contains) is used here.
	if !pathseg.HasLayer(pkgPath, "event") {
		return scopeNone
	}
	// validate/convert are leaf packages (validators/converters, not
	// consumers/producers) — matched by the package's own name via EndsWith.
	if pathseg.EndsWith(pkgPath, "validate") || pathseg.EndsWith(pkgPath, "convert") {
		return scopeNone
	}
	switch {
	case pathseg.EndsWith(pkgPath, "consumer"):
		return scopeConsumer
	case pathseg.EndsWith(pkgPath, "producer"):
		return scopeProducer
	default:
		return scopeNone
	}
}

func run(pass *analysis.Pass, cfg Settings, loggers map[string]struct{}) (any, error) {
	sc := pkgScope(pass.Pkg.Path())
	if sc == scopeNone {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !fn.Name.IsExported() {
				continue
			}
			if !ctorName.MatchString(fn.Name.Name) {
				continue
			}
			if exclude.Match(cfg.Exclude, "", fn.Name.Name) {
				continue
			}
			obj, ok := pass.TypesInfo.Defs[fn.Name].(*types.Func)
			if !ok {
				continue
			}
			sig, ok := obj.Type().(*types.Signature)
			if !ok {
				continue
			}
			if !returnsLocalStructPtr(pass.Pkg, sig) {
				continue
			}
			check(pass, sc, fn, sig, loggers)
		}
	}
	return nil, nil
}

func check(pass *analysis.Pass, sc scope, fn *ast.FuncDecl, sig *types.Signature, loggers map[string]struct{}) {
	has := hasLoggerParam(sig, loggers)
	switch sc {
	case scopeConsumer:
		if !has {
			pass.Reportf(fn.Name.Pos(),
				"%s: a consumer constructor must take a logger and enrich it with broker/consumer fields "+
					"(see event.md). Fix: add a logger parameter (e.g. *slog.Logger) and attach the "+
					"broker/consumer fields in the constructor",
				ruleID)
		}
	case scopeProducer:
		if has {
			pass.Reportf(fn.Name.Pos(),
				"%s: a producer constructor must not take a logger; errors are propagated to the caller. "+
					"Fix: remove the logger (intentional exception: //nolint:gideventctor)",
				ruleID)
		}
	}
}

// returnsLocalStructPtr reports whether the signature returns (as the first
// result) a pointer to a struct type declared in the current package.
func returnsLocalStructPtr(pkg *types.Package, sig *types.Signature) bool {
	results := sig.Results()
	if results.Len() == 0 {
		return false
	}
	first := results.At(0)
	ptr, ok := first.Type().(*types.Pointer)
	if !ok {
		return false
	}
	named, ok := types.Unalias(ptr.Elem()).(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	if obj.Pkg() != pkg {
		return false
	}
	_, ok = named.Underlying().(*types.Struct)
	return ok
}

// hasLoggerParam reports whether any of the parameters has a logger type from
// the configured allowlist (matched by <package>.<Type>).
func hasLoggerParam(sig *types.Signature, loggers map[string]struct{}) bool {
	params := sig.Params()
	for i := 0; i < params.Len(); i++ {
		param := params.At(i)
		key, ok := loggerKey(param.Type())
		if !ok {
			continue
		}
		if _, ok := loggers[key]; ok {
			return true
		}
	}
	return false
}

// loggerKey returns the <package>.<Type> identity of t (unwrapping a pointer or
// alias), e.g. "logrus.Logger" for *logrus.Logger. ok is false for unnamed or
// package-less types.
func loggerKey(t types.Type) (string, bool) {
	switch tt := t.(type) {
	case *types.Pointer:
		return loggerKey(tt.Elem())
	case *types.Alias:
		return loggerKey(types.Unalias(tt))
	case *types.Named:
		obj := tt.Obj()
		pkg := obj.Pkg()
		if pkg == nil {
			return "", false
		}
		return pkg.Name() + "." + obj.Name(), true
	}
	return "", false
}
