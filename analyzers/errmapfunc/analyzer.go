// Package errmapfunc implements rule GID-242: a dedicated function that
// classifies its own error parameter via errors.Is is forbidden.
//
// Owner's principle (absolute, no exceptions, and NOT specific to any one
// target — gRPC status, HTTP status, a log level, a business code, another
// error, anything): universal error-mapping/handling functions invite a
// "handle everything the same way, from everywhere" helper. Real code
// produces a bounded set of errors — map/handle that set inline, at the
// place the error is produced (the handler/interceptor), not through a
// shared function called from everywhere else that classifies an error
// parameter and branches on it.
//
// Detect: a top-level FuncDecl F such that
//   - F has a NAMED parameter of type error (e.g. err error), AND
//   - F's body calls errors.Is(<that parameter>, ...) (the standard library
//     errors.Is — chain inspection stays allowed, see GID-146) anywhere.
//
// Both hold together → F is a dedicated error classifier/mapper, reported
// on F's declaration. What F does with the classification (build a gRPC
// status, an HTTP status, a log message, another error, ...) is irrelevant —
// the forbidden shape is "a function that decides something by testing its
// OWN error parameter against sentinels," not any particular target type.
//
// Discriminator (NOT reported): inline handling inside a handler/interceptor
// method, where errors.Is branches on a LOCAL variable (the result of an
// inner call, e.g. res, err := u.Do(...)) rather than on F's own parameter.
// The key question is always: does errors.Is inspect the function's error
// PARAMETER, or a value produced inside the function body?
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
	Doc: ruleID + ": a dedicated function that classifies its own error parameter via errors.Is is forbidden " +
		"(not specific to any target — gRPC/HTTP status, logging, another error, ...). " +
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

// checkFunc reports fn if it is a dedicated error classifier/mapper:
// errors.Is is called anywhere in its body on fn's own error parameter.
func checkFunc(pass *analysis.Pass, fn *ast.FuncDecl) {
	errParams := errorParams(pass, fn)
	if len(errParams) == 0 {
		return
	}
	mapsParam := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if isErrorsIsOnParam(pass, call, errParams) {
			mapsParam = true
		}
		return true
	})
	if mapsParam {
		pass.Reportf(fn.Name.Pos(),
			"%s: a dedicated function that classifies its own error parameter via errors.Is is forbidden — "+
				"handle the bounded set of errors inline, at the call site (in the handler/interceptor where the "+
				"error occurs), whatever the target (status code, log level, another error, ...). "+
				"Fix: remove the function, inline the switch errors.Is(...) into the caller",
			ruleID)
	}
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

// isErrorsIsOnParam reports whether call is errors.Is(x, ...) (the standard
// library errors.Is) where x resolves to one of errParams.
func isErrorsIsOnParam(pass *analysis.Pass, call *ast.CallExpr, errParams map[types.Object]bool) bool {
	const stdErrorsPkgPath = "errors"

	fn := typeutil.Callee(pass.TypesInfo, call)
	f, ok := fn.(*types.Func)
	if !ok || f.Name() != "Is" {
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
