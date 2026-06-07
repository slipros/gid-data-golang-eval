// Package logctx реализует правило GID-155: вывод в лог сопровождается
// контекстом и ошибкой.
//
//   - в функции с параметром context.Context лог-вызов обязан содержать
//     WithContext в цепочке;
//   - лог уровня Error* обязан содержать WithError.
//
// «WithError если есть error в области видимости» в общем виде требует
// анализа потока — детерминированная часть привязана к уровню Error.
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

// Analyzer — правило GID-155: лог-вызовы сопровождаются WithContext (при наличии ctx) и WithError (на уровне Error).
var Analyzer = &analysis.Analyzer{
	Name: "gidlogctx",
	Doc:  ruleID + ": лог-вызовы сопровождаются WithContext (при наличии ctx) и WithError (на уровне Error)",
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

// walkFunc обходит тело функции; вложенные функции-литералы проверяются
// со своим набором параметров (наличие ctx — у ближайшей функции).
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
	// Диагностика на имени терминального метода — у многострочной цепочки
	// позиция call указывает на её первую строку.
	pos := sels[0].Sel.Pos()
	if hasCtx && !slices.Contains(names, "WithContext") {
		pass.Reportf(pos,
			"%s: лог-вызов в функции с ctx обязан содержать WithContext(ctx)", ruleID)
	}
	if strings.HasPrefix(terminal, "Error") && !slices.Contains(names, "WithError") {
		pass.Reportf(pos,
			"%s: лог уровня Error обязан содержать WithError(err)", ruleID)
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
