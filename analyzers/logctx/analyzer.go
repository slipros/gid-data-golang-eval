// Package logctx implements rule GID-155: a log output is accompanied by
// context and an error.
//
//   - in a function with a context.Context parameter, a log call must contain
//     WithContext in the chain;
//   - an Error*-level log must contain WithError.
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

// Analyzer — rule GID-155: log calls include WithContext (when ctx is present) and WithError (at Error level). Fix: add WithContext(ctx)/WithError(err).
var Analyzer = &analysis.Analyzer{
	Name: "gidlogctx",
	Doc:  ruleID + ": log calls include WithContext (when ctx is present) and WithError (at Error level). Fix: add WithContext(ctx)/WithError(err)",
	Run:  run,
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
	terminal, ok := lgr.IsTerminal(pass, call)
	if !ok {
		return
	}
	sels, _ := lgr.Chain(pass, call)
	names := lgr.ChainNames(sels)
	// The diagnostic is on the terminal method name — for a multi-line chain
	// the call position points to its first line.
	pos := sels[0].Sel.Pos()
	if hasCtx && !slices.Contains(names, "WithContext") {
		pass.Reportf(pos,
			"%s: a log call in a function with ctx must include WithContext(ctx). Fix: add WithContext(ctx)", ruleID)
	}
	if strings.HasPrefix(terminal, "Error") && !slices.Contains(names, "WithError") {
		pass.Reportf(pos,
			"%s: an Error-level log must include WithError(err). Fix: add WithError(err)", ruleID)
	}
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
