// Package nilslice реализует правило GID-185 (Uber/Google: nil is a valid slice):
// пустой композит-литерал слайса `[]T{}` лишний — nil-слайс полноценно валиден
// (его можно итерировать, к нему можно делать append, len(nil) == 0).
//
// Что матчится:
//   - `return []T{}` — пустой литерал слайса в return-операторе
//     → «возвращайте nil вместо пустого слайса»;
//   - `s := []T{}` и `var s = []T{}` — инициализация переменной пустым литералом
//     → «объявляйте zero-value слайс: var s []T».
//
// Что НЕ матчится:
//   - непустые литералы (`[]T{1, 2}`) — это данные, а не «пустота»;
//   - `[]T{}` как аргумент вызова или значение поля структуры — там пустой
//     (не-nil) слайс может быть осознанной семантикой (например, JSON-маршалинг
//     `[]` против `null`);
//   - map-литералы (`map[K]V{}`) и массивы (`[N]T{}`) — правило только про слайсы;
//   - `make([]T, ...)` — это зона правил prealloc, не наша.
//
// LoadMode: TypesInfo — нужны типы, чтобы отличить слайс от массива/мапы.
// Сгенерированный код (ast.IsGenerated) пропускается.
// Точечное отключение: //nolint:gidnilslice.
package nilslice

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-185"

// Analyzer — правило GID-185: return/declare a nil slice instead of an empty literal []T{}. Fix: use nil or var s []T.
var Analyzer = &analysis.Analyzer{
	Name: "gidnilslice",
	Doc:  ruleID + ": return/declare a nil slice instead of an empty literal []T{}. Fix: use nil or var s []T",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.ReturnStmt:
				for _, res := range node.Results {
					if isEmptySliceLit(pass, res) {
						pass.Reportf(res.Pos(),
							"%s: return nil instead of an empty slice. Fix: a nil slice is valid", ruleID)
					}
				}
			case *ast.AssignStmt:
				// s := []T{} — короткое объявление переменной.
				if node.Tok != token.DEFINE {
					return true
				}
				for _, rhs := range node.Rhs {
					if isEmptySliceLit(pass, rhs) {
						pass.Reportf(rhs.Pos(),
							"%s: declare a zero-value slice. Fix: var s []T", ruleID)
					}
				}
			case *ast.ValueSpec:
				// var s = []T{} — объявление через var с инициализатором.
				for _, val := range node.Values {
					if isEmptySliceLit(pass, val) {
						pass.Reportf(val.Pos(),
							"%s: declare a zero-value slice. Fix: var s []T", ruleID)
					}
				}
			}
			return true
		})
	}
	return nil, nil
}

// isEmptySliceLit: выражение — пустой композит-литерал, чей тип (по TypesInfo)
// является слайсом. Массивы, мапы и непустые литералы отсекаются.
func isEmptySliceLit(pass *analysis.Pass, expr ast.Expr) bool {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return false
	}
	if len(lit.Elts) != 0 {
		return false // непустой литерал — это данные.
	}
	t := pass.TypesInfo.TypeOf(lit)
	if t == nil {
		return false
	}
	_, isSlice := t.Underlying().(*types.Slice)
	return isSlice
}
