// Package flagmain реализует правило GID-192 (flags):
//
//   - Регистрация флагов через пакет stdlib "flag" (функции flag.String/Int/
//     Bool/.../flag.Parse/flag.Var и методы flag.FlagSet) разрешена только
//     в пакете main: флаги объявляет бинарь, библиотека принимает параметры.
//   - В пакете main имя флага (первый строковый константный аргумент
//     flag.String/Int/Bool/Duration/Float64/Var и т.п.) должно быть в
//     snake_case: заглавные буквы и дефисы запрещены, цифры и `_` допустимы.
//
// Camel-case имя ПЕРЕМЕННОЙ, в которую кладётся флаг, здесь не проверяется —
// это зона revive/ST1003.
//
// Детект пакета flag — через TypesInfo (путь пакета "flag"), поэтому свой
// локальный пакет с именем "flag" под правило не подпадает.
//
// Тестовые файлы (*_test.go) и пакеты с суффиксом _test пропускаются: flag
// в тестах бывает легитимен. Сгенерированный код (ast.IsGenerated) тоже
// пропускается. LoadMode — TypesInfo.
package flagmain

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-192"

// flagPkgPath — путь пакета регистрации флагов в stdlib.
const flagPkgPath = "flag"

// Analyzer — правило GID-192: flag.* только в пакете main; имена флагов snake_case.
var Analyzer = &analysis.Analyzer{
	Name: "gidflagmain",
	Doc:  ruleID + ": регистрация флагов только в пакете main, имя флага в snake_case",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	// Пакет с суффиксом _test (внешний тестовый пакет) — пропускаем целиком.
	if strings.HasSuffix(pass.Pkg.Name(), "_test") {
		return nil, nil
	}
	isMain := pass.Pkg.Name() == "main"

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		// Файлы *_test.go пропускаем — flag в тестах легитимен.
		if isTestFile(pass, file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			fn := flagFunc(pass, call)
			if fn == nil {
				return true
			}

			if !isMain {
				pass.Reportf(call.Pos(),
					"%s: регистрация флага вне пакета main запрещена — "+
						"флаги объявляет бинарь, библиотека принимает параметры", ruleID)
				return true
			}

			// В пакете main проверяем имя флага на snake_case.
			pos, name, ok := flagName(pass, fn, call)
			if !ok {
				return true
			}
			if !isSnakeCase(name) {
				pass.Reportf(pos, "%s: имя флага %q — используйте snake_case", ruleID, name)
			}
			return true
		})
	}
	return nil, nil
}

// isTestFile сообщает, что файл — *_test.go.
func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	name := pass.Fset.Position(file.Pos()).Filename
	return strings.HasSuffix(name, "_test.go")
}

// flagFunc возвращает *types.Func, если вызов — это функция пакета flag или
// метод типа из пакета flag (например *flag.FlagSet). Иначе nil.
func flagFunc(pass *analysis.Pass, call *ast.CallExpr) *types.Func {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}
	fn, ok := pass.TypesInfo.ObjectOf(sel.Sel).(*types.Func)
	if !ok {
		return nil
	}
	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		return nil
	}

	// Метод типа из пакета flag (например *flag.FlagSet).
	if recv := sig.Recv(); recv != nil {
		if isFlagType(recv.Type()) {
			return fn
		}
		return nil
	}

	// Пакетная функция flag.*.
	if pkg := fn.Pkg(); pkg != nil && pkg.Path() == flagPkgPath {
		return fn
	}
	return nil
}

// isFlagType сообщает, относится ли тип к пакету flag (например flag.FlagSet).
func isFlagType(t types.Type) bool {
	switch tt := t.(type) {
	case *types.Pointer:
		return isFlagType(tt.Elem())
	case *types.Alias:
		return isFlagType(types.Unalias(tt))
	case *types.Named:
		obj := tt.Obj()
		pkg := obj.Pkg()
		return pkg != nil && pkg.Path() == flagPkgPath
	}
	return false
}

// flagName извлекает имя флага из вызова регистрирующей функции flag.
// Возвращает позицию аргумента-имени, само имя и ok=true, только если имя
// задано строковой константой. Для функций без имени флага (Parse, Parsed,
// Args, NArg, …) и для динамических (не-константных) имён ok=false.
func flagName(pass *analysis.Pass, fn *types.Func, call *ast.CallExpr) (token.Pos, string, bool) {
	idx, ok := nameArgIndex(fn.Name())
	if !ok || idx >= len(call.Args) {
		return token.NoPos, "", false
	}
	arg := call.Args[idx]
	tv, ok := pass.TypesInfo.Types[arg]
	if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
		return token.NoPos, "", false
	}
	return arg.Pos(), constant.StringVal(tv.Value), true
}

// nameArgIndex возвращает индекс аргумента с именем флага для функции/метода
// регистрации флага. Группы:
//   - String/Int/Int64/Uint/Uint64/Float64/Bool/Duration/Func/BoolFunc
//     (name первым) — индекс 0;
//   - Var/StringVar/IntVar/Int64Var/UintVar/Uint64Var/BoolVar/Float64Var/
//     DurationVar/TextVar (первым идёт указатель или Value, имя — вторым) —
//     индекс 1.
func nameArgIndex(name string) (int, bool) {
	switch name {
	case "String", "Int", "Int64", "Uint", "Uint64", "Float64",
		"Bool", "Duration", "Func", "BoolFunc":
		return 0, true
	case "Var", "StringVar", "IntVar", "Int64Var", "UintVar", "Uint64Var",
		"BoolVar", "Float64Var", "DurationVar", "TextVar":
		return 1, true
	}
	return 0, false
}

// isSnakeCase сообщает, что имя флага в snake_case: только строчные буквы,
// цифры и `_`. Заглавные буквы и дефисы запрещены. Пустое имя считаем
// корректным (не наша забота).
func isSnakeCase(name string) bool {
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '_':
		default:
			return false
		}
	}
	return true
}
