// Package errmapfunc implements rule GID-242: a dedicated error-MAPPER
// function — one that classifies its own error parameter via errors.Is and
// turns it into something else (an error, a gRPC status, a status code, a
// message) — is forbidden.
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
//   - F RETURNS something, and that something is NOT a lone bool
//     (discriminator #1, see below), AND
//   - F actually PRODUCES a value of its own (discriminator #3, see below).
//
// settings.packages lets a project add its own errors-facade package paths
// (e.g. an internal errors wrapper that re-exports Is/As) without a code
// change; when empty, defaultPackages is used. It governs shape (a) only —
// shape (b) is signature-based and needs no package list.
//
// All of them hold together → F is a dedicated error mapper, reported on F's
// declaration.
//
// Discriminator #1 — the RETURN type (added 2026-07-12, widened 2026-08-06):
// the ONLY legitimate shape is a bool-predicate — a function whose single
// result is a bool (func isNotFound(err error) bool { return errors.Is(...) },
// func isRetryable(err error) bool, func isCustom(err error) bool { var t
// *CustomErr; return errors.As(err, &t) }). It does not translate the error
// into anything, it merely answers a yes/no question about it. Everything else
// a classifier can hand back IS the translation, whatever its Go type:
// error, *status.Status, codes.Code, an HTTP status int, a message string.
// The first cut of this discriminator asked "does F return error?", which read
// the mapper's OUTPUT TYPE as the thing being ruled on — but the rule is about
// the error being translated away from its origin, and gRPC transport types
// carry it just as well. That gap shipped a textbook mapper split across two
// functions (incident 2026-08-06, resource-registry internal/server/grpc/errors:
// func Code(err error) codes.Code + func Converter(err error) *status.Status,
// both classifying err through the package's own IsNotFound/IsAlreadyExists
// predicates, both invisible to an error-return-only detector).
// A function with NO results is not a mapper either — it produces nothing to
// map into (a classifier that only logs or counts).
//
// Discriminator #2 — the PARAMETER vs a local: inline handling inside a
// handler/interceptor method, where errors.Is branches on a LOCAL variable
// (the result of an inner call, e.g. res, err := u.Do(...)) rather than on
// F's own parameter, is NOT reported. The question is always: does errors.Is
// inspect the function's error PARAMETER, or a value produced inside the body?
//
// Discriminator #3 — MAPPING vs merely observing (added 2026-08-04 alongside
// shape (b)): a mapper replaces the error. F is reported only when it either
// assigns to its own error parameter (err = entity.ErrNoResult), assigns to a
// NAMED result (out = ErrX; return), or returns, in some branch, an expression
// that is NOT that bare parameter (errors.WithStack(ErrX), status.Error(...),
// codes.NotFound, nil). A function that classifies the error only to decide
// how to LOG or count it and always hands the very same value back
// (if isTemporary(err) { log.Warn() }; return err) produces nothing of its own
// — it is not a mapper and is not reported. Without this discriminator, shape
// (b) would report such observers.
//
// When F returns error, only its ERROR results are weighed — an observer with
// a (T, error) signature legitimately returns a zero T alongside the untouched
// err. When F returns no error at all (the codes.Code / *status.Status mapper),
// every result counts: nothing it hands back can be the error parameter itself,
// so any non-empty return is a value of its own.
//
// Generated code (ast.IsGenerated) is skipped.
//
// settings.exclude ("Function" | "Type.Method", the same centralized-exception
// mechanism as giderrtext/gidmapout — see internal/exclude): some frameworks
// DICTATE the signature of the one place error translation is allowed to live,
// leaving no call site to inline into. gdgrpcserver.WithErrorConverters (the
// gid.team gRPC server library, gitlab.gid.team/gid-data/tech/golang/libs/grpc)
// takes interceptor.ErrorConverterFunc = func(error) *status.Status — the
// registered function's signature is fixed by the library, not chosen by the
// service. A validator-result-to-gRPC-status converter registered there
// (ValidationErrorConverter: errors.As(err, &result) to classify, then
// status.FromError(...) to translate) has no caller-side switch to fold the
// mapping into — the interceptor calls the registered func(error) *status.Status
// directly. That shape is the rule's canonical false positive (incident
// 2026-08-06, resource-registry internal/server/grpc/integration and
// advertising-api internal/server/grpc/advertising, both named
// ValidationErrorConverter). The exclusion is per function/method, same as
// every other centralized exception in this repo — it does not touch
// discriminators #1-#3, so a domain mapper (func Code(err error) codes.Code)
// living anywhere else keeps getting caught.
package errmapfunc

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
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
	// Exclude — functions/methods that are the one framework-mandated place
	// error translation is allowed to live (e.g. a gdgrpcserver error
	// converter, whose signature — func(error) *status.Status — is fixed by
	// interceptor.ErrorConverterFunc, not chosen by the service): "Function"
	// (any receiver) or "Type.Method". See the package doc for why this
	// exists.
	Exclude []string `json:"exclude"`
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
			"or a bool-predicate AND returns anything but a lone bool — error, *status.Status, codes.Code, a message) " +
			"is forbidden; bool-predicates and wrappers are fine. " +
			"Fix: remove the function, inline the switch errors.Is(...) into the caller",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, classifierPkgs, s.Exclude)
		},
	}
}

