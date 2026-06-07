// Package upwardimport реализует правило GID-131 (no-upward-import):
// дочерний пакет не импортирует родительский.
//
// Направление зависимостей: общее выносится вниз, родитель импортирует
// детей, а не наоборот. Если путь импорта является строгим префиксом пути
// текущего пакета по сегментам (pkgPath начинается с impPath + "/"), значит
// дочерний пакет тянет родителя — инверсия зависимости.
//
// Самоимпорт в Go невозможен, соседние пакеты и внешние модули не матчатся
// по определению (их путь не является префиксом пути текущего пакета).
// Префикс считается по сегментам пути, а не по строке: "a/parentx" не
// является дочерним для "a/parent".
//
// Сгенерированный код (ast.IsGenerated) пропускается. LoadMode — TypesInfo
// (нужен pass.Pkg.Path()).
package upwardimport

import (
	"go/ast"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-131"

// Analyzer — правило GID-131: дочерний пакет не импортирует родительский.
var Analyzer = &analysis.Analyzer{
	Name: "gidupwardimport",
	Doc: ruleID + ": дочерний пакет не импортирует родительский — " +
		"инвертируйте зависимость: общее выносится вниз, родитель импортирует детей",
	Run: run,
}

func run(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, imp := range file.Imports {
			impPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			// Строгий префикс по сегментам: pkgPath начинается с impPath + "/".
			// Это означает, что impPath — родитель текущего пакета.
			if strings.HasPrefix(pkgPath, impPath+"/") {
				pass.Reportf(imp.Pos(),
					"%s: дочерний пакет импортирует родительский %s — "+
						"инвертируйте зависимость: общее выносится вниз, родитель импортирует детей",
					ruleID, impPath)
			}
		}
	}
	return nil, nil
}
