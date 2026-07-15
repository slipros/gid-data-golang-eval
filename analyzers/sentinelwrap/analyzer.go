// Package sentinelwrap implements GID-244 (gidsentinelwrap): mapping a boundary
// error to a sentinel is done by reassigning err then wrapping once, not by
// wrapping the sentinel directly in a guard branch that duplicates the outer
// errors.Wrap.
//
// GID-176 already blesses the reassign-then-Wrap pattern and treats every
// errors.Wrap as correct, so it does not catch the shape below — two branches
// over the same err with the same context message, one wrapping a sentinel and
// one wrapping err:
//
//	if err != nil {
//	    if IsNoResult(err) {
//	        return errors.Wrap(entity.ErrNoResult, "update key") // GID-244
//	    }
//	    return errors.Wrap(err, "update key")
//	}
//
// The fix collapses the two wraps into one:
//
//	if err != nil {
//	    if IsNoResult(err) {
//	        err = entity.ErrNoResult
//	    }
//	    return errors.Wrap(err, "update key")
//	}
//
// The match is deliberately narrow to avoid false positives (all required):
//   - a guard `if <pred>(err, ...) { return errors.Wrap(<staticErr>, "msg") }`
//     with no else, no init, and a body of EXACTLY one single-value return;
//   - <staticErr> is a package-level static error (or a named error literal),
//     not the err variable;
//   - <pred>(err, ...) is a call whose first error-typed argument is err;
//   - a mirror `return errors.Wrap(err, "msg")` in the SAME block, over the
//     same err object, with an IDENTICAL string-literal message.
//
// Only errors.Wrap is considered (not Wrapf); messages must be string literals.
// pkg/errors is detected by the import path github.com/pkg/errors. Generated
// code (ast.IsGenerated) is skipped.
package sentinelwrap

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
)

const ruleID = "GID-244"

// Analyzer — GID-244 with default settings (no exclusions).
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — methods exempted from the rule: "Function" / "Method" or
	// "Package.Function" / "Type.Method".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-244 analyzer.
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidsentinelwrap",
		Doc: ruleID + ": map a boundary error to a sentinel by reassign-then-wrap-once, " +
			"not by wrapping the sentinel in a guard branch that duplicates the outer errors.Wrap",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s)
		},
	}
}

func run(pass *analysis.Pass, s Settings) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if exclude.Match(s.Exclude, recvTypeName(fn), fn.Name.Name) {
				continue
			}
			checkFunc(pass, fn)
		}
	}
	return nil, nil
}

func checkFunc(pass *analysis.Pass, fn *ast.FuncDecl) {
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}
		const diagMessage = ruleID + ": a sentinel wrapped in a guard branch duplicates the outer errors.Wrap (same message). " +
			"Fix: reassign then wrap once: if IsNoResult(err) { err = ErrNoResult }; return errors.Wrap(err, ...)"
		for _, stmt := range block.List {
			guard := asSentinelGuard(pass, stmt)
			if guard == nil {
				continue
			}
			if hasMirrorWrap(pass, block, guard) {
				pass.Reportf(guard.pos, diagMessage)
			}
		}
		return true
	})
}

// guardInfo — the parts of a sentinel guard needed to find its mirror.
type guardInfo struct {
	pos    token.Pos    // the guard `if` position (where the diagnostic is reported)
	errObj types.Object // the err variable tested by the predicate
	msg    string       // the guard's Wrap message (string-literal Value)
}

// asSentinelGuard reports whether stmt is a guard of the form
// `if <pred>(err, ...) { return errors.Wrap(<staticErr>, "msg") }` and, if so,
// returns the err object and the message. Returns nil otherwise.
func asSentinelGuard(pass *analysis.Pass, stmt ast.Stmt) *guardInfo {
	ifStmt, ok := stmt.(*ast.IfStmt)
	if !ok || ifStmt.Init != nil || ifStmt.Else != nil {
		return nil
	}
	// Body must be exactly one single-value return.
	if len(ifStmt.Body.List) != 1 {
		return nil
	}
	ret, ok := ifStmt.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(ret.Results) != 1 {
		return nil
	}
	msg, ok := wrapStaticSentinel(pass, ret.Results[0])
	if !ok {
		return nil
	}
	errObj := errArgOfPredicate(pass, ifStmt.Cond)
	if errObj == nil {
		return nil
	}
	return &guardInfo{pos: ifStmt.Pos(), errObj: errObj, msg: msg}
}

// wrapStaticSentinel reports whether expr is `errors.Wrap(<staticErr>, "msg")`
// where <staticErr> is a static error and "msg" is a string literal; it returns
// the message value.
func wrapStaticSentinel(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	call, ok := expr.(*ast.CallExpr)
	if !ok || pkgErrorsCallName(pass, call) != "Wrap" || len(call.Args) != 2 {
		return "", false
	}
	if !isStaticError(pass, call.Args[0]) {
		return "", false
	}
	return stringLit(call.Args[1])
}

