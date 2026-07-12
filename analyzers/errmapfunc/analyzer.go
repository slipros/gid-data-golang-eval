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
//   - F's body calls errors.Is(<that parameter>, ...) OR
//     errors.As(<that parameter>, &target) (the standard library errors.Is /
//     errors.As — chain inspection / type-assert stays allowed, see GID-146)
//     with that parameter as the first argument, anywhere, AND
//   - F's result list includes error (F returns error, or (T, error), ...).
//
// All three hold together → F is a dedicated error mapper, reported on F's
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
// Generated code (ast.IsGenerated) is skipped.
package errmapfunc

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

const ruleID = "GID-242"

// Analyzer — rule GID-242 (no settings: the rule is absolute, no exceptions).
var Analyzer = &analysis.Analyzer{
	Name: "giderrmapfunc",
	Doc: ruleID + ": a dedicated error-mapper function (classifies its own error parameter via errors.Is " +
		"AND returns error — maps error to error/status) is forbidden; bool-predicates and wrappers are fine. " +
		"Fix: remove the function, inline the switch errors.Is(...) into the caller",
	Run: run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			checkFunc(pass, fn)
		}
	}
	return nil, nil
}

// checkFunc reports fn if it is a dedicated error mapper: it has an error
// parameter, its body calls errors.Is on that parameter, AND it returns
// error. A function that does not return error (e.g. a bool predicate) is a
// legitimate classifier/helper and is not reported.
func checkFunc(pass *analysis.Pass, fn *ast.FuncDecl) {
	errParams := errorParams(pass, fn)
	if len(errParams) == 0 {
		return
	}
	if !funcReturnsError(pass, fn) {
		return
	}
	mapsParam := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if isErrorsClassifyOnParam(pass, call, errParams) {
			mapsParam = true
		}
		return true
	})
	if mapsParam {
		pass.Reportf(fn.Name.Pos(),
			"%s: a dedicated error-mapper function is forbidden — it classifies its own error parameter via "+
				"errors.Is/errors.As and returns error (maps error to error/status). Map the bounded set of errors "+
				"inline, at the call site (in the handler/interceptor where the error occurs). A bool-predicate "+
				"(func isNotFound(err error) bool) is a legitimate classifier, not a mapper. "+
				"Fix: remove the function, inline the switch errors.Is(...) into the caller",
			ruleID)
	}
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
// errors.As(x, ...) (the standard library errors package) where x — the
// first argument — resolves to one of errParams.
func isErrorsClassifyOnParam(pass *analysis.Pass, call *ast.CallExpr, errParams map[types.Object]bool) bool {
	const stdErrorsPkgPath = "errors"

	fn := typeutil.Callee(pass.TypesInfo, call)
	f, ok := fn.(*types.Func)
	if !ok || (f.Name() != "Is" && f.Name() != "As") {
		return false
	}
	pkg := f.Pkg()
	if pkg == nil || pkg.Path() != stdErrorsPkgPath {
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
