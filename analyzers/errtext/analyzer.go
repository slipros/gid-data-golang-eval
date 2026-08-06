// Package errtext implements rule GID-256: the TEXT of an error is not passed
// into an error constructor.
//
// Owner's principle: a cause travels inside the error chain, not as a string
// inside another error's message. `<err>.Error()` keeps the text and drops
// everything else — errors.Is/errors.As can no longer reach the cause, and the
// stack collected with it goes too:
//
//	wrapped := errors.Wrapf(err, "confirm yandex audience segment %d", segmentID)
//	return errors.WithMessage(ErrServerError, wrapped.Error())
//
// The stack Wrapf collected on the first line is discarded by the second, along
// with the chain down to *net.OpError / context.DeadlineExceeded. What ships is
// a package-level sentinel whose only stack points at the package init
// (GID-177). The consumer of that client repeated the shape from the other
// side: msg := err.Error() before replacing err with a domain sentinel, then
// errors.Wrap(err, msg) to carry the text across (incident 2026-08-06,
// ad-cabinet-connector internal/client/yandexaudience +
// internal/domain/service/yandex_channel.go).
//
// The shape appears when a stable CLASS has to go out and the CAUSE has to
// survive: a sentinel var cannot carry a cause, so the text gets smuggled in a
// string. The fix is a class that can carry one — a typed error:
//
//	type ServerError struct{ Err error }
//	func (e *ServerError) Unwrap() error { return e.Err }
//	return errors.Wrapf(&ServerError{Err: err}, "confirm segment %d", segmentID)
//
// Detect: a call to an error constructor — github.com/pkg/errors New / Errorf /
// Wrap / Wrapf / WithMessage / WithMessagef, or fmt.Errorf — one of whose
// MESSAGE arguments is the text of an error:
//
//	(a) <errExpr>.Error() written inline;
//	(b) an identifier whose EVERY assignment in the function is such a call
//	    (msg := err.Error()) — the flattening one line above the constructor.
//	    A variable reassigned from anything else is unknown at the call and is
//	    left alone: the rule does not guess.
//
// The first argument of Wrap/Wrapf/WithMessage/WithMessagef is the error being
// wrapped, not a message, and is never flagged.
//
// Out of scope: err.Error() anywhere that is not an error constructor — a log
// field, a gRPC status message (status.Error(codes.Internal, err.Error()) — a
// status message is a string by contract), a comparison, a struct field.
//
// Generated code and _test.go are skipped: a fixture legitimately rebuilds an
// expected error from a text — it only has to LOOK like the production one,
// and no one branches on its chain downstream (GID-250).
package errtext

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const (
	ruleID = "GID-256"

	pkgErrorsPath = "github.com/pkg/errors"
	fmtPath       = "fmt"
)

// messageStart — for each known constructor, the index of its first MESSAGE
// argument. Wrap/Wrapf/WithMessage/WithMessagef take the error being wrapped
// first; New/Errorf take the message right away.
var messageStart = map[string]map[string]int{
	pkgErrorsPath: {
		"New":          0,
		"Errorf":       0,
		"Wrap":         1,
		"Wrapf":        1,
		"WithMessage":  1,
		"WithMessagef": 1,
	},
	fmtPath: {
		"Errorf": 0,
	},
}

// Analyzer — rule GID-256 with no exclusions configured.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — functions allowed to flatten an error into a message:
	// "Function" (any receiver) or "Type.Method".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-256 analyzer from linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "giderrtext",
		Doc: ruleID + ": the text of an error (err.Error()) is passed into an error constructor — the chain " +
			"and the stack are dropped, only the text survives. " +
			"Fix: put the cause in the chain (errors.Wrap(err, \"context\")); for a stable class, " +
			"use a typed error carrying it (type ServerError struct { Err error } with Unwrap)",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s.Exclude)
		},
	}
}

func run(pass *analysis.Pass, excluded []string) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) || srcfile.IsTest(pass, file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if exclude.Match(excluded, receiverType(fn), fn.Name.Name) {
				continue
			}
			checkFunc(pass, fn)
		}
	}
	return nil, nil
}