func run(pass *analysis.Pass, classifierPkgs map[string]bool, excluded []string) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if exclude.Match(excluded, receiverType(fn), fn.Name.Name) {
				continue
			}
			checkFunc(pass, fn, classifierPkgs)
		}
	}
	return nil, nil
}

// checkFunc reports fn if it is a dedicated error mapper: it has an error
// parameter, its body classifies that parameter (errors.Is/As from a
// classifier package, or a bool-predicate), AND it hands back a translation of
// it — anything other than a lone bool. A bool-predicate answers a yes/no
// question about the error instead of translating it, and a function with no
// results translates it into nothing at all; both are legitimate.
func checkFunc(pass *analysis.Pass, fn *ast.FuncDecl, classifierPkgs map[string]bool) {
	errParams := errorParams(pass, fn)
	if len(errParams) == 0 {
		return
	}
	if !funcReturnsTranslation(pass, fn) {
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
	if classifies && producesOwnValue(pass, fn, errParams) {
		pass.Reportf(fn.Name.Pos(),
			"%s: a dedicated error-mapper function is forbidden — it classifies its own error parameter "+
				"(errors.Is/errors.As, or a bool-predicate such as IsNoResult(err)) and hands back a translation "+
				"of it: an error, a *status.Status, a codes.Code, a message. Map the bounded set of errors inline, "+
				"at the call site (in the repository method/handler where the error occurs): "+
				"if IsNoResult(err) { err = entity.ErrNoResult }; return errors.Wrap(err, \"select x\"). "+
				"Only a bool-predicate (func isNotFound(err error) bool) is a legitimate classifier, not a mapper",
			ruleID)
	}
}

// funcReturnsTranslation reports whether fn hands back a translation of the
// error it classifies — discriminator #1. Every result shape counts (error,
// *status.Status, codes.Code, an HTTP status int, a message string) except the
// two that translate nothing: a lone bool (a predicate answers a yes/no
// question about the error) and no results at all (a classifier that only logs
// or counts).
func funcReturnsTranslation(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
		return false
	}
	return !isBoolPredicate(pass, fn)
}

// isBoolPredicate reports whether fn's result list is exactly one unnamed-or-
// named bool: func isNotFound(err error) bool.
func isBoolPredicate(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	fields := fn.Type.Results.List
	if len(fields) != 1 || len(fields[0].Names) > 1 {
		return false
	}
	resultType := pass.TypesInfo.TypeOf(fields[0].Type)
	if resultType == nil {
		return false
	}
	basic, ok := resultType.Underlying().(*types.Basic)
	return ok && basic.Kind() == types.Bool
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

// producesOwnValue reports whether fn replaces the error instead of merely
// observing it (discriminator #3): it assigns to its own error parameter or to
// a named result, or some return hands back an expression that is not that
// bare parameter. A classifier that always returns the very same value it
// received (logging, metrics) maps nothing and is left alone.
//
// When fn returns error, only its error results are weighed — an observer with
// a (T, error) signature returns a zero T next to the untouched err, and that
// zero T must not read as a translation. When fn returns no error at all, its
// results cannot BE the error parameter, so every one of them counts.
func producesOwnValue(pass *analysis.Pass, fn *ast.FuncDecl, errParams map[types.Object]bool) bool {
	errorResultsOnly := funcReturnsError(pass, fn)
	targets := map[types.Object]bool{}
	for obj := range errParams {
		targets[obj] = true
	}
	for obj := range namedResults(pass, fn, errorResultsOnly) {
		targets[obj] = true
	}
	produces := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range stmt.Lhs {
				if paramObject(pass, lhs, targets) != nil {
					produces = true
				}
			}
		case *ast.ReturnStmt:
			for _, res := range stmt.Results {
				if errorResultsOnly && !isErrorType(pass.TypesInfo.TypeOf(res)) {
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

// namedResults collects the objects of fn's NAMED results, so that a mapper
// written as `func code(err error) (c codes.Code) { if isX(err) { c = … };
// return }` — which never returns an expression at all — still counts as
// producing a value of its own. errorsOnly restricts the set to error results,
// mirroring producesOwnValue: when fn returns error, filling in a non-error
// result (a zero T next to the untouched err) is not a translation.
func namedResults(pass *analysis.Pass, fn *ast.FuncDecl, errorsOnly bool) map[types.Object]bool {
	out := map[types.Object]bool{}
	if fn.Type.Results == nil {
		return out
	}
	for _, field := range fn.Type.Results.List {
		if errorsOnly && !isErrorType(pass.TypesInfo.TypeOf(field.Type)) {
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

// receiverType returns the name of fn's receiver type ("Client" for both Client
// and *Client), or "" for a plain function — the form settings.exclude matches.
func receiverType(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}
	expr := fn.Recv.List[0].Type
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if id, ok := expr.(*ast.Ident); ok {
		return id.Name
	}
	return ""
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
