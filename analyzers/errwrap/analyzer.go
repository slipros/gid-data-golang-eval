// Package errwrap implements the per-layer error handling rules:
//
//   - GID-176 (giderrwrap): every external call — a third-party library, a DB
//     connection, an HTTP/Kafka client, the standard library — has its error
//     wrapped with errors.Wrap, in ANY layer. Two call shapes count as an
//     external call:
//     (a) a direct call to a function/method whose declaring package lies
//     outside the current module (stdlib counts as external too) — this is
//     checked everywhere, regardless of layer;
//     (b) an interface-method call (an injected external dependency, e.g.
//     c.conn.Select(...)) inside the scoped boundary layers — /client/**,
//     /dal/repository, /event/** (a Kafka producer/consumer talks to an
//     external system). A call to a local package function (a pure SQL
//     builder build.Select(...)) or a concrete-type method that is not (a) is
//     not a boundary call and may be enriched with WithMessage or passed
//     through as is.
//     An external error is neither passed through as is (return err) nor
//     enriched without context (WithStack/WithMessage) — Wrap is required: it
//     collects the stack AND adds the mandatory context. To map it to a
//     sentinel, reassign then wrap once (the sentinel-then-Wrap pattern stays
//     legal). Inside the application (/domain/**), Wrap is forbidden for a
//     same-module non-static error (a same-module call result or a function
//     parameter) — its stack, if any, was already collected upstream; context
//     is added with WithMessage. Wrap of an external-call error inside
//     /domain/** is required, not forbidden — the domain may be the first
//     place that calls out (e.g. a DB connection reached directly from a
//     service). Returning a static error (package-level var, a named error
//     type) is not a GID-176 violation (that is GID-177 territory).
//
//   - GID-177 (gidstaticerr): static errors are wrapped with WithStack.
//     Returning a static error (a package-level error var ErrSome or a
//     composite literal / address of a named error type BigError{}/&BigError{})
//     without a wrapper lacks a stack — errors.WithStack is required (or
//     errors.Wrap if context is needed). A wrapped error (WithStack/Wrap) is fine.
//     The var declarations themselves are not touched (they are not returns).
//
//   - GID-237 (gidwithmessage): in /domain/service, errors.WithMessage and
//     errors.WithMessagef are banned — a service converts the error and wraps
//     it with errors.WithStack; adding a message to an incoming error belongs
//     to /domain/usecase.
//
// pkg/errors is detected by the import path github.com/pkg/errors.
// Generated code (ast.IsGenerated) is skipped.
package errwrap

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const (
	ruleIDWrap           = "GID-176"
	ruleIDStatic         = "GID-177"
	ruleIDServiceMessage = "GID-237"

	pkgErrorsPath = "github.com/pkg/errors"
)

// boundaryScopes — layers where an interface-method call (an injected
// external dependency) counts as a GID-176 boundary call: /client/**,
// /dal/repository, /event/** (a Kafka producer/consumer talks to an external
// system). A direct call into a package outside the current module (see
// isExternalCall) is a boundary call everywhere, regardless of this list.
const (
	// errSourceExternal — the value comes from a direct call to a function/method
	// whose declaring package lies outside the current module (mechanism a):
	// a boundary call in ANY layer.
	errSourceExternal errSource = iota + 1
	// errSourceInterface — the value comes from an interface-method call on an
	// injected dependency (mechanism b): a boundary call only inside boundaryScopes.
	errSourceInterface
)

var boundaryScopes = [][]string{
	{"client"},
	{"dal", "repository"},
	{"event"},
}

// serviceMessageScopes — layers where GID-237 bans errors.WithMessage/WithMessagef.
var serviceMessageScopes = [][]string{
	{"domain", "service"},
}

// WrapAnalyzer — GID-176 with default settings (no exclusions).
var WrapAnalyzer = NewWrapAnalyzer(Settings{})

// StaticAnalyzer — GID-177 with default settings (no exclusions).
var StaticAnalyzer = NewStaticAnalyzer(Settings{})

// ServiceMessageAnalyzer — GID-237 with default settings (no exclusions).
var ServiceMessageAnalyzer = NewServiceMessageAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — names of constructor/error exclusions that collect
	// the stack themselves (for example, gderror.NewUnhandledValueError),
	// or of methods exempted from a per-function rule (GID-237): "Function"
	// / "Method" or "Package.Function" / "Type.Method".
	Exclude []string `json:"exclude"`
}

