package codepolicy

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

func checkNilSuccessReturns(pass *analysis.Pass, fn *ast.FuncType, body *ast.BlockStmt) {
	if !returnsError(pass, fn) {
		return
	}
	var results []types.Type
	for _, field := range fn.Results.List {
		count := max(1, len(field.Names))
		for range count {
			results = append(results, pass.TypesInfo.TypeOf(field.Type))
		}
	}
	if len(results) < 2 {
		return
	}
	ast.Inspect(body, func(n ast.Node) bool {
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}
		ret, ok := n.(*ast.ReturnStmt)
		if !ok || len(ret.Results) != len(results) || !nilReturnExpr(pass, ret.Results[len(ret.Results)-1]) {
			return true
		}
		for i, t := range results[:len(results)-1] {
			if nonSliceNilable(t) && nilReturnExpr(pass, ret.Results[i]) && !boxesNilValue(pass, t, ret.Results[i]) {
				report(pass, ret, "298", "non-slice nil result is returned with a nil error", "return an error for an absent result; only slices may be nil on success")
				break
			}
		}
		return true
	})
}

func nilReturnExpr(pass *analysis.Pass, expr ast.Expr) bool {
	if universeNil(pass, expr) {
		return true
	}
	call, ok := ast.Unparen(expr).(*ast.CallExpr)
	if !ok || len(call.Args) != 1 || !pass.TypesInfo.Types[call.Fun].IsType() {
		return false
	}
	return universeNil(pass, call.Args[0])
}

func boxesNilValue(pass *analysis.Pass, target types.Type, expr ast.Expr) bool {
	if _, ok := target.Underlying().(*types.Interface); !ok || universeNil(pass, expr) {
		return false
	}
	exprType := pass.TypesInfo.TypeOf(expr)
	_, isInterface := exprType.Underlying().(*types.Interface)
	return !isInterface
}
