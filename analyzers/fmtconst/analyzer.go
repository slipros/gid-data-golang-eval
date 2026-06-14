// Package fmtconst implements rule GID-186 (Uber: format strings outside
// Printf): the format string of printf-style functions must be a string
// literal or a constant, not a variable. When a variable sits in the format
// position, go vet cannot statically check that the verbs match the
// arguments — the diagnostic requires declaring the format as a separate
// const.
//
// What is matched (a variable in the format position):
//   - fmt.Printf/Sprintf/Errorf — arg 0; fmt.Fprintf — arg 1;
//   - github.com/pkg/errors Errorf — arg 0, Wrapf/WithMessagef — arg 1;
//   - log.Printf/Fatalf — arg 0.
//
// What is NOT matched (the format is a literal/constant):
//   - a string literal ("format %s");
//   - a const identifier (pass.TypesInfo gives constant value != nil);
//   - concatenation of constants ("a"+"b") — its value is a constant too.
//
// Boundary: functions without a format position (fmt.Sprint) and same-named
// functions of foreign packages / a local printf are not matched — the target
// functions are recognized by the typed package path (TypesInfo, typeutil.Callee).
//
// Generated code (ast.IsGenerated) is skipped. LoadMode — TypesInfo.
package fmtconst

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

const ruleID = "GID-186"

const (
	pkgFmt       = "fmt"
	pkgLog       = "log"
	pkgPkgErrors = "github.com/pkg/errors"
)

// Analyzer — GID-186 (gidfmtconst).
var Analyzer = &analysis.Analyzer{
	Name: "gidfmtconst",
	Doc:  ruleID + ": the format string of printf-like functions must be a literal or const, not a variable. Fix: declare a const format string",
	Run:  run,
}

// targetFuncs — printf-style functions and the index of the format argument in
// their call (accounting for the receiver/first argument: in Fprintf the format
// comes after the writer). The key is the package path.
var targetFuncs = map[string]map[string]int{
	pkgFmt: {
		"Printf":  0,
		"Sprintf": 0,
		"Errorf":  0,
		"Fprintf": 1,
	},
	pkgPkgErrors: {
		"Errorf":       0,
		"Wrapf":        1,
		"WithMessagef": 1,
	},
	pkgLog: {
		"Printf": 0,
		"Fatalf": 0,
	},
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
			checkCall(pass, call)
			return true
		})
	}
	return nil, nil
}

func checkCall(pass *analysis.Pass, call *ast.CallExpr) {
	idx, ok := formatArgIndex(pass, call)
	if !ok {
		return
	}
	if idx >= len(call.Args) {
		return
	}
	arg := call.Args[idx]
	if isConstString(pass, arg) {
		return
	}
	// The format is not a constant; make sure it is a string expression
	// (a variable of type string) and not something else.
	if !isStringExpr(pass, arg) {
		return
	}
	pass.Reportf(arg.Pos(),
		"%s: the format string is a variable. Fix: declare a const, otherwise vet cannot check the arguments", ruleID)
}

// formatArgIndex returns the index of the format argument if call invokes
// one of the target printf functions; otherwise ok=false. The function is
// recognized by the typed object (typeutil.Callee) and its package path.
func formatArgIndex(pass *analysis.Pass, call *ast.CallExpr) (int, bool) {
	fn := typeutil.Callee(pass.TypesInfo, call)
	f, ok := fn.(*types.Func)
	if !ok || f.Pkg() == nil {
		return 0, false
	}
	pkg := f.Pkg()
	byName, ok := targetFuncs[pkg.Path()]
	if !ok {
		return 0, false
	}
	idx, ok := byName[f.Name()]
	return idx, ok
}

// isConstString reports that the expression is a string constant (a literal,
// a const identifier, concatenation of constants). The constant value is
// available via pass.TypesInfo (tv.Value != nil).
func isConstString(pass *analysis.Pass, expr ast.Expr) bool {
	tv, ok := pass.TypesInfo.Types[expr]
	if !ok {
		return false
	}
	return tv.Value != nil
}

// isStringExpr reports that the expression has a string type.
func isStringExpr(pass *analysis.Pass, expr ast.Expr) bool {
	tv, ok := pass.TypesInfo.Types[expr]
	if !ok || tv.Type == nil {
		return false
	}
	basic, ok := tv.Type.Underlying().(*types.Basic)
	if !ok {
		return false
	}
	return basic.Info()&types.IsString != 0
}
