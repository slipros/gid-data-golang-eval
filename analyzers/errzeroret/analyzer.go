// Package errzeroret implements rule GID-243: on error, every non-error
// result must be nil/zero.
//
// Absolute rule (no exceptions): a return statement with two or more results
// whose LAST result is a DEFINITELY non-nil error must not carry a non-zero
// value in any of the preceding, non-error results.
//
// "Definitely non-nil error" — the last operand is either:
//   - (a) a constructing call: status.Error/status.Errorf
//     (google.golang.org/grpc/status), errors.New/Wrap/Wrapf/Errorf/
//     WithStack/WithMessage/WithMessagef (github.com/pkg/errors), or
//     fmt.Errorf; or
//   - (b) a plain identifier, and the return is lexically inside an
//     `if <e> != nil { ... }` block that guards that same <e>.
//
// "Zero" — nil, the boolean false, a zero basic literal (0, 0.0, ""), or an
// empty composite literal T{}. A variable, a populated composite literal
// (T{A: 1}), the address of any composite literal (&T{}, &T{A: 1} — a
// non-nil pointer; the zero VALUE of a pointer is nil, not an address), or a
// call — none of these count as zero.
//
// Deliberately NOT reported: an unconditional final forward
// (return resp, err, where err is a plain variable neither guarded by an
// enclosing `if err != nil` nor a constructing call) — the legitimate
// interceptor/pass-through shape used at the end of a wrapping function.
//
// Generated code (ast.IsGenerated) is skipped.
package errzeroret

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

const ruleID = "GID-243"

// Analyzer — rule GID-243 (no settings: the rule is absolute, no exceptions).
var Analyzer = &analysis.Analyzer{
	Name: "giderrzeroret",
	Doc: ruleID + ": on error, non-error results must be nil/zero. " +
		"Fix: return nil / T{} alongside the error",
	Run: run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			walkBlock(pass, fn.Body, nil)
		}
	}
	return nil, nil
}

// walkBlock walks the statements of block, threading the set of error
// objects the current scope has proven non-nil via an enclosing
// `if <e> != nil { ... }` guard.
func walkBlock(pass *analysis.Pass, block *ast.BlockStmt, guarded map[types.Object]bool) {
	if block == nil {
		return
	}
	for _, stmt := range block.List {
		walkStmt(pass, stmt, guarded)
	}
}

func walkStmt(pass *analysis.Pass, stmt ast.Stmt, guarded map[types.Object]bool) {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		walkBlock(pass, s, guarded)
	case *ast.IfStmt:
		bodyGuard := guarded
		if obj, ok := nonNilGuardObject(pass, s.Cond); ok {
			bodyGuard = withGuard(guarded, obj)
		}
		walkBlock(pass, s.Body, bodyGuard)
		switch e := s.Else.(type) {
		case *ast.BlockStmt:
			walkBlock(pass, e, guarded)
		case *ast.IfStmt:
			walkStmt(pass, e, guarded)
		}
	case *ast.ForStmt:
		walkBlock(pass, s.Body, guarded)
	case *ast.RangeStmt:
		walkBlock(pass, s.Body, guarded)
	case *ast.SwitchStmt:
		walkCaseClauses(pass, s.Body, guarded)
	case *ast.TypeSwitchStmt:
		walkCaseClauses(pass, s.Body, guarded)
	case *ast.SelectStmt:
		for _, c := range s.Body.List {
			if cc, ok := c.(*ast.CommClause); ok {
				for _, st := range cc.Body {
					walkStmt(pass, st, guarded)
				}
			}
		}
	case *ast.LabeledStmt:
		walkStmt(pass, s.Stmt, guarded)
	case *ast.ReturnStmt:
		checkReturn(pass, s, guarded)
	}
}

func walkCaseClauses(pass *analysis.Pass, body *ast.BlockStmt, guarded map[types.Object]bool) {
	if body == nil {
		return
	}
	for _, c := range body.List {
		if cc, ok := c.(*ast.CaseClause); ok {
			for _, st := range cc.Body {
				walkStmt(pass, st, guarded)
			}
		}
	}
}

func withGuard(guarded map[types.Object]bool, obj types.Object) map[types.Object]bool {
	out := make(map[types.Object]bool, len(guarded)+1)
	for k := range guarded {
		out[k] = true
	}
	out[obj] = true
	return out
}

