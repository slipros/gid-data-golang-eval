// Package errmapfunc implements rule GID-242: a dedicated error-MAPPER
// function — one that classifies its own error parameter via errors.Is and
// returns an error (maps error → error/status) — is forbidden.
//
// Owner's principle (absolute, no exceptions, and NOT specific to any one
// error-return target — a mapped error, a gRPC status error, an HTTP error,
// anything the function returns as its error result): a shared error-mapper
// translates errors from layer to layer and gets called from everywhere.
// Real code produces a bounded set of errors — map that set inline, at the
// place the error is produced (the handler/interceptor), not through a
// dedicated mapper function.
//
// Detect: a top-level FuncDecl F such that ALL of
//   - F has a NAMED parameter of type error (e.g. err error), AND
//   - F's body CLASSIFIES that parameter, in either of two shapes:
//     (a) errors.Is(<that parameter>, ...) / errors.As(<that parameter>,
//     &target) — where errors is any of the configured classifier packages
//     (settings.packages; default: the standard library "errors" and
//     github.com/pkg/errors, which forwards Is/As to stdlib since v0.9.1;
//     gid.team code uses pkg/errors, GID-146) — with that parameter as the
//     first argument, anywhere (chain inspection / type-assert stays
//     allowed, see GID-146);
//     (b) a call to ANY bool-predicate over an error — a function or method
//     whose signature is func(error, ...) bool — with that parameter as the
//     first argument: gdpostgres.IsUniqueViolation(err), IsNoResult(err),
//     s.isRetryable(err). A driver library exports its classification as
//     such predicates rather than through errors.Is sentinels, so shape (a)
//     alone let a real mapper through (incident 2026-08-04, resource-registry
//     repository.MapError). Predicate calls are matched by SIGNATURE, in any
//     package — including this one, AND
//   - F's result list includes error (F returns error, or (T, error), ...), AND
//   - F actually PRODUCES an error of its own (discriminator #3, see below).
//
// settings.packages lets a project add its own errors-facade package paths
// (e.g. an internal errors wrapper that re-exports Is/As) without a code
// change; when empty, defaultPackages is used. It governs shape (a) only —
// shape (b) is signature-based and needs no package list.
//
// All of them hold together → F is a dedicated error mapper, reported on F's
// declaration.
//
// Discriminator #1 — the RETURN type (added 2026-07-12, owner refinement):
// only functions that RETURN error are mappers. A bool-predicate over the
// error parameter (func isNotFound(err error) bool { return errors.Is(...) },
// func isRetryable(err error) bool, func isCustom(err error) bool { var t
// *CustomErr; return errors.As(err, &t) }) is a legitimate classifier/helper,
// not a mapper, and is NOT reported — it does not translate the error into a
// new error/status, it merely answers a yes/no question about it.
//
// Discriminator #2 — the PARAMETER vs a local: inline handling inside a
// handler/interceptor method, where errors.Is branches on a LOCAL variable
// (the result of an inner call, e.g. res, err := u.Do(...)) rather than on
// F's own parameter, is NOT reported. The question is always: does errors.Is
// inspect the function's error PARAMETER, or a value produced inside the body?
//
// Discriminator #3 — MAPPING vs merely observing (added 2026-08-04 alongside
// shape (b)): a mapper replaces the error. F is reported only when it either
// assigns to its own error parameter (err = entity.ErrNoResult) or returns,
// in some branch, an error expression that is NOT that bare parameter
// (errors.WithStack(ErrX), status.Error(...), nil). A function that classifies
// the error only to decide how to LOG or count it and always hands the very
// same value back (if isTemporary(err) { log.Warn() }; return err) produces no
// error of its own — it is not a mapper and is not reported. Without this
// discriminator, shape (b) would report such observers.
//
// Generated code (ast.IsGenerated) is skipped.
package errmapfunc

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

const ruleID = "GID-242"

// defaultPackages — the errors-classifier packages whose Is/As calls are
// recognized: the standard library and github.com/pkg/errors (which forwards
// Is/As to stdlib since v0.9.1). gid.team code uses pkg/errors exclusively
// (GID-146). A project can replace this list via settings.packages.
var defaultPackages = []string{
	"errors",
	"github.com/pkg/errors",
}

