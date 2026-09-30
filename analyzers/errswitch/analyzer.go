// Package errswitch implements rule GID-279: an error is first checked with an
// `if err != nil`, and only then classified — a switch over an error variable
// sits inside that `if`.
//
//	id, err := client.Upload(ctx, body)
//
//	switch {                       // GID-279: no `if err != nil` before the switch
//	case err == nil:
//	    return id, nil
//	case errors.Is(err, ErrRate):
//	    return 0, errors.Wrap(ErrUnavailable, "upload")
//	default:
//	    return 0, err
//	}
//
// Without the guard the success path becomes one of the branches that classify
// failures: the reader searches the whole switch for what happens on success,
// and a `default` returning the raw error (GID-176) comes for free. With the
// guard the classes are narrowed where the error is known to exist:
//
//	id, err := client.Upload(ctx, body)
//	if err != nil {
//	    switch {
//	    case errors.Is(err, ErrRate):
//	        err = ErrUnavailable
//	    }
//
//	    return 0, errors.Wrap(err, "upload")
//	}
//
//	return id, nil
//
// `errors.Is`, `errors.As` and a check of a call such as `ctx.Err() != nil` are
// fine as cases — inside the guard.
//
// Detect: a `switch` that examines an error VARIABLE — a local variable or a
// field of an error type, met in a case condition of a tagless
// switch (`err == nil`, `errors.Is(err, X)`, `errors.As(err, &t)`, `ready &&
// err == nil`), as the tag (`switch err { case ErrX: }`) or as the operand of a
// type switch (`switch e := err.(type)`) — and is not inside an `if` body that
// has checked that same variable: `if err != nil { … }` (also as a conjunct,
// `err != nil && ready`, and with an init statement, `if err := f(); err != nil`),
// or the else branch of `if err == nil { … } else { … }`. Not judged: a call
// result (`ctx.Err() != nil`), a package-level sentinel (`service.ErrX`) on its
// own, a PARAMETER (the caller has checked it: `errorResponse(ctx, f, err)`,
// `isRetryable(err error) bool` classify an error that is already known to
// exist), and a switch that mentions no error variable. Reported on the `switch`
// keyword. Generated code and _test.go files are not judged: a test's switch
// over a mocked call is scaffolding.
package errswitch

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const ruleID = "GID-279"

// Analyzer — rule GID-279.
var Analyzer = &analysis.Analyzer{
	Name: "giderrswitch",
	Doc: ruleID + ": a switch over an error variable sits inside if err != nil. " +
		"Fix: check the error with if first, if err != nil { switch … ; return …, errors.Wrap(err, \"op\") }, " +
		"and return the result after the check",
	Requires: astwalk.Requires,
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	skip := func(file *ast.File) bool {
		return ast.IsGenerated(file) || srcfile.IsTest(pass, file)
	}
	filter := []ast.Node{
		(*ast.FuncDecl)(nil), (*ast.FuncLit)(nil),
		(*ast.IfStmt)(nil), (*ast.SwitchStmt)(nil), (*ast.TypeSwitchStmt)(nil),
	}

	params := map[*types.Var]bool{}
	markParam := func(v *types.Var) { params[v] = true }
	var enclosing []*ast.IfStmt
	astwalk.Around(pass, filter, skip, func(_ *ast.File, n ast.Node, push bool) bool {
		switch fn := n.(type) {
		case *ast.FuncDecl:
			if push {
				collectParams(pass, markParam, fn.Recv, fn.Type)
			}

			return true
		case *ast.FuncLit:
			if push {
				collectParams(pass, markParam, nil, fn.Type)
			}

			return true
		}
		if ifStmt, ok := n.(*ast.IfStmt); ok {
			if push {
				enclosing = append(enclosing, ifStmt)
			} else {
				enclosing = enclosing[:len(enclosing)-1]
			}

			return true
		}
		if !push {
			return true
		}
		keys := examinedErrors(pass, params, n)
		if len(keys) == 0 {
			return true
		}
		for _, key := range keys {
			if guarded(enclosing, n, key) {
				return true
			}
		}
		pass.Reportf(n.Pos(),
			"%s: the error %q is classified with a switch that is not inside `if %s != nil`. "+
				"Fix: check the error with if first — if %s != nil { switch … ; return …, errors.Wrap(%s, \"op\") } — "+
				"and return the result after the check",
			ruleID, keys[0], keys[0], keys[0], keys[0])

		return true
	})

	return nil, nil
}

