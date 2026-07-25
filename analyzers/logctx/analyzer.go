// Package logctx implements rule GID-155: a log output is accompanied by
// context and an error. The requirement is the same on both stacks, only its
// shape differs (see internal/lgr):
//
//   - in a function with a context.Context parameter, a log call carries the
//     context — WithContext in the chain (logrus) or the *Context variant of
//     the terminal call, InfoContext(ctx, …) (slog, which has no WithContext);
//   - an Error*-level log carries the error — WithError (logrus) or the error
//     passed as an argument, slog.Any("error", err) (slog).
//
// "WithError if there is an error in scope" in the general case requires flow
// analysis — the deterministic part is tied to the Error level.
package logctx

import (
	"go/ast"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/lgr"
)

const ruleID = "GID-155"

// Analyzer — rule GID-155: a log call carries the context (when ctx is present) and, at Error level, the error. Fix: WithContext(ctx)/WithError(err) on logrus, InfoContext(ctx, …)/slog.Any("error", err) on slog.
var Analyzer = &analysis.Analyzer{
	Name: "gidlogctx",
	Doc: ruleID + ": a log call carries the context (when ctx is present) and, at Error level, the error. " +
		"Fix: WithContext(ctx)/WithError(err) on logrus, InfoContext(ctx, …)/slog.Any(\"error\", err) on slog",
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
			walkFunc(pass, fn.Type, fn.Body)
		}
	}
	return nil, nil
}

// walkFunc traverses a function body; nested function literals are checked
// with their own set of parameters (the presence of ctx is that of the nearest function).
func walkFunc(pass *analysis.Pass, fnType *ast.FuncType, body *ast.BlockStmt) {
	hasCtx := funcHasCtx(pass, fnType)
	ast.Inspect(body, func(n ast.Node) bool {
		switch nn := n.(type) {
		case *ast.FuncLit:
			walkFunc(pass, nn.Type, nn.Body)
			return false
		case *ast.CallExpr:
			checkCall(pass, nn, hasCtx)
		}
		return true
	})
}

func checkCall(pass *analysis.Pass, call *ast.CallExpr, hasCtx bool) {
	terminal, kind, ok := lgr.Terminal(pass, call)
	if !ok {
		return
	}
	sels, _ := lgr.Chain(pass, call)
	names := lgr.ChainNames(sels)
	// The diagnostic is on the terminal method name — for a multi-line chain
	// the call position points to its first line.
	pos := sels[0].Sel.Pos()
	if hasCtx && !carriesContext(terminal, names, kind) {
		pass.Reportf(pos, "%s: a log call in a function with ctx must carry the context. "+
			"Fix: add WithContext(ctx) (logrus) or call the *Context variant — InfoContext(ctx, …) (slog)", ruleID)
	}
	if strings.HasPrefix(terminal, "Error") && !carriesError(pass, call, names, kind) {
		pass.Reportf(pos, "%s: an Error-level log must carry the error. "+
			"Fix: add WithError(err) (logrus) or pass the error as an attribute — "+
			"ErrorContext(ctx, msg, slog.Any(\"error\", err)) (slog)", ruleID)
	}
}

// carriesContext reports whether the call passes the context to the logger:
// logrus does it with WithContext in the chain, slog with the *Context variant
// of the terminal method (InfoContext(ctx, …)), which takes ctx as its first
// argument.
func carriesContext(terminal string, names []string, kind lgr.Kind) bool {
	if slices.Contains(names, "WithContext") {
		return true
	}
	return kind == lgr.KindSlog && strings.HasSuffix(terminal, "Context")
}

// carriesError reports whether the Error-level call carries the error itself:
// logrus does it with WithError, slog with an argument of type error
// (slog.Any("error", err) resolves to slog.Attr, so the raw err passed into it
// is what we look for anywhere in the call arguments).
func carriesError(pass *analysis.Pass, call *ast.CallExpr, names []string, kind lgr.Kind) bool {
	if slices.Contains(names, "WithError") {
		return true
	}
	if kind != lgr.KindSlog {
		return false
	}
	found := false
	ast.Inspect(call, func(n ast.Node) bool {
		if found {
			return false
		}
		expr, ok := n.(ast.Expr)
		if !ok {
			return true
		}
		if tv, ok := pass.TypesInfo.Types[expr]; ok && isErrorType(tv.Type) {
			found = true
			return false
		}
		return true
	})
	return found
}

// isErrorType reports whether t implements the built-in error interface.
func isErrorType(t types.Type) bool {
	if t == nil {
		return false
	}
	errType := types.Universe.
		Lookup("error").
		Type()
	errIface, ok := errType.Underlying().(*types.Interface)
	if !ok {
		return false
	}
	return types.Implements(t, errIface)
}

func funcHasCtx(pass *analysis.Pass, fnType *ast.FuncType) bool {
	if fnType.Params == nil {
		return false
	}
	for _, field := range fnType.Params.List {
		t := pass.TypesInfo.TypeOf(field.Type)
		named, ok := t.(*types.Named)
		if !ok {
			continue
		}
		obj := named.Obj()
		pkg := obj.Pkg()
		if pkg != nil && pkg.Path() == "context" && obj.Name() == "Context" {
			return true
		}
	}
	return false
}