// Analyzer — rule GID-242 with the default classifier-package list.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Packages — errors-classifier package import paths whose Is/As calls
	// count. Replaces the default list (stdlib "errors" + github.com/pkg/errors).
	Packages []string `json:"packages"`
}

// NewAnalyzer builds the GID-242 analyzer from linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	pkgs := s.Packages
	if len(pkgs) == 0 {
		pkgs = defaultPackages
	}
	classifierPkgs := make(map[string]bool, len(pkgs))
	for _, p := range pkgs {
		classifierPkgs[p] = true
	}
	return &analysis.Analyzer{
		Name: "giderrmapfunc",
		Doc: ruleID + ": a dedicated error-mapper function (classifies its own error parameter via errors.Is/errors.As " +
			"AND returns error — maps error to error/status) is forbidden; bool-predicates and wrappers are fine. " +
			"Fix: remove the function, inline the switch errors.Is(...) into the caller",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, classifierPkgs)
		},
	}
}

func run(pass *analysis.Pass, classifierPkgs map[string]bool) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			checkFunc(pass, fn, classifierPkgs)
		}
	}
	return nil, nil
}

// checkFunc reports fn if it is a dedicated error mapper: it has an error
// parameter, its body calls errors.Is/As (from a classifier package) on that
// parameter, AND it returns error. A function that does not return error
// (e.g. a bool predicate) is a legitimate classifier/helper and is not reported.
func checkFunc(pass *analysis.Pass, fn *ast.FuncDecl, classifierPkgs map[string]bool) {
	errParams := errorParams(pass, fn)
	if len(errParams) == 0 {
		return
	}
	if !funcReturnsError(pass, fn) {
		return
	}
	classifies := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if isErrorsClassifyOnParam(pass, call, errParams, classifierPkgs) ||
			isPredicateOnParam(pass, call, errParams) {
			classifies = true
		}
		return true
	})
	if classifies && producesOwnError(pass, fn, errParams) {
		pass.Reportf(fn.Name.Pos(),
			"%s: a dedicated error-mapper function is forbidden — it classifies its own error parameter "+
				"(errors.Is/errors.As, or a bool-predicate such as IsNoResult(err)) and returns an error of its "+
				"own (maps error to error/status). Map the bounded set of errors inline, at the call site (in the "+
				"repository method/handler where the error occurs): "+
				"if IsNoResult(err) { err = entity.ErrNoResult }; return errors.Wrap(err, \"select x\"). "+
				"A bool-predicate (func isNotFound(err error) bool) is a legitimate classifier, not a mapper",
			ruleID)
	}
}

// isPredicateOnParam reports whether call is a bool-predicate over one of
// errParams: a call to a function or method whose signature is
// func(error, ...) bool, with that parameter as the FIRST argument
// (gdpostgres.IsUniqueViolation(err), IsNoResult(err), s.isRetryable(err)).
// Matching is by SIGNATURE on the resolved callee, so any package counts —
// a driver library publishes its error classification exactly this way, and
// the errors.Is/As shape alone never sees it.
func isPredicateOnParam(pass *analysis.Pass, call *ast.CallExpr, errParams map[types.Object]bool) bool {
	fn := typeutil.Callee(pass.TypesInfo, call)
	f, ok := fn.(*types.Func)
	if !ok {
		return false
	}
	sig, ok := f.Type().(*types.Signature)
	if !ok || !isErrorBoolPredicate(sig) {
		return false
	}
	if len(call.Args) == 0 {
		return false
	}
	return paramObject(pass, call.Args[0], errParams) != nil
}

// isErrorBoolPredicate reports whether sig is func(error, ...) bool — the
// first parameter is an error and the only result is a bool.
func isErrorBoolPredicate(sig *types.Signature) bool {
	params := sig.Params()
	results := sig.Results()
	if params.Len() == 0 || results.Len() != 1 {
		return false
	}
	firstParam := params.At(0)
	if !isErrorType(firstParam.Type()) {
		return false
	}
	result := results.At(0)
	resultType := result.Type()
	basic, ok := resultType.Underlying().(*types.Basic)
	return ok && basic.Kind() == types.Bool
}

