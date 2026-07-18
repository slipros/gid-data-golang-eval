// Package loggernew implements rule GID-214 (logger-singleton):
//
//   - GID-214 (gidloggernew): the logger is created once in the composition root.
//     Calls to logrus.New() and logrus.StandardLogger() (package
//     github.com/sirupsen/logrus) are forbidden everywhere except the main
//     package and composition-root packages (path contains the internal/app
//     segments).
//
// A ready *logrus.Entry is passed through the constructor rather than created
// anew in service/repository — otherwise the unified logger configuration
// (format, hooks, level) and cross-cutting fields are lost.
//
// _test.go files and generated files are skipped: a logger in tests is fine,
// and generated code is not edited by hand.
//
// logrus is resolved via types (import path), so a call to New() from another
// package with the same name is not flagged.
//
// LoadMode: TypesInfo — resolving the called function's package by import path
// is required.
//
// Source: libs.md (logrus: do not create new instances, pass the existing one).
package loggernew

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-214"

// bannedFuncs — package-level logrus functions that create/return the global
// logger instance.
var bannedFuncs = map[string]struct{}{
	"New":            {},
	"StandardLogger": {},
}

// Analyzer — rule GID-214: logrus.New()/StandardLogger() — only in the composition root (main, internal/app).
var Analyzer = &analysis.Analyzer{
	Name: "gidloggernew",
	Doc:  ruleID + ": logrus.New()/StandardLogger() are called only in the composition root (main, internal/app). Fix: pass a ready *logrus.Entry through the constructor",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	// composition root: package main or the app layer (internal/app, anchored to
	// the module root) — creating a logger here is allowed. Anchoring keeps a
	// nested "app" package below another layer out of the exemption.
	if pass.Pkg.Name() == "main" || pathseg.HasLayer(pass.Pkg.Path(), "app") {
		return nil, nil
	}

	for _, file := range pass.Files {
		if ast.IsGenerated(file) || isTestFile(pass, file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if name, ok := bannedLogrusCall(pass, call); ok {
				pass.Reportf(call.Pos(),
					"%s: logrus.%s() may be called only in the composition root (main, internal/app). "+
						"Fix: pass a ready *logrus.Entry through the constructor",
					ruleID, name)
			}
			return true
		})
	}
	return nil, nil
}

// bannedLogrusCall reports whether call is a call to the package-level function
// logrus.New()/logrus.StandardLogger(). Resolution is by types: the package name
// is taken from the object's import path, not from the selector text.
func bannedLogrusCall(pass *analysis.Pass, call *ast.CallExpr) (string, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	if _, ok := bannedFuncs[sel.Sel.Name]; !ok {
		return "", false
	}
	fn, ok := pass.TypesInfo.ObjectOf(sel.Sel).(*types.Func)
	if !ok {
		return "", false
	}
	// package-level function: no receiver (the WithField method is not flagged).
	sig, ok := fn.Type().(*types.Signature)
	if !ok || sig.Recv() != nil {
		return "", false
	}
	// logrusPkgPath — the import path of the logrus package.
	const logrusPkgPath = "github.com/sirupsen/logrus"
	pkg := fn.Pkg()
	if pkg == nil || pkg.Path() != logrusPkgPath {
		return "", false
	}
	return sel.Sel.Name, true
}

func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	tokenFile := pass.Fset.File(file.Pos())
	name := tokenFile.Name()
	const suffix = "_test.go"
	return len(name) >= len(suffix) && name[len(name)-len(suffix):] == suffix
}
