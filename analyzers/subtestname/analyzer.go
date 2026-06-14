// Package subtestname implements rule GID-191: subtest names in t.Run/b.Run
// contain no spaces or slashes.
//
// go test -run 'Test/name' matches subtests by name, replacing spaces with
// underscores and using '/' as the level separator. If a subtest name
// contains a space or '/', an exact -run on it will not work — therefore
// subtest names are written in snake_case.
//
// Only Run method calls on *testing.T / *testing.B are matched, where the
// first argument is a string LITERAL or CONSTANT (the value is known via
// pass.TypesInfo). Non-constant names (tt.name from table-driven tests) are
// not matched: table values are a separate area of responsibility.
//
// Generated code (ast.IsGenerated) is skipped. LoadMode — TypesInfo.
package subtestname

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-191"

// Analyzer — rule GID-191: subtest names in t.Run without spaces or slashes.
var Analyzer = &analysis.Analyzer{
	Name: "gidsubtestname",
	Doc:  ruleID + ": subtest names in t.Run/b.Run have no spaces or slashes (snake_case). Fix: rename to snake_case",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				checkCall(pass, call)
			}
			return true
		})
	}
	return nil, nil
}

func checkCall(pass *analysis.Pass, call *ast.CallExpr) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Run" || len(call.Args) == 0 {
		return
	}
	if !isTestingRunReceiver(pass, sel.X) {
		return
	}

	// The subtest name is the first argument. Take the constant/literal
	// value from TypesInfo; non-constant expressions (tt.name) are not matched.
	tv, ok := pass.TypesInfo.Types[call.Args[0]]
	if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
		return
	}
	name := constant.StringVal(tv.Value)

	pos := call.Args[0].Pos()
	if strings.Contains(name, "/") {
		report(pass, pos, name, "a slash '/'")
		return
	}
	if strings.ContainsRune(name, ' ') {
		report(pass, pos, name, "a space")
	}
}

func report(pass *analysis.Pass, pos token.Pos, name, what string) {
	pass.Reportf(pos,
		"%s: subtest name %q contains %s. Fix: use snake_case, "+
			"go test -run 'Test/name' will not match it", ruleID, name, what)
}

// isTestingRunReceiver reports whether expression x has type *testing.T or
// *testing.B (the receiver of the Run method from the testing package).
func isTestingRunReceiver(pass *analysis.Pass, x ast.Expr) bool {
	t := pass.TypesInfo.TypeOf(x)
	if t == nil {
		return false
	}
	ptr, ok := t.(*types.Pointer)
	if !ok {
		return false
	}
	named, ok := ptr.Elem().(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	if pkg == nil || pkg.Path() != "testing" {
		return false
	}
	return obj.Name() == "T" || obj.Name() == "B"
}