// producesOwnError reports whether fn replaces the error instead of merely
// observing it (discriminator #3): it assigns to its own error parameter, or
// some return hands back an error expression that is not that bare parameter.
// A classifier that always returns the very same value it received (logging,
// metrics) maps nothing and is left alone.
func producesOwnError(pass *analysis.Pass, fn *ast.FuncDecl, errParams map[types.Object]bool) bool {
	produces := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range stmt.Lhs {
				if paramObject(pass, lhs, errParams) != nil {
					produces = true
				}
			}
		case *ast.ReturnStmt:
			for _, res := range stmt.Results {
				if !isErrorType(pass.TypesInfo.TypeOf(res)) {
					continue
				}
				if paramObject(pass, res, errParams) == nil {
					produces = true
				}
			}
		}
		return true
	})
	return produces
}

// paramObject returns the object expr refers to when it is a plain identifier
// resolving to one of params; nil otherwise. Both Uses (a read) and Defs
// (the declaration itself) are consulted so an assignment target resolves too.
func paramObject(pass *analysis.Pass, expr ast.Expr, params map[types.Object]bool) types.Object {
	id, ok := expr.(*ast.Ident)
	if !ok {
		return nil
	}
	obj := pass.TypesInfo.Uses[id]
	if obj == nil {
		obj = pass.TypesInfo.Defs[id]
	}
	if obj == nil || !params[obj] {
		return nil
	}
	return obj
}

// funcReturnsError reports whether fn's result list includes an error.
func funcReturnsError(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	if fn.Type.Results == nil {
		return false
	}
	for _, field := range fn.Type.Results.List {
		if isErrorType(pass.TypesInfo.TypeOf(field.Type)) {
			return true
		}
	}
	return false
}

// errorParams collects the objects of fn's NAMED parameters of type error.
// An unnamed error parameter (just "error" in the signature) is not in
// scope: it cannot be referenced by errors.Is inside the body at all.
func errorParams(pass *analysis.Pass, fn *ast.FuncDecl) map[types.Object]bool {
	out := map[types.Object]bool{}
	if fn.Type.Params == nil {
		return out
	}
	for _, field := range fn.Type.Params.List {
		if len(field.Names) == 0 {
			continue
		}
		if !isErrorType(pass.TypesInfo.TypeOf(field.Type)) {
			continue
		}
		for _, name := range field.Names {
			if obj := pass.TypesInfo.Defs[name]; obj != nil {
				out[obj] = true
			}
		}
	}
	return out
}

// isErrorsClassifyOnParam reports whether call is errors.Is(x, ...) or
// errors.As(x, ...) — where errors is any of classifierPkgs — and x, the
// first argument, resolves to one of errParams. Matching is done on the
// RESOLVED callee package (typeutil.Callee), so a source-level import alias
// (stderrors "errors", pkgerrors "github.com/pkg/errors") is handled
// automatically. The default classifierPkgs cover the standard library and
// github.com/pkg/errors; settings.packages replaces them.
func isErrorsClassifyOnParam(
	pass *analysis.Pass, call *ast.CallExpr, errParams map[types.Object]bool, classifierPkgs map[string]bool,
) bool {
	fn := typeutil.Callee(pass.TypesInfo, call)
	f, ok := fn.(*types.Func)
	if !ok || (f.Name() != "Is" && f.Name() != "As") {
		return false
	}
	pkg := f.Pkg()
	if pkg == nil || !classifierPkgs[pkg.Path()] {
		return false
	}
	if len(call.Args) == 0 {
		return false
	}
	id, ok := call.Args[0].(*ast.Ident)
	if !ok {
		return false
	}
	obj := pass.TypesInfo.Uses[id]
	return obj != nil && errParams[obj]
}

func isErrorType(t types.Type) bool {
	if t == nil {
		return false
	}
	errObj := types.Universe.Lookup("error")
	errType := errObj.Type()
	iface, ok := errType.Underlying().(*types.Interface)
	if !ok {
		return false
	}
	return types.Implements(t, iface)
}