// examinedErrors returns the error variables (as written) that a switch
// examines: in case conditions, the tag, or the type-switch operand.
func examinedErrors(pass *analysis.Pass, params map[*types.Var]bool, n ast.Node) []string {
	var roots []ast.Node
	switch sw := n.(type) {
	case *ast.SwitchStmt:
		if sw.Tag != nil {
			roots = append(roots, sw.Tag)
		}
		for _, stmt := range sw.Body.List {
			if clause, ok := stmt.(*ast.CaseClause); ok && sw.Tag == nil {
				for _, cond := range clause.List {
					roots = append(roots, cond)
				}
			}
		}
	case *ast.TypeSwitchStmt:
		if assert := typeAssert(sw); assert != nil {
			roots = append(roots, assert.X)
		}
	}

	var keys []string
	seen := map[string]bool{}
	for _, root := range roots {
		ast.Inspect(root, func(node ast.Node) bool {
			expr, ok := node.(ast.Expr)
			if !ok {
				return true
			}
			if key, ok := errorVariable(pass, params, expr); ok && !seen[key] {
				seen[key] = true
				keys = append(keys, key)
			}

			return true
		})
	}

	return keys
}

// typeAssert returns the `x.(type)` guard of a type switch.
func typeAssert(sw *ast.TypeSwitchStmt) *ast.TypeAssertExpr {
	var guard ast.Expr
	switch s := sw.Assign.(type) {
	case *ast.ExprStmt:
		guard = s.X
	case *ast.AssignStmt:
		if len(s.Rhs) == 1 {
			guard = s.Rhs[0]
		}
	}
	assert, ok := guard.(*ast.TypeAssertExpr)
	if !ok {
		return nil
	}

	return assert
}

// collectParams records the parameters (and the receiver) of a function: an
// error received from the caller has been checked there.
func collectParams(pass *analysis.Pass, mark func(*types.Var), recv *ast.FieldList, sig *ast.FuncType) {
	for _, list := range []*ast.FieldList{recv, sig.Params} {
		if list == nil {
			continue
		}
		for _, field := range list.List {
			for _, name := range field.Names {
				if v, ok := pass.TypesInfo.Defs[name].(*types.Var); ok {
					mark(v)
				}
			}
		}
	}
}

// errorVariable reports whether expr is a local variable or a field of an error
// type, and returns it as written. A package-level variable (a sentinel,
// `service.ErrX`), a parameter and a call result are not judged.
func errorVariable(pass *analysis.Pass, params map[*types.Var]bool, expr ast.Expr) (string, bool) {
	var obj types.Object
	switch e := expr.(type) {
	case *ast.Ident:
		obj = pass.TypesInfo.Uses[e]
	case *ast.SelectorExpr:
		if sel, ok := pass.TypesInfo.Selections[e]; ok && sel.Kind() == types.FieldVal {
			obj = sel.Obj()
		}
	}
	variable, ok := obj.(*types.Var)
	if !ok || params[variable] || !implementsError(variable.Type()) {
		return "", false
	}
	if pkg := variable.Pkg(); pkg != nil && variable.Parent() == pkg.Scope() {
		return "", false
	}

	return types.ExprString(expr), true
}

// implementsError reports whether t is an interface type that implements error:
// `error` itself or an interface embedding it. A concrete error type is not an
// examined error — it is the target of errors.As(err, &target).
func implementsError(t types.Type) bool {
	iface, ok := t.Underlying().(*types.Interface)
	if !ok {
		return false
	}
	errObj := types.Universe.Lookup("error")
	errType := errObj.Type()
	errIface, ok := errType.Underlying().(*types.Interface)

	return ok && types.Implements(iface, errIface)
}

// guarded reports whether the node sits in the body of an enclosing `if` that
// has checked key: `if key != nil` (alone or as a conjunct of &&), or in the
// else branch of `if key == nil`.
func guarded(enclosing []*ast.IfStmt, n ast.Node, key string) bool {
	for _, ifStmt := range enclosing {
		if within(n, ifStmt.Body) && hasNilCheck(ifStmt.Cond, key, token.NEQ) {
			return true
		}
		if ifStmt.Else != nil && within(n, ifStmt.Else) && isNilCheck(ifStmt.Cond, key, token.EQL) {
			return true
		}
	}

	return false
}

func within(n, outer ast.Node) bool {
	return outer.Pos() <= n.Pos() && n.End() <= outer.End()
}

// hasNilCheck reports whether cond is `key <op> nil` or an && chain with such a conjunct.
func hasNilCheck(cond ast.Expr, key string, op token.Token) bool {
	cond = ast.Unparen(cond)
	if bin, ok := cond.(*ast.BinaryExpr); ok && bin.Op == token.LAND {
		return hasNilCheck(bin.X, key, op) || hasNilCheck(bin.Y, key, op)
	}

	return isNilCheck(cond, key, op)
}

// isNilCheck reports whether cond is exactly `key <op> nil` (nil on either side).
func isNilCheck(cond ast.Expr, key string, op token.Token) bool {
	bin, ok := ast.Unparen(cond).(*ast.BinaryExpr)
	if !ok || bin.Op != op {
		return false
	}
	switch {
	case isNilIdent(bin.Y):
		return types.ExprString(bin.X) == key
	case isNilIdent(bin.X):
		return types.ExprString(bin.Y) == key
	}

	return false
}

func isNilIdent(expr ast.Expr) bool {
	id, ok := ast.Unparen(expr).(*ast.Ident)

	return ok && id.Name == "nil"
}
