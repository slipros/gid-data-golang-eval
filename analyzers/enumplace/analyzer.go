// Package enumplace реализует правило GID-211 (enum-location):
//
//   - GID-211 (gidenumplace): enum DAL-слоя живут в /dal/entity/enum,
//     отдельный файл на сущность. Объявление string-enum в любом другом
//     пакете DAL-слоя (например /dal/entity или /dal/repository) — нарушение.
//
// Enum здесь — именованный тип с underlying string, у которого в том же
// пакете есть ≥1 const этого типа (та же техника детекции, что в enumstring).
// Alias (type X = string) не считается enum — это зона GID-123. Domain-слой
// не задевается: в model enum живёт прямо в model (норма, GID-132).
//
// Источник: entity.md «Enum (entity/enum/): каждый enum живёт в отдельном
// файле по имени сущности».
package enumplace

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-211"

// Analyzer — правило GID-211: enum DAL-слоя живут в /dal/entity/enum (отдельный файл на сущность).
var Analyzer = &analysis.Analyzer{
	Name: "gidenumplace",
	Doc:  ruleID + ": enum DAL-слоя живут в /dal/entity/enum (отдельный файл на сущность)",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()

	// Scope: только DAL-слой, исключая каноническое место /dal/entity/enum.
	if !pathseg.Contains(pkgPath, "dal") {
		return nil, nil
	}
	if pathseg.Contains(pkgPath, "dal", "entity", "enum") {
		return nil, nil
	}

	withConsts := enumTypesWithConsts(pass)
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				checkEnum(pass, ts, withConsts)
			}
		}
	}
	return nil, nil
}

func checkEnum(pass *analysis.Pass, ts *ast.TypeSpec, withConsts map[*types.Named]struct{}) {
	obj, ok := pass.TypesInfo.Defs[ts.Name].(*types.TypeName)
	if !ok {
		return
	}
	named, ok := obj.Type().(*types.Named)
	if !ok {
		return
	}
	if _, isEnum := withConsts[named]; !isEnum {
		return
	}
	pass.Reportf(ts.Name.Pos(),
		"%s: enum %s живёт в /dal/entity/enum (отдельный файл по имени сущности)",
		ruleID, ts.Name.Name)
}

// enumTypesWithConsts — string-типы пакета, имеющие ≥1 const-значение.
// Alias на basic-тип сюда не попадает: alias не создаёт *types.Named,
// const такого alias имеет тип universe string. Это техника из enumstring.
func enumTypesWithConsts(pass *analysis.Pass) map[*types.Named]struct{} {
	out := map[*types.Named]struct{}{}
	for _, obj := range pass.TypesInfo.Defs {
		c, ok := obj.(*types.Const)
		if !ok {
			continue
		}
		named, ok := c.Type().(*types.Named)
		if !ok {
			continue
		}
		namedObj := named.Obj()
		if namedObj.Pkg() != pass.Pkg {
			continue
		}
		basic, ok := named.Underlying().(*types.Basic)
		if !ok || basic.Kind() != types.String {
			continue
		}
		out[named] = struct{}{}
	}
	return out
}
