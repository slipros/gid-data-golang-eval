// Package constvarorder реализует правило GID-130: порядок объявлений
// в файле — import, затем const-блоки, затем var-блоки, затем типы и функции.
// Все const всегда сверху файла под import; var — под const (если const есть).
package constvarorder

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-130"

// Ранг класса декларации: порядок в файле не должен убывать.
const (
	rankImport = iota
	rankConst
	rankVar
	rankOther
)

// Analyzer — правило GID-130: declaration order in a file must be import, const, var, then types and functions. Fix: move const/var blocks to the top.
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