// nonNilGuardObject reports the object of an `if <e> != nil` condition, when
// <e> is of type error.
func nonNilGuardObject(pass *analysis.Pass, cond ast.Expr) (types.Object, bool) {
	bin, ok := cond.(*ast.BinaryExpr)
	if !ok || bin.Op != token.NEQ {
		return nil, false
	}
	var candidate ast.Expr
	switch {
	case isNilIdent(bin.Y):
		candidate = bin.X
	case isNilIdent(bin.X):
		candidate = bin.Y
	default:
		return nil, false
	}
	idExpr, ok := candidate.(*ast.Ident)
	if !ok {
		return nil, false
	}
	obj := pass.TypesInfo.Uses[idExpr]
	if obj == nil || !isErrorType(obj.Type()) {
		return nil, false
	}
	return obj, true
}

func isNilIdent(e ast.Expr) bool {
	id, ok := e.(*ast.Ident)
	return ok && id.Name == "nil"
}

// checkReturn reports ret if its last result is a definitely non-nil error
// and at least one of the preceding results is not a zero value.
func checkReturn(pass *analysis.Pass, ret *ast.ReturnStmt, guarded map[types.Object]bool) {
	if len(ret.Results) < 2 {
		return
	}
	last := ret.Results[len(ret.Results)-1]
	if !isDefinitelyNonNilError(pass, last, guarded) {
		return
	}
	for _, res := range ret.Results[:len(ret.Results)-1] {
		if !isZeroValue(res) {
			pass.Reportf(ret.Pos(),
				"%s: on error, non-error results must be nil/zero (got a non-zero value alongside a non-nil "+
					"error). Fix: return nil / T{} alongside the error",
				ruleID)
			return
		}
	}
}

func isDefinitelyNonNilError(pass *analysis.Pass, expr ast.Expr, guarded map[types.Object]bool) bool {
	switch e := expr.(type) {
	case *ast.CallExpr:
		return isConstructingErrorCall(pass, e)
	case *ast.Ident:
		if e.Name == "nil" {
			return false
		}
		obj := pass.TypesInfo.Uses[e]
		return obj != nil && guarded[obj]
	}
	return false
}

// isConstructingErrorCall reports whether call constructs a new error value:
// status.Error/Errorf, github.com/pkg/errors' New/Wrap/Wrapf/Errorf/
// WithStack/WithMessage/WithMessagef, or fmt.Errorf.
func isConstructingErrorCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	fn := typeutil.Callee(pass.TypesInfo, call)
	f, ok := fn.(*types.Func)
	if !ok {
		return false
	}
	pkg := f.Pkg()
	if pkg == nil {
		return false
	}
	switch pkg.Path() {
	case "google.golang.org/grpc/status":
		return f.Name() == "Error" || f.Name() == "Errorf"
	case "github.com/pkg/errors":
		switch f.Name() {
		case "New", "Wrap", "Wrapf", "Errorf", "WithStack", "WithMessage", "WithMessagef":
			return true
		}
		return false
	case "fmt":
		return f.Name() == "Errorf"
	}
	return false
}

// isZeroValue reports whether expr is a zero-value literal: nil, false, a
// zero basic literal (0, 0.0, ""), or an empty composite literal T{}. A
// variable, a populated composite literal, the address of any composite
// literal (&T{}/&T{A: 1} — a non-nil pointer; the pointer zero value is nil,
// not an address), and a call are all NOT zero.
func isZeroValue(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name == "nil" || e.Name == "false"
	case *ast.BasicLit:
		return isZeroBasicLit(e)
	case *ast.CompositeLit:
		return len(e.Elts) == 0
	}
	return false
}

func isZeroBasicLit(lit *ast.BasicLit) bool {
	switch lit.Kind {
	case token.INT:
		v := strings.ReplaceAll(lit.Value, "_", "")
		n, err := strconv.ParseInt(v, 0, 64)
		return err == nil && n == 0
	case token.FLOAT:
		v := strings.ReplaceAll(lit.Value, "_", "")
		f, err := strconv.ParseFloat(v, 64)
		return err == nil && f == 0
	case token.STRING:
		s, err := strconv.Unquote(lit.Value)
		return err == nil && s == ""
	}
	return false
}

func isErrorType(t types.Type) bool {
	if t == nil {
		return false
	}
	errObj := types.Universe.Lookup("error")
	errType := errObj.Type()
	iface, ok := errType.Underlying().(*types.Interface)
	if !ok {
		return false
	}
	return types.Implements(t, iface)
}
