// Package chainperline implements rule GID-196: a call chain of min-calls
// or more links (2 by default) is formatted with one call per line,
// including the first:
//
//	q := builder.
//		Select("id").
//		From("snapshots").
//		Where(cond)
//
// A chain link is a method/function call via a selector whose receiver is
// another call (including through intermediate fields: a.B().c.D()). Type
// conversions do not count as links. Logrus chains are the domain of GID-156
// (gidlogchain) and are skipped here to avoid duplicating the diagnostic.
// A single inline call is allowed; *_test.go and generated code are not checked.
package chainperline

import (
	"go/ast"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/lgr"
)

const ruleID = "GID-196"

// Analyzer — rule GID-196 with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings — settings of rule GID-196 from .golangci.yml.
type Settings struct {
	// MinCalls — the threshold: a chain of this many calls must be
	// multi-line. 0 → default (2).
	MinCalls int `json:"min-calls"`
}

// NewAnalyzer builds the GID-196 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	const defaultMinCalls = 2
	minCalls := s.MinCalls
	if minCalls < 2 {
		minCalls = defaultMinCalls
	}
	return &analysis.Analyzer{
		Name: "gidchainperline",
		Doc:  ruleID + ": a call chain must put each call on its own line, including the first. Fix: break the chain so every .Method() starts a new line.",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, minCalls)
		},
	}
}

func run(pass *analysis.Pass, minCalls int) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) || isTestFile(pass, file) {
			continue
		}
		// Links of already-processed chains do not form their own chains;
		// their arguments are still checked independently.
		visited := map[*ast.CallExpr]struct{}{}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if _, ok := visited[call]; ok {
				return true
			}
			sels, base := chain(pass, call, visited)
			if len(sels) < minCalls {
				return true
			}
			if isLogrusChain(pass, sels) {
				return true // the domain of GID-156
			}
			checkLines(pass, sels, base)
			return true
		})
	}
	return nil, nil
}

// chain collects the call chain from the outer call inward. It returns the
// link selectors (from the outermost to the innermost) and the base expression
// the chain starts on.
func chain(
	pass *analysis.Pass,
	call *ast.CallExpr,
	visited map[*ast.CallExpr]struct{},
) (sels []*ast.SelectorExpr, base ast.Expr) {
	cur := call
	for {
		sel, ok := cur.Fun.(*ast.SelectorExpr)
		if !ok || isConversion(pass, cur.Fun) {
			return sels, cur // a function call or a conversion — the base of the chain
		}
		visited[cur] = struct{}{}
		sels = append(sels, sel)
		inner := innermostCall(sel.X)
		if inner == nil {
			return sels, sel.X
		}
		cur = inner
	}
}

// innermostCall — the nearest call under the expression, through intermediate
// fields and parentheses (a.B().c → the call of B).
func innermostCall(e ast.Expr) *ast.CallExpr {
	for {
		switch v := e.(type) {
		case *ast.ParenExpr:
			e = v.X
		case *ast.SelectorExpr:
			e = v.X
		case *ast.CallExpr:
			return v
		default:
			return nil
		}
	}
}

func isConversion(pass *analysis.Pass, fun ast.Expr) bool {
	tv, ok := pass.TypesInfo.Types[fun]
	return ok && tv.IsType()
}

func isLogrusChain(pass *analysis.Pass, sels []*ast.SelectorExpr) bool {
	for _, sel := range sels {
		if lgr.IsMethodSel(pass, sel) {
			return true
		}
	}
	return false
}

// checkLines: the call lines strictly increase, starting from the line after
// the end of the base expression — each call on its own line.
func checkLines(pass *analysis.Pass, sels []*ast.SelectorExpr, base ast.Expr) {
	prevLine := pass.Fset.Position(base.End()).Line
	for i := len(sels) - 1; i >= 0; i-- {
		line := pass.Fset.Position(sels[i].Sel.Pos()).Line
		if line <= prevLine {
			pass.Reportf(sels[i].Sel.Pos(),
				"%s: a chain of %d calls must put one call per line, including the first. "+
					"Fix: break each .Method() onto its own line.",
				ruleID, len(sels))
			return
		}
		prevLine = line
	}
}

func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	return strings.HasSuffix(filepath.Base(pass.Fset.Position(file.Pos()).Filename), "_test.go")
}
