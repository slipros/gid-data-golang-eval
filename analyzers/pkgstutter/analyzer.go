// Package pkgstutter реализует правило GID-193 (no-pkg-stutter): экспортируемый
// символ верхнего уровня (тип, функция, var, const) не должен начинаться с имени
// пакета. Снаружи такой символ читается с заиканием: widget.WidgetOptions,
// widget.WidgetService — имя пакета уже даёт контекст, префикс лишний.
// Достаточно widget.Options, widget.Service.
//
// Сравнение по границе CamelCase-слова: имя пакета должно совпасть с первым
// словом символа целиком. Пакет widget матчит WidgetOptions/WidgetCount, но
// пакет log НЕ матчит Logger (Logger начинается со слова "Logger", а не "Log"),
// а пакет conv НЕ матчит Convert (слово "Convert", а не "Conv").
//
// Исключения (не матчим):
//   - конструкторы New* — наш GID-104 требует New<Entity>, конфликт решён в его пользу;
//   - методы (есть ресивер) и неэкспортируемые символы;
//   - пакет main.
//
// Сгенерированные файлы (ast.IsGenerated) пропускаются.
// LoadMode — Syntax: типы не нужны, хватает имени пакета и AST.
package pkgstutter

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-193"

// Analyzer — правило GID-193: экспортируемый символ не повторяет имя пакета (widget.WidgetOptions).
var Analyzer = &analysis.Analyzer{
	Name: "gidpkgstutter",
	Doc:  ruleID + ": экспортируемый символ не повторяет имя пакета — снаружи это заикание (widget.WidgetOptions)",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	pkgName := pass.Pkg.Name()
	if pkgName == "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				checkFunc(pass, pkgName, d)
			case *ast.GenDecl:
				checkGenDecl(pass, pkgName, d)
			}
		}
	}
	return nil, nil
}

func checkFunc(pass *analysis.Pass, pkgName string, fn *ast.FuncDecl) {
	if fn.Recv != nil {
		return // метод — у него есть ресивер, имя читается как value.Method
	}
	name := fn.Name.Name
	if strings.HasPrefix(name, "New") {
		return // конструктор New* — GID-104 главнее
	}
	report(pass, pkgName, name, fn.Name.Pos())
}

func checkGenDecl(pass *analysis.Pass, pkgName string, gd *ast.GenDecl) {
	for _, spec := range gd.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			report(pass, pkgName, s.Name.Name, s.Name.Pos())
		case *ast.ValueSpec:
			for _, ident := range s.Names {
				report(pass, pkgName, ident.Name, ident.Pos())
			}
		}
	}
}

// report выводит диагностику, если name экспортируемый и его первое
// CamelCase-слово совпадает с именем пакета (регистронезависимо).
func report(pass *analysis.Pass, pkgName, name string, pos token.Pos) {
	if !ast.IsExported(name) {
		return
	}
	if !stutters(pkgName, name) {
		return
	}
	suffix := name[len(pkgName):]
	pass.Reportf(pos,
		"%s: %s повторяет имя пакета %s — снаружи это %s.%s; уберите префикс",
		ruleID, name, pkgName, pkgName, suffix)
}

// stutters сообщает, начинается ли символ с имени пакета как с отдельного
// CamelCase-слова. Сравнение регистронезависимо, но граница слова учитывается:
// после префикса длиной len(pkgName) должна начинаться новая заглавная буква
// (следующее слово), иначе имя пакета — лишь часть другого слова (log → Logger).
func stutters(pkgName, name string) bool {
	if len(name) <= len(pkgName) {
		return false // точное совпадение или короче — нет следующего слова
	}
	if !strings.EqualFold(name[:len(pkgName)], pkgName) {
		return false
	}
	// Следующий рунный символ должен быть началом нового CamelCase-слова —
	// заглавной буквой. Если строчная — имя пакета лишь префикс слова.
	next := rune(name[len(pkgName)])
	return unicode.IsUpper(next)
}
