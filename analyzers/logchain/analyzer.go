// Package logchain реализует правило GID-156: цепочка logrus-вызовов
// не пишется inline — каждый вызов на своей строке, включая первый:
//
//	c.logger.
//		WithContext(ctx).
//		WithError(err).
//		WithField("some", field).
//		Info("some text")
//
// Одиночный вызов (logger.Info("x")) под правило не попадает.
package logchain

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/lgr"
)

const ruleID = "GID-156"

// Analyzer — правило GID-156: цепочка logrus из ≥2 вызовов — по вызову на строке.
var Analyzer = &analysis.Analyzer{
	Name: "gidlogchain",
	Doc:  ruleID + ": a logrus chain puts each call on its own line, including the first. Fix: break each call onto a new line",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if _, ok := lgr.IsTerminal(pass, call); !ok {
				return true
			}
			checkChain(pass, call)
			return true
		})
	}
	return nil, nil
}

func checkChain(pass *analysis.Pass, call *ast.CallExpr) {
	sels, base := lgr.Chain(pass, call)
	if len(sels) < 2 {
		return // одиночный вызов — inline допустим
	}
	// sels идут от терминала вглубь; проверяем в порядке исходника.
	prevLine := pass.Fset.Position(base.End()).Line
	for i := len(sels) - 1; i >= 0; i-- {
		line := pass.Fset.Position(sels[i].Sel.Pos()).Line
		if line <= prevLine {
			pass.Reportf(sels[i].Sel.Pos(),
				"%s: a logrus chain must put one call per line, including the first. Fix: break each call onto a new line", ruleID)
			return
		}
		prevLine = line
	}
}
