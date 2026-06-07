// Package bansymbol реализует правило GID-217 (линтер gidbansymbol):
// настраиваемый бан конкретных символов сторонних библиотек.
//
// Источник: repo.md — «Не используем gdpostgres.TQuery — прямые методы conn
// проще и достаточны». По умолчанию запрещён символ TQuery из библиотеки
// gitlab.gid.team/gid-data/tech/golang/libs/postgres.git; список можно
// переопределить через settings.symbols в .golangci.yml.
//
// Детекция: любой *ast.SelectorExpr, который через pass.TypesInfo.Uses
// резолвится в объект с заданным именем из заданного пакета. Generic-
// инстанциации (gdpostgres.TQuery[T](...)) резолвятся так же и тоже ловятся.
//
// Match пакета — по точному import-пути ИЛИ по суффиксу из сегментов пути
// (чтобы покрыть версионные пути вида .../v2). Сгенерированный код
// (ast.IsGenerated) пропускается.
package bansymbol

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-217"

// defaultSymbols — встроенный список: запрет gdpostgres.TQuery.
var defaultSymbols = []Symbol{
	{
		Pkg:  "gitlab.gid.team/gid-data/tech/golang/libs/postgres.git",
		Name: "TQuery",
		Msg:  "gdpostgres.TQuery is banned. Fix: use conn methods directly: Select, ScanRow, NamedStruct or Transaction (repo.md)",
	},
}

// Analyzer — вариант с настройками по умолчанию.
var Analyzer = NewAnalyzer(Settings{})

// Symbol — описание одного запрещённого символа.
type Symbol struct {
	// Pkg — import-путь пакета символа. Матчится точно ИЛИ по суффиксу
	// сегментов пути (например, ".../postgres.git" совпадёт с
	// ".../postgres.git/v2").
	Pkg string `json:"pkg"`
	// Name — имя экспортируемого символа (функция, тип, переменная).
	Name string `json:"name"`
	// Msg — текст подсказки в диагностике. Опционален: без него
	// используется общая формулировка.
	Msg string `json:"msg"`
}

// Settings — настройки линтера из .golangci.yml.
type Settings struct {
	// Symbols — список запрещённых символов. Если пуст — используется
	// встроенный дефолтный список.
	Symbols []Symbol `json:"symbols"`
}

// NewAnalyzer строит анализатор GID-217 с указанными настройками.
func NewAnalyzer(cfg Settings) *analysis.Analyzer {
	symbols := cfg.Symbols
	if len(symbols) == 0 {
		symbols = defaultSymbols
	}
	return &analysis.Analyzer{
		Name: "gidbansymbol",
		Doc:  ruleID + ": ban specific library symbols (configurable). Fix: replace the banned symbol with the project-approved alternative.",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, symbols)
		},
	}
}

func run(pass *analysis.Pass, symbols []Symbol) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			obj := pass.TypesInfo.Uses[sel.Sel]
			if obj == nil || obj.Pkg() == nil {
				return true
			}
			pkg := obj.Pkg()
			objPkg := pkg.Path()
			objName := obj.Name()
			//nolint:gidallptr // плагин не зависит от внутренней библиотеки gdhelper
			for _, s := range symbols {
				if s.Name != objName {
					continue
				}
				if !pkgMatches(objPkg, s.Pkg) {
					continue
				}
				report(pass, sel.Sel.Pos(), s, pkg.Name(), objName)
				break
			}
			return true
		})
	}
	return nil, nil
}

// pkgMatches сообщает, совпадает ли import-путь пакета символа с настройкой:
// точное равенство ИЛИ суффикс по сегментам пути.
func pkgMatches(objPkg, want string) bool {
	if objPkg == want {
		return true
	}
	return pathseg.EndsWith(objPkg, pathseg.Segments(want)...)
}

func report(pass *analysis.Pass, pos token.Pos, s Symbol, pkgName, name string) {
	if s.Msg != "" {
		pass.Reportf(pos, "%s: %s", ruleID, s.Msg)
		return
	}
	pass.Reportf(pos,
		"%s: symbol %s.%s is banned by gidbansymbol. "+
			"Fix: replace it with the project-approved alternative.",
		ruleID, pkgName, name)
}
