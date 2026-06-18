// Package errwrap implements the per-layer error handling rules:
//
//   - GID-176 (giderrwrap): errors from outside are wrapped with errors.Wrap.
//     At the application boundary (/client/** and /dal/repository) an error
//     from an external call is neither passed through as is (return err) nor
//     enriched without context (WithStack/WithMessage) — Wrap is required:
//     it collects the stack AND adds the mandatory context. The boundary is an
//     interface-method call (an injected external dependency, e.g.
//     c.conn.Select(...)); a call to a local package function (a pure SQL
//     builder build.Select(...), a concrete-type method) is not the boundary
//     and may be enriched with WithMessage. Inside the
//     application (/domain/**), Wrap is forbidden for an already received
//     non-static error (the stack was collected at the boundary) — context is
//     added with WithMessage. Returning a static error (package-level var, a
//     named error type) at the boundary is not a GID-176 violation (that is GID-177 territory).
//
//   - GID-177 (gidstaticerr): static errors are wrapped with WithStack.
//     Returning a static error (a package-level error var ErrSome or a
//     composite literal / address of a named error type BigError{}/&BigError{})
//     without a wrapper lacks a stack — errors.WithStack is required (or
//     errors.Wrap if context is needed). A wrapped error (WithStack/Wrap) is fine.
//     The var declarations themselves are not touched (they are not returns).
//
// pkg/errors is detected by the import path github.com/pkg/errors.
// Generated code (ast.IsGenerated) is skipped.
package errwrap

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const (
	ruleIDWrap   = "GID-176"
	ruleIDStatic = "GID-177"
)

// boundaryScopes — boundary layers for GID-176 (part 1): an external call.
var boundaryScopes = [][]string{
	{"client"},
	{"dal", "repository"},
}

// WrapAnalyzer — GID-176 with default settings (no exclusions).
var WrapAnalyzer = NewWrapAnalyzer(Settings{})

// StaticAnalyzer — GID-177 with default settings (no exclusions).
var StaticAnalyzer = NewStaticAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — names of constructor/error exclusions that collect
	// the stack themselves (for example, gderror.NewUnhandledValueError):
	// "Function" or "Package.Function".
	Exclude []string `json:"exclude"`
}

// NewWrapAnalyzer builds the GID-176 analyzer.
func NewWrapAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "giderrwrap",
		Doc: ruleIDWrap + ": errors from outside are wrapped with errors.Wrap; " +
			"inside the app, wrap a non-static error with WithMessage, not Wrap",
		Run: runWrap,
	}
}

// NewStaticAnalyzer builds the GID-177 analyzer.
func NewStaticAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidstaticerr",
		Doc:  ruleIDStatic + ": static errors are wrapped with errors.WithStack on return. Fix: wrap with errors.WithStack",
		Run: func(pass *analysis.Pass) (any, error) {
			return runStatic(pass, s)
		},
	}
}

// ===== GID-176 =====

func runWrap(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()
	boundary := inBoundary(pkgPath)
	domain := pathseg.Contains(pkgPath, "domain")
	if !boundary && !domain {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if boundary && funcReturnsError(pass, fn) {
				checkBoundaryPassThrough(pass, fn)
			}
		}
		if domain {
			checkDomainWrap(pass, file)
		}
	}
	return nil, nil
}

// checkBoundaryPassThrough — GID-176 part 1: at the boundary a non-static
// error from a call must not be passed through without Wrap.
func checkBoundaryPassThrough(pass *analysis.Pass, fn *ast.FuncDecl) {
	callErrs := localCallErrors(pass, fn)
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		for _, res := range ret.Results {
			expr := res
			// errors.WithStack(err) / errors.WithMessage(err) — a wrapper without context.
			if call, ok := expr.(*ast.CallExpr); ok {
				name := pkgErrorsCallName(pass, call)
				if name == "WithStack" || name == "WithMessage" {
					if len(call.Args) > 0 && isLocalCallErr(pass, call.Args[0], callErrs) {
						pass.Reportf(call.Pos(),
							"%s: an error from the app boundary must be wrapped with errors.Wrap (%s adds no context). "+
								"Fix: collect stack and context; to map a sentinel, reassign then wrap once: "+
								"if IsNoResult(err) { err = ErrNoResult }; return errors.Wrap(err, ...)",
							ruleIDWrap, name)
					}
					continue
				}
				// errors.Wrap / any other call — fine (Wrap is already correct).
				continue
			}
			if isLocalCallErr(pass, expr, callErrs) {
				pass.Reportf(expr.Pos(),
					"%s: an error from the app boundary must be wrapped with errors.Wrap. "+
						"Fix: collect stack and context; to map a sentinel, reassign then wrap once: "+
						"if IsNoResult(err) { err = ErrNoResult }; return errors.Wrap(err, ...)",
					ruleIDWrap)
			}
		}
		return true
	})
}

// checkDomainWrap — GID-176 part 2: in /domain/** wrapping a non-static error with Wrap is forbidden.
func checkDomainWrap(pass *analysis.Pass, file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if pkgErrorsCallName(pass, call) != "Wrap" {
			return true
		}
		if len(call.Args) == 0 {
			return true
		}
		// A static error (model.ErrX, &BigError{}) — Wrap is allowed.
		if isStaticError(pass, call.Args[0]) {
			return true
		}
		// A non-static one (a local variable from a call, etc.) — forbidden.
		if isErrorExpr(pass, call.Args[0]) {
			pass.Reportf(call.Pos(),
				"%s: the stack is already collected at the boundary. Fix: use errors.WithMessage instead of errors.Wrap for an incoming error",
				ruleIDWrap)
		}
		return true
	})
}

