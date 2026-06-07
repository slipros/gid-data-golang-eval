// Package initclean реализует правило GID-180: init() детерминированный.
// Внутри func init() запрещены:
//   - запуск горутин (go-statement) — прямо в теле init, включая вложенные
//     блоки и замыкания, объявленные в самом init;
//   - вызовы функций I/O-пакетов (os, net, net/http, database/sql,
//     io/ioutil, bufio и т.п.) — фоновую работу и I/O выполняйте из
//     main/конструктора/app.
//
// Чтение переменных окружения (os.Getenv, os.LookupEnv) разрешено — это не I/O.
//
// Список I/O-пакетов настраивается через settings.packages (заменяет дефолтный).
package initclean

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-180"

// defaultPackages — пакеты, вызовы функций которых считаются I/O.
// Матчинг по пути пакета через TypesInfo.
var defaultPackages = []string{
	"os",
	"net",
	"net/http",
	"database/sql",
	"io/ioutil",
	"bufio",
}

// allowedFuncs — функции I/O-пакетов, разрешённые в init().
// Чтение env детерминировано и I/O-эффекта не несёт.
var allowedFuncs = map[string]map[string]bool{
	"os": {
		"Getenv":    true,
		"LookupEnv": true,
	},
}

// Analyzer — вариант с дефолтным списком I/O-пакетов.
var Analyzer = NewAnalyzer(Settings{})

// Settings — настройки линтера из .golangci.yml.
type Settings struct {
	// Packages — I/O-пакеты (пути import). Заменяет дефолтный список.
	Packages []string `json:"packages"`
}

// NewAnalyzer строит анализатор GID-180 из настроек линтера (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	pkgs := s.Packages
	if len(pkgs) == 0 {
		pkgs = defaultPackages
	}
	ioPkgs := make(map[string]bool, len(pkgs))
	for _, p := range pkgs {
		ioPkgs[p] = true
	}
	return &analysis.Analyzer{
		Name: "gidinitclean",
		Doc:  ruleID + ": init() детерминированный — без goroutine и I/O-вызовов",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, ioPkgs)
		},
	}
}

func run(pass *analysis.Pass, ioPkgs map[string]bool) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name.Name != "init" || fn.Body == nil {
				continue
			}
			checkInitBody(pass, fn.Body, ioPkgs)
		}
	}
	return nil, nil
}

// checkInitBody обходит тело init() целиком (включая вложенные блоки и
// тела замыканий, объявленных прямо в init) и репортит запрещённые
// конструкции.
func checkInitBody(pass *analysis.Pass, body *ast.BlockStmt, ioPkgs map[string]bool) {
	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.GoStmt:
			pass.Reportf(node.Pos(),
				"%s: goroutine в init() запрещена — init детерминированный, "+
					"фоновую работу запускайте из main/app", ruleID)
		case *ast.CallExpr:
			pkg, fn := selectorPkgFunc(pass, node)
			if pkg == "" || !ioPkgs[pkg] {
				return true
			}
			if allowedFuncs[pkg][fn] {
				return true
			}
			pass.Reportf(node.Pos(),
				"%s: I/O-вызов %s.%s в init() запрещён — "+
					"выполняйте в main/конструкторе", ruleID, pkg, fn)
		}
		return true
	})
}

// selectorPkgFunc возвращает путь пакета и имя функции для вызова вида
// pkg.Func(...). Для не-пакетных вызовов (методы, локальные функции)
// возвращает пустую строку пакета.
func selectorPkgFunc(pass *analysis.Pass, call *ast.CallExpr) (pkgPath, fnName string) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", ""
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return "", ""
	}
	pkgName, ok := pass.TypesInfo.Uses[ident].(*types.PkgName)
	if !ok {
		return "", ""
	}
	imported := pkgName.Imported()
	return imported.Path(), sel.Sel.Name
}
