// Package enumconvert implements rule GID-143 (linter gidenumconvert):
// a map-based enum conversion must handle a missing key via
// gderror.NewUnhandledValueError.
//
// Applies only in convert packages (the last path segment is convert).
// Detected is a map indexing m[key] whose key type is a named type with
// underlying string (an enum per GID-123), and whose value type is also
// a named type (an enum→enum / enum→model-type conversion):
//
//   - the indexing is not in comma-ok form (a single assignment / expression) —
//     a missing key silently yields the zero value and cannot be handled;
//   - the comma-ok form is present, but the body of the same function has no
//     call to gderror.NewUnhandledValueError — the missing key is not handled.
//
// Maps with basic keys (string, int) are not matched. Outside convert packages
// nothing is matched. Generated code (ast.IsGenerated) is skipped.
// LoadMode — TypesInfo (the key/value types are needed).
package enumconvert

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-143"

// Analyzer — rule GID-143: a map-based enum conversion handles
// a missing key via gderror.NewUnhandledValueError.
var Analyzer = &analysis.Analyzer{
	Name: "gidenumconvert",
	Doc:  ruleID + ": enum map conversion must handle a missing key via gderror.NewUnhandledValueError. Fix: use comma-ok and return gderror.NewUnhandledValueError",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	// Scope: only convert packages (the last path segment).
	if !pathseg.EndsWith(pass.Pkg.Path(), "convert") {
		return nil, nil
	}
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

// checkFunc checks all enum map indexings in the function body.
func checkFunc(pass *analysis.Pass, fn *ast.FuncDecl) {
	hasHandler := callsUnhandledValueError(pass, fn.Body)
	commaOk := commaOkIndexes(fn.Body)
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		idx, ok := n.(*ast.IndexExpr)
		if !ok {
			return true
		}
		if !isEnumMapIndex(pass, idx) {
			return true
		}
		if _, ok := commaOk[idx]; ok {
			// the comma-ok form is present — an explicit handler call is needed in the same function.
			if !hasHandler {
				pass.Reportf(idx.Pos(),
					"%s: a missing enum-conversion key must be handled with gderror.NewUnhandledValueError",
					ruleID)
			}
			return true
		}
		// Not comma-ok: a missing key silently yields the zero value.
		pass.Reportf(idx.Pos(),
			"%s: enum conversion via map without comma-ok. "+
				"Fix: a missing key must return gderror.NewUnhandledValueError",
			ruleID)
		return true
	})
}

// isEnumMapIndex reports that idx is an indexing of a map whose key is
// a named string type (an enum) and whose value is a named type.
func isEnumMapIndex(pass *analysis.Pass, idx *ast.IndexExpr) bool {
	t := pass.TypesInfo.TypeOf(idx.X)
	if t == nil {
		return false
	}
	m, ok := t.Underlying().(*types.Map)
	if !ok {
		return false
	}
	return isNamedString(m.Key()) && isNamed(m.Elem())
}

// isNamedString: a named type with underlying string (an enum per GID-123).
func isNamedString(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	basic, ok := named.Underlying().(*types.Basic)
	return ok && basic.Kind() == types.String
}

// isNamed: a named type (enum→enum / enum→model type).
func isNamed(t types.Type) bool {
	_, ok := t.(*types.Named)
	return ok
}

// commaOkIndexes collects the map indexings used in comma-ok form
// (v, ok := m[k] / v, ok = m[k]) — an RHS of one expression with two LHS.
func commaOkIndexes(body *ast.BlockStmt) map[*ast.IndexExpr]struct{} {
	out := map[*ast.IndexExpr]struct{}{}
	ast.Inspect(body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok || len(assign.Lhs) != 2 || len(assign.Rhs) != 1 {
			return true
		}
		if idx, ok := assign.Rhs[0].(*ast.IndexExpr); ok {
			out[idx] = struct{}{}
		}
		return true
	})
	return out
}

// callsUnhandledValueError reports that the body contains a call to
// NewUnhandledValueError. Matched by symbol name only — the import path of
// the errors helper library is not pinned (the same constructor lives under
// different module paths: gitlab.gid.team/.../helper.git/errors,
// git.k8s.nomilk.space/go-library/ehelper, …).
func callsUnhandledValueError(pass *analysis.Pass, body *ast.BlockStmt) bool {
	// unhandledCtor — the constructor for handling a missing key.
	const unhandledCtor = "NewUnhandledValueError"
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		fn, ok := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
		if !ok || fn.Pkg() == nil {
			return true
		}
		if fn.Name() == unhandledCtor {
			found = true
			return false
		}
		return true
	})
	return found
}
