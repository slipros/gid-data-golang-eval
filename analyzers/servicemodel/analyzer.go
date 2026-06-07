// Package servicemodel реализует правило GID-151: API domain-сервиса
// работает только с model. Сервис принимает model, внутри конвертирует
// её в entity для репозитория, полученную entity конвертирует обратно
// и всегда возвращает model.
//
// Проверка: у экспортируемых методов в корне /domain/service параметры
// и возвращаемые значения не ссылаются на типы из /dal/entity (рекурсивно —
// через указатели, слайсы, мапы и поля). Внутри тела метода entity
// допустима — этим занимается конвертация.
package servicemodel

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-151"

// Analyzer — правило GID-151: экспортируемые методы сервиса принимают и возвращают model, не entity.
var Analyzer = &analysis.Analyzer{
	Name: "gidservicemodel",
	Doc:  ruleID + ": экспортируемые методы сервиса принимают и возвращают model, не entity",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if !pathseg.EndsWith(pass.Pkg.Path(), "domain", "service") {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || !fn.Name.IsExported() {
				continue
			}
			checkSignature(pass, fn)
		}
	}
	return nil, nil
}

func checkSignature(pass *analysis.Pass, fn *ast.FuncDecl) {
	obj, ok := pass.TypesInfo.Defs[fn.Name].(*types.Func)
	if !ok {
		return
	}
	sig, ok := obj.Type().(*types.Signature)
	if !ok {
		return
	}
	check := func(tuple *types.Tuple, kind string) {
		for v := range tuple.Variables() {
			if leaked := findEntityType(v.Type(), map[types.Type]bool{}); leaked != "" {
				pass.Reportf(fn.Name.Pos(),
					"%s: метод %q использует entity-тип %s (%s) — API сервиса принимает и возвращает model, "+
						"конвертация в entity выполняется внутри",
					ruleID, fn.Name.Name, leaked, kind)
			}
		}
	}
	check(sig.Params(), "параметр")
	check(sig.Results(), "результат")
}

// findEntityType рекурсивно ищет в типе ссылку на тип из /dal/entity
// и возвращает её имя, либо пустую строку.
func findEntityType(t types.Type, seen map[types.Type]bool) string {
	if t == nil || seen[t] {
		return ""
	}
	seen[t] = true
	switch tt := t.(type) {
	case *types.Named:
		obj := tt.Obj()
		if pkg := obj.Pkg(); pkg != nil && pathseg.Contains(pkg.Path(), "dal", "entity") {
			return pkg.Name() + "." + obj.Name()
		}
		return findEntityType(tt.Underlying(), seen)
	case *types.Alias:
		return findEntityType(types.Unalias(tt), seen)
	case *types.Pointer:
		return findEntityType(tt.Elem(), seen)
	case *types.Slice:
		return findEntityType(tt.Elem(), seen)
	case *types.Array:
		return findEntityType(tt.Elem(), seen)
	case *types.Map:
		if leaked := findEntityType(tt.Key(), seen); leaked != "" {
			return leaked
		}
		return findEntityType(tt.Elem(), seen)
	case *types.Chan:
		return findEntityType(tt.Elem(), seen)
	}
	return ""
}