// NewWrapAnalyzer builds the GID-176 analyzer.
func NewWrapAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "giderrwrap",
		Doc: ruleIDWrap + ": every external call's error is wrapped with errors.Wrap, in any layer; " +
			"inside /domain/**, wrap a same-module non-static error with WithMessage, not Wrap",
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

// NewServiceMessageAnalyzer builds the GID-237 analyzer.
func NewServiceMessageAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidwithmessage",
		Doc: ruleIDServiceMessage + ": errors.WithMessage is not used in a service. Fix: convert the error " +
			"to a model error and wrap with errors.WithStack; adding message context belongs to usecase",
		Run: func(pass *analysis.Pass) (any, error) {
			return runServiceMessage(pass, s)
		},
	}
}

// ===== GID-176 =====

// errSource — why a local error variable counts toward the GID-176 boundary.
type errSource int

func runWrap(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()
	boundary := inBoundary(pkgPath)
	domain := pathseg.HasLayer(pkgPath, "domain")
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if !funcReturnsError(pass, fn) {
				continue
			}
			callErrs := classifyCallErrors(pass, fn, boundary)
			checkMustWrap(pass, fn, callErrs)
			if domain {
				checkDomainWrapBan(pass, fn, callErrs)
			}
		}
	}
	return nil, nil
}

// checkMustWrap — GID-176: a tracked external-call error (see classifyCallErrors)
// must not be passed through or enriched without a stack (WithStack/WithMessage);
// only errors.Wrap collects the stack and adds context. Applies in every layer:
// classifyCallErrors decides per-error whether it is tracked at all.
func checkMustWrap(pass *analysis.Pass, fn *ast.FuncDecl, callErrs map[types.Object]errSource) {
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
					if len(call.Args) > 0 && isTrackedCallErr(pass, call.Args[0], callErrs) {
						pass.Reportf(call.Pos(),
							"%s: an error from an external call must be wrapped with errors.Wrap (%s adds no context). "+
								"Fix: collect stack and context; to map a sentinel, reassign then wrap once: "+
								"if IsNoResult(err) { err = ErrNoResult }; return errors.Wrap(err, ...)",
							ruleIDWrap, name)
					}
					continue
				}
				// errors.Wrap / any other call — fine (Wrap is already correct).
				continue
			}
			if isTrackedCallErr(pass, expr, callErrs) {
				pass.Reportf(expr.Pos(),
					"%s: an error from an external call must be wrapped with errors.Wrap. "+
						"Fix: collect stack and context; to map a sentinel, reassign then wrap once: "+
						"if IsNoResult(err) { err = ErrNoResult }; return errors.Wrap(err, ...)",
					ruleIDWrap)
			}
		}
		return true
	})
}

// checkDomainWrapBan — GID-176 in /domain/**: wrapping a same-module non-static
// error with Wrap is forbidden (its stack, if any, was already collected
// upstream) — WithMessage is used instead. Wrapping an external-call error
// (see classifyCallErrors) is required, not forbidden, so it is not reported here.
func checkDomainWrapBan(pass *analysis.Pass, fn *ast.FuncDecl, callErrs map[types.Object]errSource) {
	ast.Inspect(fn.Body, func(n ast.Node) bool {
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
		arg := call.Args[0]
		// A static error (model.ErrX, &BigError{}) — Wrap is required, it collects the stack first.
		if isStaticError(pass, arg) {
			return true
		}
		if !isErrorExpr(pass, arg) {
			return true
		}
		// An external-call error — Wrap is required here too (v2), not forbidden.
		if callErrs[objectOfErrExpr(pass, arg)] == errSourceExternal {
			return true
		}
		// A same-module error (a same-module call result, a parameter) — forbidden.
		pass.Reportf(call.Pos(),
			"%s: the stack is already collected upstream for a same-module error. "+
				"Fix: use errors.WithMessage instead of errors.Wrap for an incoming error",
			ruleIDWrap)
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

// ===== GID-237 =====

func runServiceMessage(pass *analysis.Pass, s Settings) (any, error) {
	if !inServiceMessageScope(pass.Pkg.Path()) {
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
			if exclude.Match(s.Exclude, recvTypeName(fn), fn.Name.Name) {
				continue
			}
			checkNoServiceMessage(pass, fn)
		}
	}
	return nil, nil
}

func checkNoServiceMessage(pass *analysis.Pass, fn *ast.FuncDecl) {
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		name := pkgErrorsCallName(pass, call)
		if name == "WithMessage" || name == "WithMessagef" {
			pass.Reportf(call.Pos(),
				"%s: errors.WithMessage is not used in a service. Fix: convert the error to a model error and wrap "+
					"with errors.WithStack; adding message context belongs to usecase",
				ruleIDServiceMessage)
		}
		return true
	})
}

