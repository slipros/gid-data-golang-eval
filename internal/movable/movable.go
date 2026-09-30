// Package movable answers whether a function of a service/usecase package can
// move onto a type of the model layer: it must not use its receiver, must not
// reach for symbols of its own package, and the type it moves onto must be a
// model struct or enum. Shared by GID-195 (gidmodelmethod) and GID-278
// (gidpuremodel), which differ only in which functions they judge.
package movable

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

// ModelNamed returns the named type of a parameter written as T or *T, where T
// is a non-interface type of the model layer (struct, enum, …). A method cannot
// be added to an interface, so an interface does not "own" behaviour.
func ModelNamed(pass *analysis.Pass, expr ast.Expr) (*types.Named, bool) {
	t := pass.TypesInfo.TypeOf(expr)
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return nil, false
	}
	if _, ok := named.Underlying().(*types.Interface); ok {
		return nil, false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	if pkg == nil || !pathseg.HasLayer(pkg.Path(), "domain", "model") {
		return nil, false
	}

	return named, true
}

// UsesReceiver reports whether the method body accesses the receiver. An
// unnamed or blank receiver is never used.
func UsesReceiver(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	if len(fn.Recv.List) == 0 || len(fn.Recv.List[0].Names) == 0 {
		return false
	}
	recv := fn.Recv.List[0].Names[0]
	if recv.Name == "_" {
		return false
	}
	obj := pass.TypesInfo.Defs[recv]
	if obj == nil || fn.Body == nil {
		return false
	}
	used := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && pass.TypesInfo.Uses[id] == obj {
			used = true
		}

		return !used
	})

	return used
}

// DependsOnPackage reports whether the function (signature and body) refers to
// package-level symbols of its own package — such a function is not movable.
func DependsOnPackage(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	self := pass.TypesInfo.Defs[fn.Name]
	depends := false
	check := func(n ast.Node) {
		ast.Inspect(n, func(node ast.Node) bool {
			id, ok := node.(*ast.Ident)
			if !ok {
				return true
			}
			obj := pass.TypesInfo.Uses[id]
			if obj == nil || obj == self || obj.Pkg() != pass.Pkg {
				return true
			}
			switch obj.(type) {
			case *types.PkgName, *types.Label:
				return true // imports and labels are not a dependency
			}
			// A package-level symbol (Parent == package scope) or a member
			// of a package type — a field/method (Parent == nil).
			if obj.Parent() == pass.Pkg.Scope() || obj.Parent() == nil {
				depends = true
			}

			return !depends
		})
	}
	check(fn.Type)
	if fn.Body != nil {
		check(fn.Body)
	}

	return depends
}

// RecvTypeName returns the name of the receiver type ("" for a function).
func RecvTypeName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}
	t := fn.Recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	if ident, ok := t.(*ast.Ident); ok {
		return ident.Name
	}

	return ""
}
