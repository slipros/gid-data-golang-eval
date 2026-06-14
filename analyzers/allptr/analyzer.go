// Package allptr implements rule GID-004: a for range iteration over a slice
// of structs must go through gdhelper.AllPtr (go-styleguide, "Iterating over
// entity slices") — this avoids copying the elements.
//
// Correct code `for _, v := range gdhelper.AllPtr(s)` is not flagged:
// AllPtr returns an iterator (range-over-func), not a slice.
package allptr

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-004"

// Analyzer — rule GID-004: iterate over a slice of structs via gdhelper.AllPtr.
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

// isStructSlice reports whether the type is a slice of structs. Slices of
// pointers ([]*T) are not flagged — there is no element copying there.
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
