// Package allptr реализует правило GID-004: итерация for range по слайсу
// структур выполняется через gdhelper.AllPtr (go-styleguide, «Итерация по
// слайсам сущностей») — это исключает копирование элементов.
//
// Корректный код `for _, v := range gdhelper.AllPtr(s)` правило не задевает:
// AllPtr возвращает итератор (range-over-func), а не слайс.
package allptr

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-004"

// Analyzer — правило GID-004: итерация по слайсу структур — через gdhelper.AllPtr.
var Analyzer = &analysis.Analyzer{
	Name: "gidallptr",
	Doc:  ruleID + ": iterate over a slice of structs via gdhelper.AllPtr. Fix: range over gdhelper.AllPtr(items) to get pointers instead of copies.",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	const helperPkg = "gitlab.gid.team/gid-data/tech/golang/libs/helper.git"
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			rng, ok := n.(*ast.RangeStmt)
			if !ok {
				return true
			}
			if isStructSlice(pass.TypesInfo.TypeOf(rng.X)) {
				pass.Reportf(rng.X.Pos(),
					"%s: ranging over a slice of structs copies each element. "+
						"Fix: range over gdhelper.AllPtr(items) (%s) to iterate pointers.",
					ruleID, helperPkg)
			}
			return true
		})
	}
	return nil, nil
}

// isStructSlice сообщает, является ли тип слайсом структур. Слайсы
// указателей ([]*T) не задевает — там копирования элементов нет.
func isStructSlice(t types.Type) bool {
	if t == nil {
		return false
	}
	slice, ok := t.Underlying().(*types.Slice)
	if !ok {
		return false
	}
	elem := slice.Elem()
	_, isStruct := elem.Underlying().(*types.Struct)
	return isStruct
}
