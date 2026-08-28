// Package unhandledswitch implements rule GID-274: a switch over a named enum or
// discriminator handles an unknown value with the standard unhandled-value
// error.
//
// A value of a named string type with a package-level constant of that same
// type is an enum for this rule. A plain string switch is not treated as an
// enum because its finite set cannot be determined from the package's
// declarations.
//
// An enum switch is compliant only when its default clause has one statement:
// a direct return of errors.WithStack(gderror.NewUnhandledValueError(the
// switch tag expression)). The calls are resolved as static, package-qualified
// functions by symbol name. Generated files and _test.go files are skipped.
package unhandledswitch

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const ruleID = "GID-274"

// Analyzer — rule GID-274: enum/discriminator switches must return the
// unhandled-value error from their default case.
var Analyzer = &analysis.Analyzer{
	Name:     "gidunhandledswitch",
	Doc:      ruleID + ": a switch over an enum or discriminator must have a default that directly returns errors.WithStack(gderror.NewUnhandledValueError(the switch value)). Fix: add default: return errors.WithStack(gderror.NewUnhandledValueError(value)).",
	Requires: astwalk.Requires,
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	const defaultContractDiagnostic = ruleID + ": enum/discriminator switch default must directly return " +
		"errors.WithStack(gderror.NewUnhandledValueError(the switch value)). " +
		"Fix: add default: return errors.WithStack(gderror.NewUnhandledValueError(value))."

	skip := func(file *ast.File) bool {
		return ast.IsGenerated(file) || srcfile.IsTest(pass, file)
	}

	astwalk.NodesOf(pass, skip, func(_ *ast.File, sw *ast.SwitchStmt) {
		if sw.Tag == nil || !isEnum(pass.TypesInfo.TypeOf(sw.Tag)) {
			return
		}
		if hasCompliantDefault(pass, sw) {
			return
		}

		pass.Reportf(sw.Switch, defaultContractDiagnostic)
	})

	return nil, nil
}

// isEnum reports whether t is a string-based named type with a package-level
// constant of exactly that type. A named type without constants is an ordinary
// domain value, not a finite discriminator for this rule.
func isEnum(t types.Type) bool {
	if t == nil {
		return false
	}

	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return false
	}

	basic, ok := named.Underlying().(*types.Basic)
	if !ok || basic.Info()&types.IsString == 0 {
		return false
	}

	return hasTypedConst(named)
}

// hasTypedConst reports whether the package that declares named has a
// package-level constant whose type is named. This keeps local aliases and
// arbitrary named strings out of the enum set while recognizing imported enum
// types as well.
func hasTypedConst(named *types.Named) bool {
	obj := named.Obj()
	if obj == nil {
		return false
	}
	pkg := obj.Pkg()
	if pkg == nil {
		return false
	}
	scope := pkg.Scope()

	for _, name := range scope.Names() {
		constant, ok := scope.Lookup(name).(*types.Const)
		if !ok {
			continue
		}
		if types.Identical(types.Unalias(constant.Type()), named) {
			return true
		}
	}

	return false
}

// hasCompliantDefault reports whether the switch has a default clause whose
// only statement directly returns the required wrapped error for the switch
// tag. A default with no return, extra statements, a different handler, or a
// different value is intentionally not compliant.
func hasCompliantDefault(pass *analysis.Pass, sw *ast.SwitchStmt) bool {
	for _, stmt := range sw.Body.List {
		clause, ok := stmt.(*ast.CaseClause)
		if !ok || clause.List != nil {
			continue
		}
		if len(clause.Body) != 1 {
			return false
		}

		ret, ok := clause.Body[0].(*ast.ReturnStmt)
		if !ok || len(ret.Results) != 1 {
			return false
		}

		return isRequiredUnhandledReturn(pass, ret.Results[0], sw.Tag)
	}

	return false
}

// isRequiredUnhandledReturn recognizes exactly
//
//	return errors.WithStack(gderror.NewUnhandledValueError(switchTag))
//
// Both calls must be static package-qualified functions with the required
// symbol name. Import paths are deliberately not used: the helper package has
// more than one module path in the codebase, while dynamic calls and methods
// do not prove the required contract.
func isRequiredUnhandledReturn(pass *analysis.Pass, result, tag ast.Expr) bool {
	const (
		withStackName           = "WithStack"
		unhandledValueErrorName = "NewUnhandledValueError"
	)

	outer, ok := unparen(result).(*ast.CallExpr)
	if !ok || len(outer.Args) != 1 || staticPackageFuncName(pass, outer) != withStackName {
		return false
	}

	inner, ok := unparen(outer.Args[0]).(*ast.CallExpr)
	if !ok || len(inner.Args) != 1 || staticPackageFuncName(pass, inner) != unhandledValueErrorName {
		return false
	}

	return sameExpr(pass, inner.Args[0], tag)
}

// staticPackageFuncName returns the name of a package-qualified static
// function call. Calls through a function value, methods, builtins, and
// conversions return an empty name.
func staticPackageFuncName(pass *analysis.Pass, call *ast.CallExpr) string {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || pass.TypesInfo.Selections[sel] != nil {
		return ""
	}

	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok {
		return ""
	}
	if _, ok := pass.TypesInfo.ObjectOf(pkgIdent).(*types.PkgName); !ok {
		return ""
	}

	fn := typeutil.StaticCallee(pass.TypesInfo, call)
	if fn == nil || fn.Pkg() == nil {
		return ""
	}

	sig, ok := fn.Type().(*types.Signature)
	if !ok || sig.Recv() != nil {
		return ""
	}

	return fn.Name()
}