func inServiceMessageScope(pkgPath string) bool {
	for _, scope := range serviceMessageScopes {
		if pathseg.HasLayer(pkgPath, scope...) {
			return true
		}
	}
	return false
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

// ===== shared helpers =====

func inBoundary(pkgPath string) bool {
	for _, scope := range boundaryScopes {
		if pathseg.HasLayer(pkgPath, scope...) {
			return true
		}
	}
	return false
}

// pkgErrorsCallName returns the name of the github.com/pkg/errors function
// if call invokes it; otherwise "".
func pkgErrorsCallName(pass *analysis.Pass, call *ast.CallExpr) string {
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

// classifyCallErrors collects the function's local variables whose value
// implements error and comes from a call that counts as a GID-176 boundary
// call, tagged with why:
//   - errSourceExternal — a direct call to a function/method whose declaring
//     package lies outside the current module (err := json.Unmarshal(...));
//     tracked in every layer;
//   - errSourceInterface — an interface-method call on an injected dependency
//     (err := c.conn.Select(...)); tracked only when boundary is true (the
//     current package is inside boundaryScopes).
//
// A call to a local same-module package function (a pure SQL builder, etc.)
// or a method on a concrete same-module type is neither and is not tracked.
func classifyCallErrors(pass *analysis.Pass, fn *ast.FuncDecl, boundary bool) map[types.Object]errSource {
	out := map[types.Object]errSource{}
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
		var src errSource
		switch {
		case isExternalCall(pass, call):
			src = errSourceExternal
		case boundary && isInterfaceMethodCall(pass, call):
			src = errSourceInterface
		default:
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
			out[obj] = src
		}
		return true
	})
	return out
}

// isExternalCall reports whether call invokes a function or method whose
// declaring package lies outside the current module — a direct call into a
// third-party library or the standard library (stdlib counts as external
// too). github.com/pkg/errors itself is excluded: calling New/Wrap/WithStack/
// WithMessage is the wrapping mechanism, not a dependency whose error needs
// wrapping.
func isExternalCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	fn := typeutil.Callee(pass.TypesInfo, call)
	f, ok := fn.(*types.Func)
	if !ok {
		return false
	}
	calleePkgObj := f.Pkg()
	if calleePkgObj == nil {
		return false
	}
	calleePkg := calleePkgObj.Path()
	if calleePkg == pkgErrorsPath {
		return false
	}
	return !sameModule(pass.Pkg.Path(), calleePkg)
}

// sameModule reports whether calleePkgPath (the package declaring the called
// function/method) belongs to the same module as pkgPath (the package under
// analysis). Mirrors the module-boundary convention used across the linter
// (analyzers/layerimports): for the canonical layout the /internal/ segment
// marks the module boundary; otherwise (testdata, a non-standard layout) the
// first path segment is compared. A package outside the module — including
// the standard library — is external.
func sameModule(pkgPath, calleePkgPath string) bool {
	const internalSeg = "/internal/"
	if module, _, ok := strings.Cut(pkgPath, internalSeg); ok {
		return calleePkgPath == module || strings.HasPrefix(calleePkgPath, module+internalSeg)
	}
	return firstSegment(pkgPath) == firstSegment(calleePkgPath)
}

func firstSegment(path string) string {
	seg, _, _ := strings.Cut(path, "/")
	return seg
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

func isTrackedCallErr(pass *analysis.Pass, expr ast.Expr, callErrs map[types.Object]errSource) bool {
	obj := objectOfErrExpr(pass, expr)
	if obj == nil {
		return false
	}
	_, ok := callErrs[obj]
	return ok
}

// objectOfErrExpr returns the object expr refers to when it is a plain
// identifier (a local error variable); nil otherwise.
func objectOfErrExpr(pass *analysis.Pass, expr ast.Expr) types.Object {
	id, ok := expr.(*ast.Ident)
	if !ok {
		return nil
	}
	return objectOf(pass, id)
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
