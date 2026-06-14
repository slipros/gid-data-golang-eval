// Package onlypkgerrors implements rule GID-146: only github.com/pkg/errors
// is used for error handling. Creating errors via the standard errors.New
// and fmt.Errorf is forbidden everywhere.
//
// Inspecting the error chain — std errors.Is/As/Unwrap — is not creation,
// it is allowed (pkg/errors does not have these functions).
package onlypkgerrors

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

const (
	ruleID     = "GID-146"
	allowedPkg = "github.com/pkg/errors"
)

// forbidden — std error constructors: package -> functions.
var forbidden = map[string]map[string]struct{}{
	"errors": {"New": {}, "Join": {}},
	"fmt":    {"Errorf": {}},
}

// Analyzer — rule GID-146: errors are created only via github.com/pkg/errors.
var Analyzer = &analysis.Analyzer{
	Name: "gidonlypkgerrors",
	Doc:  ruleID + ": errors are created only via " + allowedPkg,
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			f, ok := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
			if !ok || f.Pkg() == nil {
				return true
			}
			fPkg := f.Pkg()
			names, ok := forbidden[fPkg.Path()]
			if !ok {
				return true
			}
			if _, ok := names[f.Name()]; !ok {
				return true
			}
			pass.Reportf(call.Pos(),
				"%s: %s.%s is forbidden. Fix: use only %s for errors",
				ruleID, fPkg.Name(), f.Name(), allowedPkg)
			return true
		})
	}
	return nil, nil
}