// ===== GID-177 =====

func runStatic(pass *analysis.Pass, s Settings) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				ret, ok := n.(*ast.ReturnStmt)
				if !ok {
					return true
				}
				for _, res := range ret.Results {
					checkStaticReturn(pass, s, res)
				}
				return true
			})
		}
	}
	return nil, nil
}

func checkStaticReturn(pass *analysis.Pass, s Settings, expr ast.Expr) {
	// Already wrapped (WithStack/Wrap/another pkg/errors call) — fine.
	if call, ok := expr.(*ast.CallExpr); ok {
		name := pkgErrorsCallName(pass, call)
		if name == "WithStack" || name == "Wrap" || name == "Wrapf" {
			return
		}
		// An excluded constructor (collects the stack itself) — fine.
		if isExcludedCtor(pass, call, s.Exclude) {
			return
		}
		return
	}
	if isStaticError(pass, expr) {
		pass.Reportf(expr.Pos(),
			"%s: a static error is returned without a stack. Fix: wrap with errors.WithStack (or errors.Wrap if you need context)",
			ruleIDStatic)
	}
}

// ===== shared helpers =====

func inBoundary(pkgPath string) bool {
	for _, scope := range boundaryScopes {
		if pathseg.Contains(pkgPath, scope...) {
			return true
		}
	}
	return false
}

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

func isExcludedCtor(pass *analysis.Pass, call *ast.CallExpr, list []string) bool {
	fn := typeutil.Callee(pass.TypesInfo, call)
	f, ok := fn.(*types.Func)
	if !ok {
		return false
	}
	pkgName := ""
	if pkg := f.Pkg(); pkg != nil {
		pkgName = pkg.Name()
	}
	return exclude.Match(list, pkgName, f.Name())
}

func funcReturnsError(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	obj, ok := pass.TypesInfo.Defs[fn.Name].(*types.Func)
	if !ok {
		return false
	}
	sig, ok := obj.Type().(*types.Signature)
	if !ok {
		return false
	}
	results := sig.Results()
	for v := range results.Variables() {
		if isErrorType(v.Type()) {
			return true
		}
	}
	return false
}

// localCallErrors collects the function's local variables whose value comes
// from an interface-method call and implements error (err := c.conn.f();
// a, err := c.conn.f()). The application boundary is an interface-method call
// on an injected external dependency; an error from a local package function
// (a pure SQL builder, etc.) or a concrete-type method is not a boundary error.
func localCallErrors(pass *analysis.Pass, fn *ast.FuncDecl) map[types.Object]struct{} {
	out := map[types.Object]struct{}{}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		// The source is exactly one call on the right-hand side.
		if len(assign.Rhs) != 1 {
			return true
		}
		call, ok := assign.Rhs[0].(*ast.CallExpr)
		if !ok {
			return true
		}
		// Only an interface-method call is the boundary.
		if !isInterfaceMethodCall(pass, call) {
			return true
		}
		for _, lhs := range assign.Lhs {
			id, ok := lhs.(*ast.Ident)
			if !ok || id.Name == "_" {
				continue
			}
			obj := objectOf(pass, id)
			if obj == nil || !isErrorType(obj.Type()) {
				continue
			}
			out[obj] = struct{}{}
		}
		return true
	})
	return out
}

// isInterfaceMethodCall reports whether call invokes a method on a value of
// interface type — the injected external dependency at the boundary, e.g.
// c.conn.Select(...). A qualified package function (build.Select(...)) is not a
// selection, and a method on a concrete type has a non-interface receiver;
// neither is an interface-method call.
func isInterfaceMethodCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	selection, ok := pass.TypesInfo.Selections[sel]
	if !ok {
		// pkg.Func — a qualified identifier, not a method selection.
		return false
	}
	if selection.Kind() != types.MethodVal {
		return false
	}
	recv := selection.Recv()
	if recv == nil {
		return false
	}
	_, isIface := recv.Underlying().(*types.Interface)
	return isIface
}

func isLocalCallErr(pass *analysis.Pass, expr ast.Expr, callErrs map[types.Object]struct{}) bool {
	id, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	obj := objectOf(pass, id)
	if obj == nil {
		return false
	}
	_, ok = callErrs[obj]
	return ok
}

// isErrorExpr reports that the expression has type error and is not
// a static error (package-level var / named error literal).
func isErrorExpr(pass *analysis.Pass, expr ast.Expr) bool {
	tv, ok := pass.TypesInfo.Types[expr]
	if !ok {
		return false
	}
	return isErrorType(tv.Type) && !isStaticError(pass, expr)
}

// isStaticError: a package-level var of type error (ErrSome) or a
// composite literal / address of a named error type (BigError{}, &BigError{}).
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
	// The type must implement error by itself or via a pointer.
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
	pkgScope := pkg.Scope()
	return v.Parent() != nil && v.Parent() == pkgScope
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
