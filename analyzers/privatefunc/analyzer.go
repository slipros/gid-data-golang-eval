// Package privatefunc реализует правило GID-133: в service, usecase и
// repository нет приватных функций, принадлежащих пакету, — приватная
// функция обязана быть методом структуры. Исключение: функция,
// используемая методами нескольких сущностей одного пакета (общий хелпер).
//
// Конструкторы New<Entity> и приватные функции, используемые из них,
// считаются принадлежащими своей сущности.
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

// Analyzer — правило GID-133: приватные функции в service/usecase/repository — методы структур.
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

// privateFuncs — приватные package-level функции (кандидаты в нарушители).
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

// collectOwners: какие сущности используют каждую приватную функцию.
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
				continue // использование из других функций пакета сущность не задаёт
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

// ownerOf — сущность, которой принадлежит функция: ресивер метода
// или сущность конструктора New<Entity>.
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
