package codepolicy

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

func checkSuccessNil(pass *analysis.Pass, block *ast.BlockStmt) {
	for i := 2; i < len(block.List); i++ {
		assignment := errorResultAssignment(pass, block.List[i-2])
		if assignment == nil {
			continue
		}
		guard, ok := block.List[i-1].(*ast.IfStmt)
		if !ok || guard.Init != nil || guard.Else != nil || !endsReturn(guard.Body) {
			continue
		}
		if errObj := object(pass, assignment.Lhs[len(assignment.Lhs)-1]); errObj == nil || !successNilGuard(pass, guard.Cond, errObj) {
			continue
		}
		check, ok := block.List[i].(*ast.IfStmt)
		if !ok || check.Init != nil {
			continue
		}
		for _, lhs := range assignment.Lhs[:len(assignment.Lhs)-1] {
			result := object(pass, lhs)
			if result == nil || result.Name() == "_" || !nonSliceNilable(result.Type()) || referencesResult(pass, guard.Body, result) {
				continue
			}
			checkNilCondition(pass, check.Cond, result)
		}
	}
}

func errorResultAssignment(pass *analysis.Pass, stmt ast.Stmt) *ast.AssignStmt {
	assignment, ok := stmt.(*ast.AssignStmt)
	if !ok || len(assignment.Rhs) != 1 || len(assignment.Lhs) < 2 {
		return nil
	}
	call, ok := ast.Unparen(assignment.Rhs[0]).(*ast.CallExpr)
	if !ok {
		return nil
	}
	results, ok := pass.TypesInfo.TypeOf(call).(*types.Tuple)
	if !ok || results.Len() != len(assignment.Lhs) {
		return nil
	}
	lastResult := results.At(results.Len() - 1)
	errorType := types.Universe.Lookup("error")
	if !types.Identical(lastResult.Type(), errorType.Type()) {
		return nil
	}
	return assignment
}

func successNilGuard(pass *analysis.Pass, expr ast.Expr, errObj types.Object) bool {
	binary, ok := ast.Unparen(expr).(*ast.BinaryExpr)
	if !ok || binary.Op != token.NEQ {
		return false
	}
	return object(pass, binary.X) == errObj && universeNil(pass, binary.Y) || object(pass, binary.Y) == errObj && universeNil(pass, binary.X)
}

func nonSliceNilable(t types.Type) bool {
	if _, ok := types.Unalias(t).(*types.TypeParam); ok {
		return false
	}
	switch t.Underlying().(type) {
	case *types.Pointer, *types.Interface, *types.Map, *types.Chan, *types.Signature:
		return true
	default:
		return false
	}
}

func referencesResult(pass *analysis.Pass, node ast.Node, result types.Object) bool {
	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && pass.TypesInfo.ObjectOf(id) == result {
			found = true
		}
		return !found
	})
	return found
}

func checkNilCondition(pass *analysis.Pass, expr ast.Expr, result types.Object) {
	switch expr := ast.Unparen(expr).(type) {
	case *ast.UnaryExpr:
		if expr.Op == token.NOT {
			checkNilCondition(pass, expr.X, result)
		}
	case *ast.BinaryExpr:
		switch expr.Op {
		case token.LAND, token.LOR:
			// A call on the left may mutate the result before the right is evaluated.
			checkNilCondition(pass, expr.X, result)
			if !conditionCalls(expr.X) {
				checkNilCondition(pass, expr.Y, result)
			}
		case token.EQL, token.NEQ:
			if object(pass, expr.X) == result && universeNil(pass, expr.Y) || object(pass, expr.Y) == result && universeNil(pass, expr.X) {
				report(pass, expr, "298", "non-slice result is checked for nil after error handling", "ensure the callee returns an error for an absent result, then remove this nil comparison and keep other conditions; only slices may be nil on success")
			}
		}
	}
}

func universeNil(pass *analysis.Pass, expr ast.Expr) bool {
	return object(pass, expr) == types.Universe.Lookup("nil")
}

func conditionCalls(expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if _, ok := n.(*ast.CallExpr); ok {
			found = true
		}
		return !found
	})
	return found
}
