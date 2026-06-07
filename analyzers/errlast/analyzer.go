// Package errlast реализует правило GID-190: соглашения Google по ошибкам
// в результатах функций и методов.
//
// Проверка 1 (error не последний). Если среди результатов есть тип error
// и после него идут другие результаты — это нарушение: error обязан быть
// последним возвращаемым значением.
//
// Проверка 2 (конкретный error-тип в результате). Результат функции —
// конкретный тип, реализующий error (именованный тип или указатель на него,
// например *MyError / MyError), а не интерфейс error. Конкретный тип в
// interface-позиции даёт классическую typed-nil ловушку: возвращённый
// nil-указатель в переменной типа error != nil. Возвращать следует
// интерфейс error.
//
// НЕ матчатся проверкой 2:
//   - сам интерфейс error;
//   - интерфейсные типы, расширяющие error (кастомный error-интерфейс —
//     осознанное решение автора);
//   - функции-конструкторы ошибок в файлах error.go / errors.go — там
//     конкретный тип легитимен (это конструкторы вида NewMyError() *MyError).
//
// Сгенерированный код (ast.IsGenerated) пропускается. LoadMode — TypesInfo.
package errlast

import (
	"go/ast"
	"go/types"
	"path/filepath"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-190"

// errorFiles — файлы, в которых приватные функции-конструкторы ошибок
// могут возвращать конкретный error-тип (проверка 2 не применяется).
var errorFiles = map[string]bool{
	"error.go":  true,
	"errors.go": true,
}

// Analyzer — правило GID-190: error — последний результат, конкретные error-типы не возвращаются.
var Analyzer = &analysis.Analyzer{
	Name: "giderrlast",
	Doc:  ruleID + ": error must be the last result, and the error interface (not a concrete type) is returned. Fix: move error last and return the error interface",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	errObj := types.Universe.Lookup("error")
	errType := errObj.Type()
	errIface, ok := errType.Underlying().(*types.Interface)
	if !ok {
		return nil, nil
	}

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		tokenFile := pass.Fset.File(file.Pos())
		inErrorFile := errorFiles[filepath.Base(tokenFile.Name())]

		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}
			checkResults(pass, fn, errIface, inErrorFile)
			return true
		})
	}
	return nil, nil
}

// checkResults применяет обе проверки к результатам функции/метода.
func checkResults(pass *analysis.Pass, fn *ast.FuncDecl, errIface *types.Interface, inErrorFile bool) {
	if fn.Type.Results == nil {
		return
	}

	// Развернём group-результаты в плоский список (тип, выражение).
	type result struct {
		expr ast.Expr
		typ  types.Type
	}
	var results []*result
	for _, field := range fn.Type.Results.List {
		t := pass.TypesInfo.TypeOf(field.Type)
		if t == nil {
			continue
		}
		// Несколько имён результатов одного типа: (a, b int) — каждый — отдельный результат.
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for range count {
			results = append(results, &result{expr: field.Type, typ: t})
		}
	}
	if len(results) == 0 {
		return
	}

	for i, r := range results {
		// Проверка 1: error не последний — есть результаты после него.
		if isExactError(r.typ) && i != len(results)-1 {
			pass.Reportf(r.expr.Pos(),
				"%s: error must be the last return value. Fix: move it to the end", ruleID)
			continue
		}

		// Проверка 2: конкретный error-тип в результате.
		if inErrorFile {
			continue // конструкторы ошибок в error.go/errors.go легитимно возвращают конкретный тип
		}
		if isConcreteError(r.typ, errIface) {
			pass.Reportf(r.expr.Pos(),
				"%s: return the error interface, not %s. Fix: a concrete type in the error position causes a typed-nil trap",
				ruleID, r.typ.String())
		}
	}
}

// isExactError сообщает, является ли тип ровно интерфейсом error.
func isExactError(t types.Type) bool {
	named, ok := t.(*types.Named)
	if ok {
		obj := named.Obj()
		return obj != nil && obj.Pkg() == nil && obj.Name() == "error"
	}
	return false
}

// isConcreteError сообщает, реализует ли тип error, будучи конкретным
// (неинтерфейсным) типом — именованным или указателем на именованный.
// Интерфейсы (включая сам error и кастомные error-интерфейсы) исключаются.
func isConcreteError(t types.Type, errIface *types.Interface) bool {
	// Интерфейсный тип (error и его расширения) — не конкретный.
	if _, isIface := t.Underlying().(*types.Interface); isIface {
		return false
	}

	// Интересуют только именованные типы и указатели на именованные.
	switch u := t.(type) {
	case *types.Named:
		// ok
	case *types.Pointer:
		if _, ok := u.Elem().(*types.Named); !ok {
			return false
		}
	default:
		return false
	}

	return types.Implements(t, errIface) || types.Implements(types.NewPointer(t), errIface)
}
