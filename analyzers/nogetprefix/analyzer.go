// Package nogetprefix реализует правило GID-101: методы для получения
// значений не имеют префикса Get (go-styleguide, «Именование методов»).
//
// Исключение — сгенерированный код (protobuf и т.п.), где Get-префикс
// является частью контракта.
package nogetprefix

import (
	"go/ast"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-101"

// Analyzer — правило GID-101: запрет префикса Get в именах методов.
var Analyzer = &analysis.Analyzer{
	Name: "gidnogetprefix",
	Doc:  ruleID + ": запрет префикса Get в именах методов",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil {
				continue
			}
			if hasGetPrefix(fn.Name.Name) {
				pass.Reportf(fn.Name.Pos(),
					"%s: метод %q использует префикс Get — геттеры именуются без него: %q",
					ruleID, fn.Name.Name, strings.TrimPrefix(fn.Name.Name, "Get"))
			}
		}
	}
	return nil, nil
}

// hasGetPrefix сообщает, начинается ли имя со слова Get: голое "Get" или
// "Get" + слово с заглавной буквы ("GetJob"). Имена вида "Getaway",
// где get — часть другого слова, не считаются нарушением.
func hasGetPrefix(name string) bool {
	if name == "Get" {
		return true
	}
	rest, ok := strings.CutPrefix(name, "Get")
	if !ok {
		return false
	}
	r, _ := utf8.DecodeRuneInString(rest)
	return unicode.IsUpper(r) || unicode.IsDigit(r)
}
