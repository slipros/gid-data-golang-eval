// Package eventctor implements rule GID-216 (event-ctor-deps): dependencies
// of event-layer constructors (source: event.md).
//
//   - GID-216 (gideventctor): in the consumer layer the constructor must take
//     a logger (*logrus.Logger or *logrus.Entry) — the consumer builds an Entry
//     with broker/consumer fields; in the producer layer the constructor takes
//     no logger — errors are propagated to the caller.
//
// Scope is determined by the import path segments: consumer — a package with
// segments event and consumer, producer — with segments event and producer.
// Subpackages with segments validate and convert are excluded: they hold
// validators and converters, not consumers/producers.
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

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/lgr"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-216"

// Check modes for an event-layer package (the scope type is declared below).
const (
	scopeNone scope = iota
	scopeConsumer
	scopeProducer
)

// ctorName — a constructor: an exported name of the form New + a capital letter.
var ctorName = regexp.MustCompile(`^New[A-Z]`)

// Analyzer — variant with default settings (no exclusions).
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — names of excluded constructors (for example "NewOrderConsumer").
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-216 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(cfg Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gideventctor",
		Doc:  ruleID + ": consumer constructors take a logrus logger, producer constructors do not. Fix: add *logrus.Logger to consumer constructors, remove it from producers",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, cfg)
		},
	}
}

// scope — the check mode for an event-layer package.
type scope int

func pkgScope(pkgPath string) scope {
	if !pathseg.Contains(pkgPath, "event") {
		return scopeNone
	}
	// validate/convert hold validators and converters, not consumers/producers.
	if pathseg.Contains(pkgPath, "validate") || pathseg.Contains(pkgPath, "convert") {
		return scopeNone
	}
	switch {
	case pathseg.Contains(pkgPath, "consumer"):
		return scopeConsumer
	case pathseg.Contains(pkgPath, "producer"):
		return scopeProducer
	default:
		return scopeNone
	}
}

func run(pass *analysis.Pass, cfg Settings) (any, error) {
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
			check(pass, sc, fn, sig)
		}
	}
	return nil, nil
}

func check(pass *analysis.Pass, sc scope, fn *ast.FuncDecl, sig *types.Signature) {
	has := hasLoggerParam(sig)
	switch sc {
	case scopeConsumer:
		if !has {
			pass.Reportf(fn.Name.Pos(),
				"%s: a consumer constructor must take *logrus.Logger and build an Entry with broker/consumer fields "+
					"(see event.md). Fix: add a logger *logrus.Logger parameter and build the Entry with WithField "+
					"in the constructor",
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

// hasLoggerParam reports whether any of the parameters has a logrus type
// (*logrus.Logger or *logrus.Entry).
func hasLoggerParam(sig *types.Signature) bool {
	params := sig.Params()
	for i := 0; i < params.Len(); i++ {
		param := params.At(i)
		if lgr.IsType(param.Type()) {
			return true
		}
	}
	return false
}