// errArgOfPredicate returns the object of the first error-typed identifier
// argument of a predicate call (e.g. IsNoResult(err), errors.Is(err, X)); nil
// when cond is not such a call.
func errArgOfPredicate(pass *analysis.Pass, cond ast.Expr) types.Object {
	call, ok := cond.(*ast.CallExpr)
	if !ok {
		return nil
	}
	for _, arg := range call.Args {
		id, ok := arg.(*ast.Ident)
		if !ok {
			continue
		}
		obj := objectOf(pass, id)
		if obj != nil && isErrorType(obj.Type()) {
			return obj
		}
	}
	return nil
}

// hasMirrorWrap reports whether block directly contains a mirror return
// `return errors.Wrap(err, "msg")` over guard.errObj with guard.msg.
func hasMirrorWrap(pass *analysis.Pass, block *ast.BlockStmt, guard *guardInfo) bool {
	for _, stmt := range block.List {
		ret, ok := stmt.(*ast.ReturnStmt)
		if !ok {
			continue
		}
		for _, res := range ret.Results {
			call, ok := res.(*ast.CallExpr)
			if !ok || pkgErrorsCallName(pass, call) != "Wrap" || len(call.Args) != 2 {
				continue
			}
			id, ok := call.Args[0].(*ast.Ident)
			if !ok || objectOf(pass, id) != guard.errObj {
				continue
			}
			msg, ok := stringLit(call.Args[1])
			if ok && msg == guard.msg {
				return true
			}
		}
	}
	return false
}

// ===== shared helpers =====

// pkgErrorsCallName returns the name of the github.com/pkg/errors function
// if call invokes it; otherwise "".
func pkgErrorsCallName(pass *analysis.Pass, call *ast.CallExpr) string {
	const pkgErrorsPath = "github.com/pkg/errors"
	fn := typeutil.Callee(pass.TypesInfo, call)
	f, ok := fn.(*types.Func)
	if !ok || f.Pkg() == nil {
		return ""
	}
	pkg := f.Pkg()
	if pkg.Path() != pkgErrorsPath {
		return ""
	}
	return f.Name()
}

// stringLit returns the value of a string-literal expression.
func stringLit(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	return lit.Value, true
}

// isStaticError: a package-level var of type error (ErrSome) or a composite
// literal / address of a named error type (BigError{}, &BigError{}).
func isStaticError(pass *analysis.Pass, expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident, *ast.SelectorExpr:
		obj := objectOf(pass, exprIdent(e))
		v, ok := obj.(*types.Var)
		if !ok || !isPackageLevel(v) {
			return false
		}
		return isErrorType(v.Type())
	case *ast.UnaryExpr:
		if e.Op == token.AND {
			if cl, ok := e.X.(*ast.CompositeLit); ok {
				return isNamedErrorType(pass, cl)
			}
		}
	case *ast.CompositeLit:
		return isNamedErrorType(pass, e)
	}
	return false
}

func isNamedErrorType(pass *analysis.Pass, cl *ast.CompositeLit) bool {
	tv, ok := pass.TypesInfo.Types[cl]
	if !ok {
		return false
	}
	t := tv.Type
	if _, ok := t.(*types.Named); !ok {
		return false
	}
	if isErrorType(t) {
		return true
	}
	return isErrorType(types.NewPointer(t))
}

// exprIdent extracts the target identifier from an Ident or SelectorExpr.
func exprIdent(e ast.Expr) *ast.Ident {
	switch x := e.(type) {
	case *ast.Ident:
		return x
	case *ast.SelectorExpr:
		return x.Sel
	}
	return nil
}

func objectOf(pass *analysis.Pass, id *ast.Ident) types.Object {
	if id == nil {
		return nil
	}
	if obj := pass.TypesInfo.Uses[id]; obj != nil {
		return obj
	}
	return pass.TypesInfo.Defs[id]
}

func isPackageLevel(v *types.Var) bool {
	pkg := v.Pkg()
	if pkg == nil {
		return false
	}
	return v.Parent() != nil && v.Parent() == pkg.Scope()
}

func isErrorType(t types.Type) bool {
	if t == nil {
		return false
	}
	errObj := types.Universe.Lookup("error")
	errType := errObj.Type()
	errIface, ok := errType.Underlying().(*types.Interface)
	if !ok {
		return false
	}
	return types.Implements(t, errIface)
}

// recvTypeName returns the name of fn's receiver type ("" for a free function).
func recvTypeName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}
	t := fn.Recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	if ident, ok := t.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}
