// Package enumstring реализует правила оформления enum:
//
//   - GID-124 (gidenumstring): каждый enum (именованный тип на основе string
//     с const-значениями) обязан реализовать метод String() string.
//   - GID-123 (gidenumbased): enum — именованный тип на основе string, не
//     голый string/int. Действует в /domain/model/** и /dal/entity/**.
//     Ловит alias на basic-тип (type X = string), int-enum (именованный
//     int-тип с ≥2 const-значений) и группу нетипизированных string-констант.
package enumstring

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-124"

// Analyzer — правило GID-124: an enum (string type with const values) must implement String() string. Fix: add a String() string method.
var Analyzer = &analysis.Analyzer{
	Name: "gidenumstring",
	Doc:  ruleID + ": an enum (string type with const values) must implement String() string. Fix: add a String() string method",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
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
	if hasStringMethod(named) {
		return
	}
	pass.Reportf(ts.Name.Pos(),
		"%s: enum %s must implement the String() string method. Fix: add a String() string method", ruleID, ts.Name.Name)
}

// enumTypesWithConsts — string-типы пакета, имеющие const-значения.
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

func hasStringMethod(named *types.Named) bool {
	for m := range named.Methods() {
		if m.Name() != "String" {
			continue
		}
		sig, ok := m.Type().(*types.Signature)
		if !ok {
			continue
		}
		params := sig.Params()
		results := sig.Results()
		if params.Len() != 0 || results.Len() != 1 {
			continue
		}
		result0 := results.At(0)
		resultType := result0.Type()
		basic, ok := resultType.(*types.Basic)
		if ok && basic.Kind() == types.String {
			return true
		}
	}
	return false
}
