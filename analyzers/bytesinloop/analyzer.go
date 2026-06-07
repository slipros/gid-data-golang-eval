// Package bytesinloop реализует правило GID-182 (Uber: avoid repeated
// string-to-byte conversions): конверсия строкового литерала или константы
// в []byte/[]rune внутри тела цикла вычисляется один раз перед циклом.
//
// Что матчится:
//   - []byte("literal") внутри тела for/range (включая вложенные блоки);
//   - []rune("literal") там же;
//   - []byte(constStr), где constStr — string-константа (значение
//     вычисляется через pass.TypesInfo, constant value, types.String);
//   - конверсия внутри тела замыкания, объявленного в цикле (замыкание
//     выполняется на каждой итерации).
//
// Что НЕ матчится:
//   - []byte(variable) — конверсия переменной (не константы): значение
//     может меняться, выносить нельзя;
//   - []byte("literal") вне цикла — вычисляется один раз и так;
//   - []byte(param), где param — параметр функции/замыкания.
//
// Сгенерированный код (ast.IsGenerated) пропускается. LoadMode — TypesInfo.
package bytesinloop

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-182"

// Analyzer — правило GID-182: конверсия строкового литерала/константы в []byte/[]rune внутри цикла.
var Analyzer = &analysis.Analyzer{
	Name: "gidbytesinloop",
	Doc:  ruleID + ": converting a string literal/constant to []byte/[]rune inside a loop. Fix: compute the conversion once before the loop.",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}

		// Собираем позиционные диапазоны тел всех циклов (for/range).
		// Вложенные блоки, тела замыканий, объявленных в цикле, лексически
		// находятся внутри этого диапазона — и потому считаются «в цикле».
		var loopBodies []*ast.BlockStmt
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.ForStmt:
				loopBodies = append(loopBodies, node.Body)
			case *ast.RangeStmt:
				loopBodies = append(loopBodies, node.Body)
			}
			return true
		})

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if !insideAnyLoop(call.Pos(), loopBodies) {
				return true
			}
			checkConversion(pass, call)
			return true
		})
	}
	return nil, nil
}

// insideAnyLoop сообщает, лежит ли позиция pos внутри тела хотя бы одного цикла.
func insideAnyLoop(pos token.Pos, bodies []*ast.BlockStmt) bool {
	for _, b := range bodies {
		// Lbrace < pos < Rbrace — позиция строго внутри фигурных скобок тела.
		if pos > b.Lbrace && pos < b.Rbrace {
			return true
		}
	}
	return false
}

// checkConversion: если call — это конверсия []byte(X)/[]rune(X), где X —
// строковая константа, выдаёт диагностику.
func checkConversion(pass *analysis.Pass, call *ast.CallExpr) {
	kind, ok := sliceConversionKind(call.Fun)
	if !ok {
		return
	}
	if len(call.Args) != 1 {
		return
	}
	arg := call.Args[0]
	tv, ok := pass.TypesInfo.Types[arg]
	if !ok || tv.Value == nil {
		return // не константа (переменная, параметр, вызов) — пропускаем.
	}
	// Значение — константа; убеждаемся, что её тип — строковый.
	basic, ok := tv.Type.Underlying().(*types.Basic)
	if !ok || basic.Info()&types.IsString == 0 {
		return
	}
	pass.Reportf(call.Pos(),
		"%s: converting to %s inside a loop repeats the allocation. "+
			"Fix: compute it once before the loop.", ruleID, kind)
}

// sliceConversionKind: если fun — это тип []byte или []rune (в форме
// ArrayType без длины с элементом byte/rune), возвращает строку "[]byte"
// либо "[]rune".
func sliceConversionKind(fun ast.Expr) (string, bool) {
	arr, ok := fun.(*ast.ArrayType)
	if !ok || arr.Len != nil {
		return "", false // не слайс ([N]T — массив, не конверсия здесь).
	}
	elt, ok := arr.Elt.(*ast.Ident)
	if !ok {
		return "", false
	}
	switch elt.Name {
	case "byte":
		return "[]byte", true
	case "rune":
		return "[]rune", true
	default:
		return "", false
	}
}
