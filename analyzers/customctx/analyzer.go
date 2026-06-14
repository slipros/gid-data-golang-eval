// Package customctx implements rule GID-188: a ban on custom
// context types. Per Google's decision ("custom contexts — no exceptions"),
// only the stdlib type context.Context is allowed in the ctx parameter
// position and in interface embedding. Data is passed via context.WithValue
// (the helpers live in /domain/model — GID-165/166).
package customctx

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-188"

// Analyzer — rule GID-188: custom context types are banned — only context.Context.
var Analyzer = &analysis.Analyzer{
	Name: "gidcustomctx",
	Doc:  ruleID + ": custom context types are forbidden, use context.Context. Fix: pass context.Context and store data via context.WithValue.",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	ctxIface := lookupContextInterface(pass)

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				checkTypeDecls(pass, d, ctxIface)
			case *ast.FuncDecl:
				checkFuncParams(pass, d.Type)
			}
		}

		// Parameters of function literals and function types are checked too.
		ast.Inspect(file, func(n ast.Node) bool {
			if lit, ok := n.(*ast.FuncLit); ok {
				checkFuncParams(pass, lit.Type)
			}
			return true
		})
	}
	return nil, nil
}

// lookupContextInterface returns the underlying interface of the stdlib type
// context.Context if the context package is imported (directly or
// transitively). Otherwise nil — checks 1 and 2 are not applicable.
func lookupContextInterface(pass *analysis.Pass) *types.Interface {
	for _, imp := range allImports(pass.Pkg) {
		if imp.Path() != "context" {
			continue
		}
		scope := imp.Scope()
		obj := scope.Lookup("Context")
		if obj == nil {
			return nil
		}
		named, ok := obj.Type().(*types.Named)
		if !ok {
			return nil
		}
		iface, ok := named.Underlying().(*types.Interface)
		if !ok {
			return nil
		}
		return iface
	}
	return nil
}

func allImports(pkg *types.Package) []*types.Package {
	seen := map[string]bool{}
	var out []*types.Package
	var walk func(p *types.Package)
	walk = func(p *types.Package) {
		for _, imp := range p.Imports() {
			if seen[imp.Path()] {
				continue
			}
			seen[imp.Path()] = true
			out = append(out, imp)
			walk(imp)
		}
	}
	walk(pkg)
	return out
}

// isStdlibContext reports whether the type is exactly the stdlib context.Context.
func isStdlibContext(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	return pkg != nil && pkg.Path() == "context" && obj.Name() == "Context"
}

// checkTypeDecls checks the type declarations in the package under analysis:
//   - a named type (struct/interface/any) whose method set
//     covers context.Context (case 1);
//   - an interface type embedding context.Context (case 2).
func checkTypeDecls(pass *analysis.Pass, gen *ast.GenDecl, ctxIface *types.Interface) {
	for _, spec := range gen.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}

		// Case 2: an interface embedding context.Context.
		if iface, ok := ts.Type.(*ast.InterfaceType); ok && embedsStdlibContext(pass, iface) {
			pass.Reportf(ts.Pos(),
				"%s: custom context type %s is forbidden. "+
					"Fix: pass context.Context and store data via context.WithValue "+
					"(helpers live in /domain/model, GID-165/166).",
				ruleID, ts.Name.Name)
			continue
		}

		// Case 1: a type implementing context.Context via its method set.
		if ctxIface == nil {
			continue
		}
		obj := pass.TypesInfo.Defs[ts.Name]
		if obj == nil {
			continue
		}
		named, ok := obj.Type().(*types.Named)
		if !ok {
			continue
		}
		// The stdlib context.Context itself does not count (it is not from our
		// package, but just in case avoid self-reference).
		if isStdlibContext(named) {
			continue
		}
		if types.Implements(named, ctxIface) || types.Implements(types.NewPointer(named), ctxIface) {
			pass.Reportf(ts.Pos(),
				"%s: custom context type %s is forbidden. "+
					"Fix: pass context.Context and store data via context.WithValue "+
					"(helpers live in /domain/model, GID-165/166).",
				ruleID, ts.Name.Name)
		}
	}
}

// embedsStdlibContext checks whether the interface declaration embeds
// the stdlib context.Context (an embedded field without a name).
func embedsStdlibContext(pass *analysis.Pass, iface *ast.InterfaceType) bool {
	if iface.Methods == nil {
		return false
	}
	for _, field := range iface.Methods.List {
		if len(field.Names) != 0 {
			continue // an ordinary method, not embedding
		}
		if isStdlibContext(pass.TypesInfo.TypeOf(field.Type)) {
			return true
		}
	}
	return false
}

// checkFuncParams checks case 3: a parameter named ctx whose type is
// a named non-stdlib context type.
func checkFuncParams(pass *analysis.Pass, ft *ast.FuncType) {
	if ft == nil || ft.Params == nil {
		return
	}
	for _, field := range ft.Params.List {
		hasCtxName := false
		for _, name := range field.Names {
			if name.Name == "ctx" {
				hasCtxName = true
				break
			}
		}
		if !hasCtxName {
			continue
		}

		t := pass.TypesInfo.TypeOf(field.Type)
		if t == nil || isStdlibContext(t) {
			continue
		}
		// Only a named type (non-stdlib) is of interest. Anonymous
		// types/built-ins are not our case.
		if _, ok := t.(*types.Named); !ok {
			continue
		}
		pass.Reportf(field.Type.Pos(),
			"%s: parameter ctx has type %s. Fix: use context.Context.",
			ruleID, t.String())
	}
}
