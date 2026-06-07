// Package chainperline реализует правило GID-196: цепочка вызовов из
// min-calls и более звеньев (по умолчанию 2) оформляется по одному вызову
// на строке, включая первый:
//
//	q := builder.
//		Select("id").
//		From("snapshots").
//		Where(cond)
//
// Звено цепочки — вызов метода/функции через селектор, чей получатель —
// другой вызов (в т.ч. через промежуточные поля: a.B().c.D()). Конверсии
// типов звеном не считаются. Logrus-цепочки — зона GID-156 (gidlogchain),
// здесь пропускаются, чтобы не дублировать диагностику. Одиночный вызов
// inline допустим; *_test.go и сгенерированный код не проверяются.
package chainperline

import (
	"go/ast"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/lgr"
)

const ruleID = "GID-196"

// Analyzer — правило GID-196 с настройками по умолчанию.
var Analyzer = NewAnalyzer(Settings{})

// Settings — настройки правила GID-196 из .golangci.yml.
type Settings struct {
	// MinCalls — порог: цепочка из стольких вызовов обязана быть
	// многострочной. 0 → дефолт (2).
	MinCalls int `json:"min-calls"`
}

// NewAnalyzer строит анализатор GID-196 из настроек линтера (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	const defaultMinCalls = 2
	minCalls := s.MinCalls
	if minCalls < 2 {
		minCalls = defaultMinCalls
	}
	return &analysis.Analyzer{
		Name: "gidchainperline",
		Doc:  ruleID + ": цепочка вызовов — каждый вызов на своей строке, включая первый",
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
		// Звенья уже разобранных цепочек не образуют свои цепочки;
		// их аргументы при этом проверяются независимо.
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
				return true // зона GID-156
			}
			checkLines(pass, sels, base)
			return true
		})
	}
	return nil, nil
}

// chain собирает цепочку вызовов от внешнего call вглубь. Возвращает
// селекторы звеньев (от внешнего к внутреннему) и базовое выражение,
// на котором начата цепочка.
func chain(
	pass *analysis.Pass,
	call *ast.CallExpr,
	visited map[*ast.CallExpr]struct{},
) (sels []*ast.SelectorExpr, base ast.Expr) {
	cur := call
	for {
		sel, ok := cur.Fun.(*ast.SelectorExpr)
		if !ok || isConversion(pass, cur.Fun) {
			return sels, cur // вызов функции либо конверсия — база цепочки
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

// innermostCall — ближайший вызов под выражением, сквозь промежуточные
// поля и скобки (a.B().c → вызов B).
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

// checkLines: строки вызовов строго возрастают, начиная со строки после
// конца базового выражения, — каждый вызов на своей строке.
func checkLines(pass *analysis.Pass, sels []*ast.SelectorExpr, base ast.Expr) {
	prevLine := pass.Fset.Position(base.End()).Line
	for i := len(sels) - 1; i >= 0; i-- {
		line := pass.Fset.Position(sels[i].Sel.Pos()).Line
		if line <= prevLine {
			pass.Reportf(sels[i].Sel.Pos(),
				"%s: цепочка из %d вызовов оформляется по одному вызову на строке, включая первый",
				ruleID, len(sels))
			return
		}
		prevLine = line
	}
}

func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	return strings.HasSuffix(filepath.Base(pass.Fset.Position(file.Pos()).Filename), "_test.go")
}