// checkFunc reports every error constructor inside fn that is handed the text
// of an error, inline or through a variable holding it.
func checkFunc(pass *analysis.Pass, fn *ast.FuncDecl) {
	textVars := errorTextVars(pass, fn)
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		start, isCtor := messageArgsStart(pass, call)
		if !isCtor || len(call.Args) <= start {
			return true
		}
		for _, arg := range call.Args[start:] {
			if !isErrorText(pass, arg, textVars) {
				continue
			}
			pass.Reportf(arg.Pos(),
				"%s: the text of an error is passed into an error constructor — .Error() flattens the error "+
					"into a string, so errors.Is/errors.As can no longer reach the cause and the stack collected "+
					"with it is dropped. Fix: put the cause in the chain (errors.Wrap(err, \"context\")); when a "+
					"stable class is needed too, carry it in a typed error "+
					"(type ServerError struct { Err error } with Unwrap), not in a sentinel's message",
				ruleID)
		}
		return true
	})
}

// messageArgsStart reports whether call is a known error constructor and, if
// so, the index of its first message argument. Matching is done on the
// resolved callee, so an import alias makes no difference.
func messageArgsStart(pass *analysis.Pass, call *ast.CallExpr) (start int, ok bool) {
	fn, isFunc := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
	if !isFunc || fn.Pkg() == nil {
		return 0, false
	}
	pkg := fn.Pkg()
	ctors, known := messageStart[pkg.Path()]
	if !known {
		return 0, false
	}
	start, ok = ctors[fn.Name()]
	return start, ok
}

// isErrorText reports whether expr hands over the text of an error: the call
// <errExpr>.Error() itself, or an identifier from textVars holding its result.
func isErrorText(pass *analysis.Pass, expr ast.Expr, textVars map[types.Object]bool) bool {
	if isErrorMethodCall(pass, expr) {
		return true
	}
	id, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	return textVars[pass.TypesInfo.Uses[id]]
}

// isErrorMethodCall reports whether expr is <errExpr>.Error() — a no-argument
// call of the Error method on a value that implements error.
func isErrorMethodCall(pass *analysis.Pass, expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) != 0 {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Error" {
		return false
	}
	return isErrorType(pass.TypesInfo.TypeOf(sel.X))
}

// errorTextVars collects the local variables of fn whose EVERY assignment is
// an <errExpr>.Error() call — the shape that hides the flattening one line
// above the constructor (msg := err.Error(); … errors.Wrap(err, msg)). A
// variable that is also assigned anything else holds an unknown value at the
// constructor, so it is dropped from the set rather than guessed at.
func errorTextVars(pass *analysis.Pass, fn *ast.FuncDecl) map[types.Object]bool {
	text := map[types.Object]bool{}
	other := map[types.Object]bool{}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok || len(assign.Lhs) != len(assign.Rhs) {
			return true
		}
		for i, lhs := range assign.Lhs {
			obj := assignedObject(pass, lhs)
			if obj == nil {
				continue
			}
			if isErrorMethodCall(pass, assign.Rhs[i]) {
				text[obj] = true
				continue
			}
			other[obj] = true
		}
		return true
	})
	for obj := range other {
		delete(text, obj)
	}
	return text
}

// assignedObject returns the variable expr assigns to (both := and =), or nil
// when the target is not a plain identifier.
func assignedObject(pass *analysis.Pass, expr ast.Expr) types.Object {
	id, ok := expr.(*ast.Ident)
	if !ok {
		return nil
	}
	obj := pass.TypesInfo.Defs[id]
	if obj == nil {
		obj = pass.TypesInfo.Uses[id]
	}
	if _, isVar := obj.(*types.Var); !isVar {
		return nil
	}
	return obj
}

// receiverType returns the name of fn's receiver type ("Client" for both Client
// and *Client), or "" for a plain function — the form settings.exclude matches.
func receiverType(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}
	expr := fn.Recv.List[0].Type
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if id, ok := expr.(*ast.Ident); ok {
		return id.Name
	}
	return ""
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
