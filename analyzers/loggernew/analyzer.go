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

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-214"

// bannedFuncs — package-level functions that create or hand out the global
// logger instance, per stack: logrus.New/StandardLogger and, for the stdlib
// stack, slog.New/Default/SetDefault. The two lists are separate because
// slog.Default() is the same smell as logrus.StandardLogger(), while the
// names do not overlap.
var bannedFuncs = map[string]map[string]struct{}{
	"github.com/sirupsen/logrus": {
		"New":            {},
		"StandardLogger": {},
	},
	"log/slog": {
		"New":        {},
		"Default":    {},
		"SetDefault": {},
	},
}

// Analyzer — rule GID-214: creating or grabbing the global logger (logrus.New/StandardLogger, slog.New/Default/SetDefault) — only in the composition root (main, internal/app).
var Analyzer = &analysis.Analyzer{
	Name: "gidloggernew",
	Doc: ruleID + ": the logger is created only in the composition root (main, internal/app) — " +
		"logrus.New()/StandardLogger() and slog.New()/Default()/SetDefault() are banned elsewhere. " +
		"Fix: pass a ready logger (*logrus.Entry, *slog.Logger) through the constructor",
	Requires: astwalk.Requires,
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	// composition root: package main or the app layer (internal/app, anchored to
	// the module root) — creating a logger here is allowed. Anchoring keeps a
	// nested "app" package below another layer out of the exemption.
	if pass.Pkg.Name() == "main" || pathseg.HasLayer(pass.Pkg.Path(), "app") {
		return nil, nil
	}

	skip := func(file *ast.File) bool {
		return ast.IsGenerated(file) || isTestFile(pass, file)
	}

	astwalk.NodesOf(pass, skip, func(_ *ast.File, call *ast.CallExpr) {
		if pkgName, name, ok := bannedLoggerCall(pass, call); ok {
			pass.Reportf(call.Pos(),
				"%s: %s.%s() may be called only in the composition root (main, internal/app). "+
					"Fix: pass a ready logger (*logrus.Entry, *slog.Logger) through the constructor",
				ruleID, pkgName, name)
		}
	})

	return nil, nil
}

// bannedLoggerCall reports whether call creates or grabs the global logger of
// either stack, returning the package name and the function name. Resolution
// is by types: the package is taken from the object's import path, not from
// the selector text.
func bannedLoggerCall(pass *analysis.Pass, call *ast.CallExpr) (pkgName, name string, ok bool) {
	sel, isSel := call.Fun.(*ast.SelectorExpr)
	if !isSel {
		return "", "", false
	}
	fn, isFunc := pass.TypesInfo.ObjectOf(sel.Sel).(*types.Func)
	if !isFunc {
		return "", "", false
	}
	// package-level function: no receiver (the WithField method is not flagged).
	sig, isSig := fn.Type().(*types.Signature)
	if !isSig || sig.Recv() != nil {
		return "", "", false
	}
	pkg := fn.Pkg()
	if pkg == nil {
		return "", "", false
	}
	banned, known := bannedFuncs[pkg.Path()]
	if !known {
		return "", "", false
	}
	if _, isBanned := banned[sel.Sel.Name]; !isBanned {
		return "", "", false
	}
	return pkg.Name(), sel.Sel.Name, true
}

func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	tokenFile := pass.Fset.File(file.Pos())
	name := tokenFile.Name()
	const suffix = "_test.go"
	return len(name) >= len(suffix) && name[len(name)-len(suffix):] == suffix
}
