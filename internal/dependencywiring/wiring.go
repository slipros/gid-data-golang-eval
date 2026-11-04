// Package dependencywiring identifies concrete dependencies passed to domain services.
package dependencywiring

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

type binding struct {
	expr   ast.Expr
	index  int
	writes int
}

// Edge is a dependency supplied to a concrete service constructor or literal.
type Edge struct {
	Service    *types.Named
	Dependency *types.Named
	Pos        token.Pos
}

// Edges returns wiring visible in this package. Interface values with multiple
// assignments or no concrete origin are deliberately left unresolved.
func Edges(pass *analysis.Pass) []*Edge {
	bindings := make(map[types.Object]binding)
	skip := func(file *ast.File) bool { return ast.IsGenerated(file) || srcfile.IsTest(pass, file) }
	astwalk.Nodes(pass, []ast.Node{(*ast.ValueSpec)(nil), (*ast.AssignStmt)(nil), (*ast.IncDecStmt)(nil), (*ast.RangeStmt)(nil)}, skip,
		func(_ *ast.File, node ast.Node) {
			switch n := node.(type) {
			case *ast.ValueSpec:
				for i, name := range n.Names {
					bind(pass, bindings, name, n.Values, i, true)
				}
			case *ast.AssignStmt:
				for i, lhs := range n.Lhs {
					if id, ok := lhs.(*ast.Ident); ok {
						bind(pass, bindings, id, n.Rhs, i, pass.TypesInfo.Defs[id] != nil)
					}
				}
			case *ast.RangeStmt:
				if n.Tok != token.DEFINE {
					for _, expr := range []ast.Expr{n.Key, n.Value} {
						if id, ok := expr.(*ast.Ident); ok {
							bind(pass, bindings, id, nil, 0, false)
						}
					}
				}
			case *ast.IncDecStmt:
				if id, ok := n.X.(*ast.Ident); ok {
					bind(pass, bindings, id, nil, 0, false)
				}
			}
		})
	var edges []*Edge
	appendEdge := func(service *types.Named, expr ast.Expr) {
		dep := concrete(pass, bindings, expr, 0, make(map[types.Object]bool))
		if dep != nil {
			edges = append(edges, &Edge{Service: service, Dependency: dep, Pos: expr.Pos()})
		}
	}
	astwalk.Nodes(pass, []ast.Node{(*ast.CallExpr)(nil), (*ast.CompositeLit)(nil)}, skip,
		func(_ *ast.File, node ast.Node) {
			switch n := node.(type) {
			case *ast.CallExpr:
				fn, ok := typeutil.Callee(pass.TypesInfo, n).(*types.Func)
				if !ok || !constructor(fn.Name()) {
					return
				}
				service := resultService(pass.TypesInfo.TypeOf(n))
				if service == nil {
					return
				}
				for _, arg := range n.Args {
					appendEdge(service, arg)
				}
			case *ast.CompositeLit:
				service := resultService(pass.TypesInfo.TypeOf(n))
				if service == nil {
					return
				}
				for _, elem := range n.Elts {
					if kv, ok := elem.(*ast.KeyValueExpr); ok {
						appendEdge(service, kv.Value)
					} else {
						appendEdge(service, elem)
					}
				}
			}
		})
	return edges
}

func bind(pass *analysis.Pass, bindings map[types.Object]binding, id *ast.Ident, values []ast.Expr, index int, initial bool) {
	obj := pass.TypesInfo.ObjectOf(id)
	if obj == nil {
		return
	}
	b := bindings[obj]
	b.writes++
	if len(values) == 0 || !initial {
		b.expr = nil
		bindings[obj] = b
		return
	}
	if len(values) == 1 {
		b.expr, b.index = values[0], index
	} else if index < len(values) {
		b.expr, b.index = values[index], 0
	}
	bindings[obj] = b
}

func concrete(pass *analysis.Pass, bindings map[types.Object]binding, expr ast.Expr, index int, seen map[types.Object]bool) *types.Named {
	t := pass.TypesInfo.TypeOf(expr)
	if tuple, ok := t.(*types.Tuple); ok {
		if index >= tuple.Len() {
			return nil
		}
		item := tuple.At(index)
		t = item.Type()
	}
	if named := Struct(t); named != nil {
		return named
	}
	switch e := ast.Unparen(expr).(type) {
	case *ast.Ident:
		obj := pass.TypesInfo.ObjectOf(e)
		if seen[obj] {
			return nil
		}
		seen[obj] = true
		b := bindings[obj]
		if b.writes == 1 && b.expr != nil {
			return concrete(pass, bindings, b.expr, b.index, seen)
		}
	case *ast.CallExpr:
		if tv := pass.TypesInfo.Types[e.Fun]; tv.IsType() && len(e.Args) == 1 {
			return concrete(pass, bindings, e.Args[0], 0, seen)
		}
	}
	return nil
}

// Struct unwraps pointers and aliases and excludes interface and function types.
func Struct(t types.Type) *types.Named {
	if t == nil {
		return nil
	}
	t = types.Unalias(t)
	if ptr, ok := t.(*types.Pointer); ok {
		t = types.Unalias(ptr.Elem())
	}
	named, ok := t.(*types.Named)
	if !ok {
		return nil
	}
	if _, ok := named.Underlying().(*types.Struct); !ok {
		return nil
	}
	return named
}

// InLayer classifies a concrete dependency by its declaration package.
func InLayer(t *types.Named, layer ...string) bool {
	obj := t.Obj()
	pkg := obj.Pkg()
	return pkg != nil && pathseg.HasLayer(pkg.Path(), layer...)
}

// InRootLayer excludes helper packages such as service/convert from domain roles.
func InRootLayer(t *types.Named, layer ...string) bool {
	obj := t.Obj()
	pkg := obj.Pkg()
	return pkg != nil && pathseg.HasLayer(pkg.Path(), layer...) && pathseg.EndsWith(pkg.Path(), layer...)
}

func resultService(t types.Type) *types.Named {
	if tuple, ok := t.(*types.Tuple); ok {
		if tuple.Len() == 0 {
			return nil
		}
		item := tuple.At(0)
		t = item.Type()
	}
	named := Struct(t)
	if named == nil || !InRootLayer(named, "domain", "service") {
		return nil
	}
	obj := named.Obj()
	if strings.HasSuffix(obj.Name(), "Options") {
		return nil
	}
	return named
}

func constructor(name string) bool {
	rest, ok := strings.CutPrefix(name, "New")
	if !ok {
		rest, ok = strings.CutPrefix(name, "new")
	}
	return ok && (rest == "" || (rest[0] >= 'A' && rest[0] <= 'Z'))
}
