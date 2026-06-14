// Package privatefunc implements rule GID-133: in service, usecase, and
// repository there are no private functions belonging to the package — a
// private function must be a struct method. The exception: a function used by
// methods of several entities of the same package (a shared helper).
//
// New<Entity> constructors and the private functions used from them are
// considered to belong to their entity.
package privatefunc

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-133"

var scopes = [][]string{
	{"dal", "repository"},
	{"domain", "service"},
	{"domain", "usecase"},
}

// Analyzer — rule GID-133: private functions in service/usecase/repository are struct methods.
var Analyzer = &analysis.Analyzer{
	Name: "gidprivatefunc",
	Doc:  ruleID + ": private functions in service/usecase/repository must be struct methods, not package functions. Fix: make it a method",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if !inScope(pass.Pkg.Path()) {
		return nil, nil
	}
	structs := packageStructs(pass)
	candidates := privateFuncs(pass)
	if len(candidates) == 0 {
		return nil, nil
	}
	owners := collectOwners(pass, structs, candidates)
	for _, fn := range candidates {
		obj := pass.TypesInfo.Defs[fn.Name]
		used := owners[obj]
		switch len(used) {
		case 0:
			pass.Reportf(fn.Name.Pos(),
				"%s: private function %q belongs to the package. Fix: make it a struct method "+
					"(only a function shared by several entities may stay package-level)",
				ruleID, fn.Name.Name)
		case 1:
			pass.Reportf(fn.Name.Pos(),
				"%s: private function %q is used only by entity %q. Fix: make it a method",
				ruleID, fn.Name.Name, soleKey(used))
		}
	}
	return nil, nil
}

// privateFuncs — private package-level functions (violation candidates).
func privateFuncs(pass *analysis.Pass) []*ast.FuncDecl {
	var out []*ast.FuncDecl
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name.IsExported() || fn.Name.Name == "init" {
				continue
			}
			out = append(out, fn)
		}
	}
	return out
}

// collectOwners: which entities use each private function.
func collectOwners(
	pass *analysis.Pass,
	structs map[string]struct{},
	candidates []*ast.FuncDecl,
) map[types.Object]map[string]struct{} {
	objs := map[types.Object]struct{}{}
	for _, fn := range candidates {
		objs[pass.TypesInfo.Defs[fn.Name]] = struct{}{}
	}
	owners := map[types.Object]map[string]struct{}{}
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			owner := ownerOf(fn, structs)
			if owner == "" {
				continue // usage from other package functions does not define an entity
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				id, ok := n.(*ast.Ident)
				if !ok {
					return true
				}
				obj := pass.TypesInfo.Uses[id]
				if obj == nil {
					return true
				}
				if _, isCandidate := objs[obj]; !isCandidate {
					return true
				}
				if owners[obj] == nil {
					owners[obj] = map[string]struct{}{}
				}
				owners[obj][owner] = struct{}{}
				return true
			})
		}
	}
	return owners
}

// ownerOf — the entity the function belongs to: a method's receiver
// or the entity of a New<Entity> constructor.
func ownerOf(fn *ast.FuncDecl, structs map[string]struct{}) string {
	if fn.Recv != nil {
		return recvTypeName(fn)
	}
	entity, ok := strings.CutPrefix(fn.Name.Name, "New")
	if !ok || entity == "" {
		return ""
	}
	if _, ok := structs[entity]; !ok {
		return ""
	}
	return entity
}

func packageStructs(pass *analysis.Pass) map[string]struct{} {
	out := map[string]struct{}{}
	for _, file := range pass.Files {
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
				if _, ok := ts.Type.(*ast.StructType); ok {
					out[ts.Name.Name] = struct{}{}
				}
			}
		}
	}
	return out
}

func inScope(pkgPath string) bool {
	for _, scope := range scopes {
		if pathseg.EndsWith(pkgPath, scope...) {
			return true
		}
	}
	return false
}

func recvTypeName(fn *ast.FuncDecl) string {
	t := fn.Recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	if ident, ok := t.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

func soleKey(m map[string]struct{}) string {
	for k := range m {
		return k
	}
	return ""
}
