// Package fmtconst реализует правило GID-186 (Uber: format strings outside
// Printf): format-строка printf-style функций должна быть строковым
// литералом или константой, а не переменной. Если в позиции format стоит
// переменная, go vet не может статически проверить соответствие
// verb-ов аргументам — диагностика требует объявить format отдельной
// const.
//
// Что матчится (переменная в позиции format):
//   - fmt.Printf/Sprintf/Errorf — арг 0; fmt.Fprintf — арг 1;
//   - github.com/pkg/errors Errorf — арг 0, Wrapf/WithMessagef — арг 1;
//   - log.Printf/Fatalf — арг 0.
//
// Что НЕ матчится (format — литерал/константа):
//   - строковый литерал ("формат %s");
//   - const-идентификатор (pass.TypesInfo даёт constant value != nil);
//   - конкатенация констант ("a"+"b") — её значение тоже константа.
//
// Граница: функции без позиции format (fmt.Sprint) и одноимённые функции
// чужих пакетов / локальные printf не матчатся — целевые функции
// распознаются по типизированному пути пакета (TypesInfo, typeutil.Callee).
//
// Сгенерированный код (ast.IsGenerated) пропускается. LoadMode — TypesInfo.
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
	Doc:  ruleID + ": format-строка printf-функций — литерал или const, не переменная",
	Run:  run,
}

// targetFuncs — printf-style функции и индекс аргумента-format в их вызове
// (с учётом ресивера/первого аргумента: у Fprintf format идёт после writer).
// Ключ — путь пакета.
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
	// format — не константа; убеждаемся, что это строковое выражение
	// (переменная типа string), а не что-то иное.
	if !isStringExpr(pass, arg) {
		return
	}
	pass.Reportf(arg.Pos(),
		"%s: format-строка — переменная; объявите const, иначе vet не проверит аргументы", ruleID)
}

// formatArgIndex возвращает индекс аргумента-format, если call — вызов
// одной из целевых printf-функций; иначе ok=false. Функция распознаётся
// по типизированному объекту (typeutil.Callee) и пути её пакета.
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

// isConstString сообщает, что выражение — строковая константа (литерал,
// const-идентификатор, конкатенация констант). Значение константы
// доступно через pass.TypesInfo (tv.Value != nil).
func isConstString(pass *analysis.Pass, expr ast.Expr) bool {
	tv, ok := pass.TypesInfo.Types[expr]
	if !ok {
		return false
	}
	return tv.Value != nil
}

// isStringExpr сообщает, что выражение имеет строковый тип.
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
