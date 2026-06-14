// Package enumplace implements rule GID-211 (enum-location):
//
//   - GID-211 (gidenumplace): DAL-layer enums live in /dal/entity/enum,
//     a separate file per entity. Declaring a string enum in any other
//     DAL-layer package (e.g. /dal/entity or /dal/repository) is a violation.
//
// An enum here is a named type with underlying string that has ≥1 const of
// this type in the same package (the same detection technique as in enumstring).
// An alias (type X = string) does not count as an enum — that is the domain of
// GID-123. The domain layer is not touched: in model an enum lives right in
// model (the norm, GID-132).
//
// Source: entity.md "Enum (entity/enum/): each enum lives in a separate
// file named after the entity".
package enumplace

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-211"

// Analyzer — rule GID-211: DAL-layer enums live in /dal/entity/enum (one file per entity). Fix: move the enum into /dal/entity/enum.
var Analyzer = &analysis.Analyzer{
	Name: "gidenumplace",
	Doc:  ruleID + ": DAL-layer enums live in /dal/entity/enum (one file per entity). Fix: move the enum into /dal/entity/enum",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()

	// Scope: only the DAL layer, excluding the canonical place /dal/entity/enum.
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
		"%s: enum %s must live in /dal/entity/enum (one file named after the entity). Fix: move it there",
		ruleID, ts.Name.Name)
}

// enumTypesWithConsts — the package's string types that have ≥1 const value.
// An alias of a basic type does not land here: an alias does not create a
// *types.Named, a const of such an alias has the universe string type. This is
// the technique from enumstring.
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
