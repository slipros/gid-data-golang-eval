// Package errnew implements rule GID-136 (linter giderrnew):
// errors.New from github.com/pkg/errors is allowed only in a
// package-level var declaration — static errors are declared up front
// (ErrX), not constructed at runtime.
//
// A call to errors.New inside the body of a function, method or
// func literal is a diagnostic. A package-level var declaration
// (including var blocks) ErrX = errors.New("...") is the norm.
//
// Out of scope:
//   - errors.Errorf — dynamic context is legitimate; its placement
//     is governed by GID-144/GID-145;
//   - the standard errors.New — it is already forbidden by GID-146;
//   - errors.New from any other (non github.com/pkg/errors) package.
//
// pkg/errors is detected by the import path github.com/pkg/errors via
// TypesInfo. Generated code (ast.IsGenerated) is skipped.
package errnew

import (
	"go/ast"
	"go/types"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

const ruleID = "GID-136"

// Analyzer — rule GID-136: errors.New (pkg/errors) only in a package-level var.
var Analyzer = &analysis.Analyzer{
	Name:     "giderrnew",
	Doc:      ruleID + ": errors.New (pkg/errors) only in a package-level var, not at runtime. Fix: declare a package-level var ErrX",
	Requires: astwalk.Requires,
	Run:      run,
}

// bodyFilter — the function shapes that open a runtime scope, the block that
// is their body, and the calls being judged.
var bodyFilter = []ast.Node{
	(*ast.FuncDecl)(nil),
	(*ast.FuncLit)(nil),
	(*ast.BlockStmt)(nil),
	(*ast.CallExpr)(nil),
}

// run reports calls to errors.New from pkg/errors inside the body of a
// function, method or func literal. Calls outside function bodies
// (package-level var ErrX = errors.New(...)) are untouched.
//
// A func literal body is runtime even when the literal itself is assigned to a
// package-level var: errors.New there is evaluated when the literal is called.
func run(pass *analysis.Pass) (any, error) {
	skip := func(file *ast.File) bool {
		return ast.IsGenerated(file) || srcfile.IsTest(pass, file)
	}

	// depth counts the function bodies open around the traversal; the signature
	// of a function is not its body, so only the body block opens a level.
	var (
		depth  int
		bodies = map[*ast.BlockStmt]struct{}{}
	)

	astwalk.Around(pass, bodyFilter, skip, func(_ *ast.File, n ast.Node, push bool) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if push && node.Body != nil {
				bodies[node.Body] = struct{}{}
			}
		case *ast.FuncLit:
			if push {
				bodies[node.Body] = struct{}{}
			}
		case *ast.BlockStmt:
			if _, ok := bodies[node]; ok {
				if push {
					depth++
				} else {
					depth--
				}
			}
		case *ast.CallExpr:
			if push && depth > 0 && isPkgErrorsNew(pass, node) {
				pass.Reportf(node.Pos(),
					"%s: errors.New at runtime. Fix: declare a package-level var ErrX (see GID-169: error.go)",
					ruleID)
			}
		}

		return true
	})

	return nil, nil
}

// isPkgErrorsNew reports whether call is a call to errors.New from
// github.com/pkg/errors.
func isPkgErrorsNew(pass *analysis.Pass, call *ast.CallExpr) bool {
	const pkgErrorsPath = "github.com/pkg/errors"
	fn := typeutil.Callee(pass.TypesInfo, call)
	f, ok := fn.(*types.Func)
	if !ok || f.Pkg() == nil {
		return false
	}
	pkg := f.Pkg()
	return pkg.Path() == pkgErrorsPath && f.Name() == "New"
}
