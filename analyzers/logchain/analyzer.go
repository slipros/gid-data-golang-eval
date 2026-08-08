// Package logchain implements rule GID-156: a chain of logger calls (logrus
// or slog — see internal/lgr)
// is not written inline — each call on its own line, including the first:
//
//	c.logger.
//		WithContext(ctx).
//		WithError(err).
//		WithField("some", field).
//		Info("some text")
//
// A single call (logger.Info("x")) does not fall under the rule.
package logchain

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/lgr"
)

const ruleID = "GID-156"

// Analyzer — rule GID-156: a logger chain of >=2 calls — one call per line.
var Analyzer = &analysis.Analyzer{
	Name:     "gidlogchain",
	Doc:      ruleID + ": a logger chain puts each call on its own line, including the first. Fix: break each call onto a new line",
	Requires: astwalk.Requires,
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	astwalk.NodesOf(pass, ast.IsGenerated, func(_ *ast.File, call *ast.CallExpr) {
		if _, ok := lgr.IsTerminal(pass, call); !ok {
			return
		}
		checkChain(pass, call)
	})

	return nil, nil
}

func checkChain(pass *analysis.Pass, call *ast.CallExpr) {
	sels, base := lgr.Chain(pass, call)
	if len(sels) < 2 {
		return // a single call — inline is allowed
	}
	// sels go from the terminal inward; we check in source order.
	prevLine := pass.Fset.Position(base.End()).Line
	for i := len(sels) - 1; i >= 0; i-- {
		line := pass.Fset.Position(sels[i].Sel.Pos()).Line
		if line <= prevLine {
			pass.Reportf(sels[i].Sel.Pos(),
				"%s: a logger chain must put one call per line, including the first. Fix: break each call onto a new line", ruleID)
			return
		}
		prevLine = line
	}
}
