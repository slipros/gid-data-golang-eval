// Package constvarorder implements rule GID-130: the declaration order
// in a file is import, then const blocks, then var blocks, then types and functions.
// All consts are always at the top of the file below import; vars go below const (if const exists).
package constvarorder

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-130"

// Rank of a declaration class: the order in a file must not decrease.
const (
	rankImport = iota
	rankConst
	rankVar
	rankOther
)

// Analyzer — rule GID-130: declaration order in a file must be import, const, var, then types and functions. Fix: move const/var blocks to the top.
var Analyzer = &analysis.Analyzer{
	Name: "gidconstvarorder",
	Doc:  ruleID + ": declaration order in a file must be import, const, var, then types and functions. Fix: move const/var blocks to the top",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		maxRank := rankImport
		for _, decl := range file.Decls {
			r := rank(decl)
			if r >= maxRank {
				maxRank = r
				continue
			}
			switch r {
			case rankConst:
				pass.Reportf(decl.Pos(),
					"%s: a const block must be at the top of the file, right after import and above var, types and functions. Fix: move it up", ruleID)
			case rankVar:
				pass.Reportf(decl.Pos(),
					"%s: a var block must be at the top of the file, after const and above types and functions. Fix: move it up", ruleID)
			}
		}
	}
	return nil, nil
}

func rank(decl ast.Decl) int {
	gd, ok := decl.(*ast.GenDecl)
	if !ok {
		return rankOther // FuncDecl
	}
	switch gd.Tok {
	case token.IMPORT:
		return rankImport
	case token.CONST:
		return rankConst
	case token.VAR:
		return rankVar
	default:
		return rankOther // type
	}
}
