// Package onlypkgerrors implements rule GID-146: github.com/pkg/errors is the
// only error package of the codebase. Creating errors through the std
// errors.New/errors.Join and fmt.Errorf is forbidden, and so is the std errors
// package itself — importing it under any alias.
//
// The import ban rests on a fact the rule used to get wrong: pkg/errors v0.9.0+
// re-exports Is/As/Unwrap (go113.go — one-line delegates to the std functions),
// so chain inspection needs no second package. A file that imports
// stderrors "errors" next to pkg/errors gains nothing and splits error handling
// across two packages (incident 2026-08-06, resource-registry
// internal/dal/entity/error.go: the std package pulled in for a single
// errors.Is).
//
// Whichever alias the std package hides behind, the import is reported once per
// file, and the constructor calls inside that file are left alone: the fix is
// one — drop the import, let the calls resolve to pkg/errors.
//
// A _test.go file is not judged for the import: a test of an error predicate
// takes its expected value from the std function on purpose
// (want := stderrors.Is(other.sentinel, p.sentinel)) — checking pkg/errors with
// pkg/errors would prove nothing. Creating errors through the std package stays
// forbidden in tests too.
//
// errors.Join has no counterpart in pkg/errors, and neither does
// errors.ErrUnsupported — the rare place that needs one takes a
// //nolint:gidonlypkgerrors.
package onlypkgerrors

import (
	"go/ast"
	"go/types"
	"strconv"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const (
	ruleID       = "GID-146"
	allowedPkg   = "github.com/pkg/errors"
	stdErrorsPkg = "errors"
)

// forbidden — std error constructors: package -> functions.
var forbidden = map[string]map[string]struct{}{
	stdErrorsPkg: {"New": {}, "Join": {}},
	"fmt":        {"Errorf": {}},
}

// Analyzer — rule GID-146: errors are handled only through github.com/pkg/errors.
var Analyzer = &analysis.Analyzer{
	Name:     "gidonlypkgerrors",
	Doc:      ruleID + ": errors are created and inspected only through " + allowedPkg,
	Requires: astwalk.Requires,
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	// The import already carries the whole file's fix — reporting the calls
	// inside it on top of that would repeat one diagnostic N times. Imports are
	// read off file.Imports, so this costs no traversal.
	importReported := make(map[*ast.File]bool, len(pass.Files))
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		importReported[file] = reportStdImport(pass, file)
	}

	astwalk.NodesOf(pass, ast.IsGenerated, func(file *ast.File, n *ast.CallExpr) {
		f, ok := typeutil.Callee(pass.TypesInfo, n).(*types.Func)
		if !ok || f.Pkg() == nil {
			return
		}
		fPkg := f.Pkg()
		names, ok := forbidden[fPkg.Path()]
		if !ok {
			return
		}
		if _, ok := names[f.Name()]; !ok {
			return
		}
		if importReported[file] && fPkg.Path() == stdErrorsPkg {
			return
		}
		pass.Reportf(n.Pos(),
			"%s: %s.%s is forbidden. Fix: use only %s for errors",
			ruleID, fPkg.Name(), f.Name(), allowedPkg)
	})

	return nil, nil
}

// reportStdImport reports every import of the std errors package in file and
// tells whether it reported anything.
func reportStdImport(pass *analysis.Pass, file *ast.File) bool {
	if srcfile.IsTest(pass, file) {
		return false
	}

	var reported bool
	for _, spec := range file.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil || path != stdErrorsPkg {
			continue
		}
		pass.Reportf(spec.Pos(),
			"%s: the std errors package is forbidden, %s re-exports Is/As/Unwrap. "+
				"Fix: import %q alone and call errors.Is(err, ErrNoResult)",
			ruleID, allowedPkg, allowedPkg)
		reported = true
	}

	return reported
}
