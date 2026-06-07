// Package loggernew реализует правило GID-214 (logger-singleton):
//
//   - GID-214 (gidloggernew): логгер создаётся один раз в composition root.
//     Вызовы logrus.New() и logrus.StandardLogger() (package
//     github.com/sirupsen/logrus) запрещены везде, кроме пакета main и
//     пакетов composition root (путь содержит сегменты internal/app).
//
// Готовый *logrus.Entry пробрасывается через конструктор, а не создаётся
// заново в service/repository — иначе теряется единая конфигурация логгера
// (формат, хуки, уровень) и сквозные поля.
//
// _test.go-файлы и сгенерированные файлы пропускаются: логгер в тестах —
// норма, генерируемый код не правится вручную.
//
// Резолв logrus идёт через types (import path), поэтому вызов New() из
// другого пакета с тем же именем не флагается.
//
// LoadMode: TypesInfo — нужен резолв пакета вызываемой функции по import-пути.
//
// Источник: libs.md (logrus: не создавать новые экземпляры, пробрасывать
// существующий).
package loggernew

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-214"

// bannedFuncs — package-level функции logrus, создающие/возвращающие
// глобальный экземпляр логгера.
var bannedFuncs = map[string]struct{}{
	"New":            {},
	"StandardLogger": {},
}

// Analyzer — правило GID-214: logrus.New()/StandardLogger() — только в composition root (main, internal/app).
var Analyzer = &analysis.Analyzer{
	Name: "gidloggernew",
	Doc:  ruleID + ": logrus.New()/StandardLogger() вызываются только в composition root (main, internal/app)",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	// composition root: пакет main или путь с сегментами internal/app —
	// создавать логгер здесь разрешено.
	if pass.Pkg.Name() == "main" || pathseg.Contains(pass.Pkg.Path(), "internal", "app") {
		return nil, nil
	}

	for _, file := range pass.Files {
		if ast.IsGenerated(file) || isTestFile(pass, file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if name, ok := bannedLogrusCall(pass, call); ok {
				pass.Reportf(call.Pos(),
					"%s: logrus.%s() вызывается только в composition root (main, internal/app) — "+
						"пробрасывай готовый *logrus.Entry через конструктор",
					ruleID, name)
			}
			return true
		})
	}
	return nil, nil
}

// bannedLogrusCall сообщает, является ли call вызовом package-level функции
// logrus.New()/logrus.StandardLogger(). Резолв — по типам: имя пакета берётся
// из import-пути объекта, а не из текста селектора.
func bannedLogrusCall(pass *analysis.Pass, call *ast.CallExpr) (string, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	if _, ok := bannedFuncs[sel.Sel.Name]; !ok {
		return "", false
	}
	fn, ok := pass.TypesInfo.ObjectOf(sel.Sel).(*types.Func)
	if !ok {
		return "", false
	}
	// package-level функция: получателя нет (метод WithField не флагается).
	sig, ok := fn.Type().(*types.Signature)
	if !ok || sig.Recv() != nil {
		return "", false
	}
	// logrusPkgPath — import-путь пакета logrus.
	const logrusPkgPath = "github.com/sirupsen/logrus"
	pkg := fn.Pkg()
	if pkg == nil || pkg.Path() != logrusPkgPath {
		return "", false
	}
	return sel.Sel.Name, true
}

func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	tokenFile := pass.Fset.File(file.Pos())
	name := tokenFile.Name()
	const suffix = "_test.go"
	return len(name) >= len(suffix) && name[len(name)-len(suffix):] == suffix
}
