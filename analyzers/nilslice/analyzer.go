// Package nilslice implements rule GID-185 (Uber/Google: nil is a valid slice):
// an empty slice composite literal `[]T{}` is redundant — a nil slice is fully
// valid (it can be iterated, appended to, len(nil) == 0).
//
// What is matched:
//   - `return []T{}` — an empty slice literal in a return statement
//     → "return nil instead of an empty slice";
//   - `s := []T{}` and `var s = []T{}` — initializing a variable with an empty
//     literal → "declare a zero-value slice: var s []T".
//
// What is NOT matched:
//   - non-empty literals (`[]T{1, 2}`) — that is data, not "emptiness";
//   - `[]T{}` as a call argument or a struct field value — there an empty
//     (non-nil) slice may be deliberate semantics (e.g. JSON marshaling
//     `[]` vs `null`);
//   - map literals (`map[K]V{}`) and arrays (`[N]T{}`) — the rule is only about slices;
//   - `make([]T, ...)` — that is the domain of prealloc rules, not ours.
//
// LoadMode: TypesInfo — types are needed to tell a slice from an array/map.
// Generated code (ast.IsGenerated) is skipped.
// Targeted opt-out: //nolint:gidnilslice.
package nilslice

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-185"

// Analyzer — rule GID-185: return/declare a nil slice instead of an empty literal []T{}. Fix: use nil or var s []T.
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
				// s := []T{} — a short variable declaration.
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
				// var s = []T{} — a var declaration with an initializer.
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

// isEmptySliceLit: the expression is an empty composite literal whose type
// (per TypesInfo) is a slice. Arrays, maps, and non-empty literals are filtered out.
func isEmptySliceLit(pass *analysis.Pass, expr ast.Expr) bool {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return false
	}
	if len(lit.Elts) != 0 {
		return false // a non-empty literal is data.
	}
	t := pass.TypesInfo.TypeOf(lit)
	if t == nil {
		return false
	}
	_, isSlice := t.Underlying().(*types.Slice)
	return isSlice
}
