// Package errlast implements rule GID-190: Google's conventions on errors
// in the results of functions and methods.
//
// Check 1 (error is not last). If among the results there is an error type
// and other results follow it — this is a violation: error must be the last
// returned value.
//
// Check 2 (a concrete error type in a result). A function result is a
// concrete type implementing error (a named type or a pointer to it,
// e.g. *MyError / MyError), not the error interface. A concrete type in an
// interface position causes the classic typed-nil trap: a returned nil pointer
// in a variable of type error != nil. The error interface should be returned.
//
// NOT matched by check 2:
//   - the error interface itself;
//   - interface types extending error (a custom error interface — a deliberate
//     decision by the author);
//   - error-constructor functions in error.go / errors.go files — there a
//     concrete type is legitimate (these are constructors like NewMyError() *MyError).
//
// Generated code (ast.IsGenerated) is skipped. LoadMode is TypesInfo.
package errlast

import (
	"go/ast"
	"go/types"
	"path/filepath"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-190"

// errorFiles — files in which private error-constructor functions
// may return a concrete error type (check 2 is not applied).
var errorFiles = map[string]bool{
	"error.go":  true,
	"errors.go": true,
}

// Analyzer — rule GID-190: error is the last result, concrete error types are not returned.
var Analyzer = &analysis.Analyzer{
	Name: "giderrlast",
	Doc:  ruleID + ": error must be the last result, and the error interface (not a concrete type) is returned. Fix: move error last and return the error interface",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	errObj := types.Universe.Lookup("error")
	errType := errObj.Type()
	errIface, ok := errType.Underlying().(*types.Interface)
	if !ok {
		return nil, nil
	}

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		tokenFile := pass.Fset.File(file.Pos())
		inErrorFile := errorFiles[filepath.Base(tokenFile.Name())]

		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}
			checkResults(pass, fn, errIface, inErrorFile)
			return true
		})
	}
	return nil, nil
}

// checkResults applies both checks to the results of a function/method.
func checkResults(pass *analysis.Pass, fn *ast.FuncDecl, errIface *types.Interface, inErrorFile bool) {
	if fn.Type.Results == nil {
		return
	}

	// Expand group results into a flat list (type, expression).
	type result struct {
		expr ast.Expr
		typ  types.Type
	}
	var results []*result
	for _, field := range fn.Type.Results.List {
		t := pass.TypesInfo.TypeOf(field.Type)
		if t == nil {
			continue
		}
		// Several result names of one type: (a, b int) — each is a separate result.
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for range count {
			results = append(results, &result{expr: field.Type, typ: t})
		}
	}
	if len(results) == 0 {
		return
	}

	for i, r := range results {
		// Check 1: error is not last — there are results after it.
		if isExactError(r.typ) && i != len(results)-1 {
			pass.Reportf(r.expr.Pos(),
				"%s: error must be the last return value. Fix: move it to the end", ruleID)
			continue
		}

		// Check 2: a concrete error type in a result.
		if inErrorFile {
			continue // error constructors in error.go/errors.go legitimately return a concrete type
		}
		if isConcreteError(r.typ, errIface) {
			pass.Reportf(r.expr.Pos(),
				"%s: return the error interface, not %s. Fix: a concrete type in the error position causes a typed-nil trap",
				ruleID, r.typ.String())
		}
	}
}

// isExactError reports whether the type is exactly the error interface.
func isExactError(t types.Type) bool {
	named, ok := t.(*types.Named)
	if ok {
		obj := named.Obj()
		return obj != nil && obj.Pkg() == nil && obj.Name() == "error"
	}
	return false
}

// isConcreteError reports whether the type implements error while being a
// concrete (non-interface) type — named or a pointer to a named one.
// Interfaces (including error itself and custom error interfaces) are excluded.
func isConcreteError(t types.Type, errIface *types.Interface) bool {
	// An interface type (error and its extensions) is not concrete.
	if _, isIface := t.Underlying().(*types.Interface); isIface {
		return false
	}

	// Only named types and pointers to named ones are of interest.
	switch u := t.(type) {
	case *types.Named:
		// ok
	case *types.Pointer:
		if _, ok := u.Elem().(*types.Named); !ok {
			return false
		}
	default:
		return false
	}

	return types.Implements(t, errIface) || types.Implements(types.NewPointer(t), errIface)
}
