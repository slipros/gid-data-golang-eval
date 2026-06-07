// Package subtestname реализует правило GID-191: имена subtest в t.Run/b.Run
// не содержат пробелов и слешей.
//
// go test -run 'Test/имя' матчит подтесты по имени, заменяя пробелы на
// подчёркивания и используя '/' как разделитель уровней. Если имя subtest
// содержит пробел или '/', точный -run по нему не сработает — поэтому имена
// подтестов пишутся в snake_case.
//
// Матчим только вызовы методов Run на *testing.T / *testing.B, где первый
// аргумент — строковый ЛИТЕРАЛ или КОНСТАНТА (значение известно через
// pass.TypesInfo). Неконстантные имена (tt.name из table-driven) не матчатся:
// значения таблицы — отдельная зона ответственности.
//
// Сгенерированный код (ast.IsGenerated) пропускается. LoadMode — TypesInfo.
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

// Analyzer — правило GID-191: имена subtest в t.Run без пробелов и слешей.
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

	// Имя subtest — первый аргумент. Берём значение константы/литерала
	// из TypesInfo; неконстантные выражения (tt.name) не матчим.
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

// isTestingRunReceiver сообщает, что выражение x имеет тип *testing.T или
// *testing.B (ресивер метода Run из пакета testing).
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
