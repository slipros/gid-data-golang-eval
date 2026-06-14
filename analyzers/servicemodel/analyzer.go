// Package servicemodel implements rule GID-151: the domain-service API works
// only with model. The service takes a model, internally converts it to an
// entity for the repository, converts the received entity back, and always
// returns a model.
//
// The check: parameters and results of exported methods in the root of
// /domain/service do not reference types from /dal/entity (recursively —
// through pointers, slices, maps, and fields). Inside the method body an
// entity is allowed — that is what conversion is for.
package servicemodel

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-151"

// Analyzer — rule GID-151: exported service methods take and return model, not entity. Fix: convert to entity internally.
var Analyzer = &analysis.Analyzer{
	Name: "gidservicemodel",
	Doc:  ruleID + ": exported service methods take and return model, not entity. Fix: convert to entity internally",
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
					"%s: method %q uses the entity type %s (%s). Fix: the service API takes and returns model, "+
						"convert to entity internally",
					ruleID, fn.Name.Name, leaked, kind)
			}
		}
	}
	check(sig.Params(), "parameter")
	check(sig.Results(), "result")
}

// findEntityType recursively searches the type for a reference to a type from
// /dal/entity and returns its name, or an empty string.
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
