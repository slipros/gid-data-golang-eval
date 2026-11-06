// Package allptr implements rule GID-004: a for range iteration over a slice
// of structs must go through gdhelper.AllPtr (go-styleguide, "Iterating over
// entity slices") — this avoids copying the elements.
//
// Correct code `for _, v := range gdhelper.AllPtr(s)` is not flagged:
// AllPtr returns an iterator (range-over-func), not a slice.
//
// AllPtr yields iter.Seq2[int, *T], preserving indices when needed.
// Index-only loops are checked only when every use reaches a field or takes
// an element address. Loops with numeric index uses, whole-element replacement
// or collection mutation are conservatively retained.
package allptr

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
)

const ruleID = "GID-004"

// Analyzer — rule GID-004: iterate over a slice of structs via gdhelper.AllPtr.
var Analyzer = &analysis.Analyzer{
	Name:     "gidallptr",
	Doc:      ruleID + ": iterate over a slice of structs via gdhelper.AllPtr. Fix: range over gdhelper.AllPtr(items) to get pointers instead of copies.",
	Requires: astwalk.Requires,
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	const helperPkg = "gitlab.gid.team/gid-data/tech/golang/libs/helper.git"
	astwalk.NodesOf(pass, ast.IsGenerated, func(_ *ast.File, rng *ast.RangeStmt) {
		// No element is copied. Check the narrow field/address idiom separately,
		// preserving loops whose mutation or indexing semantics need review.
		if rng.Value == nil {
			if isStructSlice(pass.TypesInfo.TypeOf(rng.X)) && onlyElementIndexUses(pass, rng) {
				pass.Reportf(rng.X.Pos(), "%s: index-only loop accesses struct elements through the slice. Fix: use for _, item := range gdhelper.AllPtr(items) (%s) and access fields through item", ruleID, helperPkg)
			}
			return
		}
		if isStructSlice(pass.TypesInfo.TypeOf(rng.X)) {
			pass.Reportf(rng.X.Pos(),
				"%s: ranging over a slice of structs copies each element. "+
					"Fix: range over gdhelper.AllPtr(items) (%s) to iterate pointers.",
				ruleID, helperPkg)
		}
	})

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

// onlyElementIndexUses permits mutation via fields/pointers, but keeps loops that
// use the numeric index or replace whole elements to avoid assuming a rewrite.
func onlyElementIndexUses(pass *analysis.Pass, rng *ast.RangeStmt) bool {
	key, ok := rng.Key.(*ast.Ident)
	if !ok || key.Name == "_" {
		return false
	}
	slice, ok := ast.Unparen(rng.X).(*ast.Ident)
	if !ok {
		return false
	}
	keyObj, sliceObj := pass.TypesInfo.ObjectOf(key), pass.TypesInfo.ObjectOf(slice)
	if keyObj == nil || sliceObj == nil {
		return false
	}
	seen, safe := false, true
	var stack []ast.Node
	ast.Inspect(rng.Body, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		if id, ok := n.(*ast.Ident); ok && pass.TypesInfo.ObjectOf(id) == keyObj {
			seen = true
			safe = safe && safeElementIndex(pass, stack, id, sliceObj)
		}
		// Reassigning/reslicing the collection or modifying it through a call
		// can change which element a numeric index would refer to.
		if id, ok := n.(*ast.Ident); ok && pass.TypesInfo.ObjectOf(id) == sliceObj {
			if len(stack) == 0 {
				safe = false
			} else {
				index, ok := stack[len(stack)-1].(*ast.IndexExpr)
				if !ok || index.X != id {
					safe = false
				}
			}
		}
		stack = append(stack, n)
		return true
	})
	return seen && safe
}

func safeElementIndex(pass *analysis.Pass, stack []ast.Node, id *ast.Ident, sliceObj types.Object) bool {
	if len(stack) < 2 {
		return false
	}
	index, ok := stack[len(stack)-1].(*ast.IndexExpr)
	if !ok || index.Index != id {
		return false
	}
	base, ok := ast.Unparen(index.X).(*ast.Ident)
	if !ok || pass.TypesInfo.ObjectOf(base) != sliceObj {
		return false
	}
	switch parent := stack[len(stack)-2].(type) {
	case *ast.SelectorExpr:
		return parent.X == index
	case *ast.UnaryExpr:
		return parent.Op == token.AND
	default:
		return false
	}
}
