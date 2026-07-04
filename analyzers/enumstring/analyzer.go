// Package enumstring implements the enum styling rules:
//
//   - GID-124 (gidenumstring): every enum (a named string-based type
//     with const values) must implement the String() string method.
//   - GID-123 (gidenumbased): an enum is a named string-based type, not a
//     bare string/int. Applies in /domain/model/**, /dal/entity/** and
//     /event/dto/**. Catches an alias to a basic type (type X = string), an
//     int-enum (a named int type with ≥2 const values) and a group of
//     untyped string constants.
package enumstring

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-124"

// Analyzer is rule GID-124: an enum (string type with const values) must implement String() string. Fix: add a String() string method.
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

// enumTypesWithConsts — the package's string types that have const values.
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
