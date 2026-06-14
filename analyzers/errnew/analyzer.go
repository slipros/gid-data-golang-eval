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

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

const ruleID = "GID-136"

// Analyzer — rule GID-136: errors.New (pkg/errors) only in a package-level var.
var Analyzer = &analysis.Analyzer{
	Name: "giderrnew",
	Doc:  ruleID + ": errors.New (pkg/errors) only in a package-level var, not at runtime. Fix: declare a package-level var ErrX",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		checkFile(pass, file)
	}
	return nil, nil
}

// checkFile walks the bodies of all functions, methods and func literals in
// the file and reports calls to errors.New from pkg/errors inside them. Calls
// outside function bodies (package-level var ErrX = errors.New(...)) are untouched.
//
// A func literal body is runtime even when the literal itself is assigned to a
// package-level var: errors.New there is evaluated when the literal is called.
func checkFile(pass *analysis.Pass, file *ast.File) {
	var bodies []*ast.BlockStmt

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Body != nil {
				bodies = append(bodies, node.Body)
			}
		case *ast.FuncLit:
			bodies = append(bodies, node.Body)
		}
		return true
	})

	for _, body := range bodies {
		ast.Inspect(body, func(n ast.Node) bool {
			// Do not descend into a nested func literal — its body is walked
			// in a separate iteration, otherwise the call would be reported twice.
			if _, ok := n.(*ast.FuncLit); ok {
				return false
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if isPkgErrorsNew(pass, call) {
				pass.Reportf(call.Pos(),
					"%s: errors.New at runtime. Fix: declare a package-level var ErrX (see GID-169: error.go)",
					ruleID)
			}
			return true
		})
	}
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
