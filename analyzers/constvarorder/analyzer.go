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

// Analyzer — правило GID-130: порядок объявлений в файле — import, const, var, типы и функции.
var Analyzer = &analysis.Analyzer{
	Name: "gidconstvarorder",
	Doc:  ruleID + ": порядок объявлений в файле — import, const, var, типы и функции",
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
					"%s: const-блок размещается сверху файла — сразу после import, выше var, типов и функций", ruleID)
			case rankVar:
				pass.Reportf(decl.Pos(),
					"%s: var-блок размещается сверху файла — после const, выше типов и функций", ruleID)
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