// sameExpr compares the switch tag with the value passed to the unhandled
// value constructor. Identifiers and selectors use type-checker object
// identity, so equal spellings in different scopes do not match. The common
// expression forms are compared structurally; unsupported forms fail closed.
func sameExpr(pass *analysis.Pass, left, right ast.Expr) bool {
	left = unparen(left)
	right = unparen(right)
	if left == right {
		return true
	}
	if left == nil || right == nil {
		return false
	}

	leftType := pass.TypesInfo.TypeOf(left)
	rightType := pass.TypesInfo.TypeOf(right)
	if leftType != nil && rightType != nil && !types.Identical(leftType, rightType) {
		return false
	}

	switch leftNode := left.(type) {
	case *ast.Ident:
		rightNode, ok := right.(*ast.Ident)
		if !ok {
			return false
		}
		leftObject := pass.TypesInfo.ObjectOf(leftNode)
		rightObject := pass.TypesInfo.ObjectOf(rightNode)
		return leftObject != nil && leftObject == rightObject

	case *ast.SelectorExpr:
		rightNode, ok := right.(*ast.SelectorExpr)
		if !ok || !sameSelectorObject(pass, leftNode, rightNode) {
			return false
		}
		return sameExpr(pass, leftNode.X, rightNode.X)

	case *ast.BasicLit:
		rightNode, ok := right.(*ast.BasicLit)
		return ok && leftNode.Kind == rightNode.Kind && leftNode.Value == rightNode.Value

	case *ast.UnaryExpr:
		rightNode, ok := right.(*ast.UnaryExpr)
		return ok && leftNode.Op == rightNode.Op && sameExpr(pass, leftNode.X, rightNode.X)

	case *ast.BinaryExpr:
		rightNode, ok := right.(*ast.BinaryExpr)
		return ok && leftNode.Op == rightNode.Op &&
			sameExpr(pass, leftNode.X, rightNode.X) && sameExpr(pass, leftNode.Y, rightNode.Y)

	case *ast.CallExpr:
		rightNode, ok := right.(*ast.CallExpr)
		if !ok || leftNode.Ellipsis.IsValid() != rightNode.Ellipsis.IsValid() ||
			!sameExpr(pass, leftNode.Fun, rightNode.Fun) || len(leftNode.Args) != len(rightNode.Args) {
			return false
		}
		for i := range leftNode.Args {
			if !sameExpr(pass, leftNode.Args[i], rightNode.Args[i]) {
				return false
			}
		}
		return true

	case *ast.IndexExpr:
		rightNode, ok := right.(*ast.IndexExpr)
		return ok && sameExpr(pass, leftNode.X, rightNode.X) && sameExpr(pass, leftNode.Index, rightNode.Index)

	case *ast.IndexListExpr:
		rightNode, ok := right.(*ast.IndexListExpr)
		if !ok || !sameExpr(pass, leftNode.X, rightNode.X) || len(leftNode.Indices) != len(rightNode.Indices) {
			return false
		}
		for i := range leftNode.Indices {
			if !sameExpr(pass, leftNode.Indices[i], rightNode.Indices[i]) {
				return false
			}
		}
		return true

	case *ast.StarExpr:
		rightNode, ok := right.(*ast.StarExpr)
		return ok && sameExpr(pass, leftNode.X, rightNode.X)

	case *ast.SliceExpr:
		rightNode, ok := right.(*ast.SliceExpr)
		return ok && leftNode.Slice3 == rightNode.Slice3 &&
			sameExpr(pass, leftNode.X, rightNode.X) &&
			sameOptionalExpr(pass, leftNode.Low, rightNode.Low) &&
			sameOptionalExpr(pass, leftNode.High, rightNode.High) &&
			sameOptionalExpr(pass, leftNode.Max, rightNode.Max)

	case *ast.TypeAssertExpr:
		rightNode, ok := right.(*ast.TypeAssertExpr)
		return ok && sameExpr(pass, leftNode.X, rightNode.X) &&
			sameOptionalExpr(pass, leftNode.Type, rightNode.Type)

	case *ast.CompositeLit:
		rightNode, ok := right.(*ast.CompositeLit)
		if !ok || !sameOptionalExpr(pass, leftNode.Type, rightNode.Type) || len(leftNode.Elts) != len(rightNode.Elts) {
			return false
		}
		for i := range leftNode.Elts {
			if !sameExpr(pass, leftNode.Elts[i], rightNode.Elts[i]) {
				return false
			}
		}
		return true

	case *ast.KeyValueExpr:
		rightNode, ok := right.(*ast.KeyValueExpr)
		return ok && sameExpr(pass, leftNode.Key, rightNode.Key) && sameExpr(pass, leftNode.Value, rightNode.Value)
	}

	return false
}

// sameSelectorObject compares the selected object, distinguishing a field
// selection from a package-qualified identifier. Field object identity alone
// is not enough: event.Kind and otherEvent.Kind select the same field object,
// so their receiver expressions must also be compared by sameExpr.
func sameSelectorObject(pass *analysis.Pass, left, right *ast.SelectorExpr) bool {
	leftSelection := pass.TypesInfo.Selections[left]
	rightSelection := pass.TypesInfo.Selections[right]
	if (leftSelection == nil) != (rightSelection == nil) {
		return false
	}
	if leftSelection != nil {
		return leftSelection.Obj() == rightSelection.Obj()
	}

	leftObject := pass.TypesInfo.ObjectOf(left.Sel)
	rightObject := pass.TypesInfo.ObjectOf(right.Sel)
	return leftObject != nil && leftObject == rightObject
}

func sameOptionalExpr(pass *analysis.Pass, left, right ast.Expr) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return sameExpr(pass, left, right)
}

func unparen(expr ast.Expr) ast.Expr {
	for {
		paren, ok := expr.(*ast.ParenExpr)
		if !ok {
			return expr
		}
		expr = paren.X
	}
}
