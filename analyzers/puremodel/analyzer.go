// Package puremodel implements rule GID-278: a function or method of a
// service/usecase that takes values of the model layer and touches neither its
// receiver nor anything of its own package is pure logic over those values —
// model behaviour. It belongs as a public method of the model type, and the
// service that only carried it is not needed.
//
// It is the wider sibling of GID-195 (gidmodelmethod): that one judges a
// private function with exactly one model parameter, this one judges the rest
// — exported ones and ones with several parameters
// (SegmentChanged(current, snapshot *model.SegmentMetric) bool).
//
// Not movable, hence not flagged: a method that uses its receiver; a function
// that refers to symbols of its own package (including package types in the
// signature); one that takes an interface, func or channel (I/O in disguise,
// context.Context included); a method that implements an interface declared in
// the same package; a generic function.
package puremodel

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/movable"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const ruleID = "GID-278"

// scopes — the layers where the rule applies. Layer roots only (pathseg.EndsWith):
// subpackages like convert/ are not affected.
var scopes = [][]string{
	{"domain", "service"},
	{"domain", "usecase"},
}

// Analyzer — rule GID-278 with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings — settings of rule GID-278 from .golangci.yml.
type Settings struct {
	// Exclude — exclusions: "Function" or "Type.Method".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-278 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidpuremodel",
		Doc: ruleID + ": a service/usecase function that uses neither its receiver nor its package and " +
			"works only with model values is model behaviour. " +
			"Fix: make it a public method of the model type and drop the service",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s)
		},
	}
}

func run(pass *analysis.Pass, s Settings) (any, error) {
	if !inScope(pass.Pkg.Path()) {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) || srcfile.IsTest(pass, file) {
			continue
		}
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				checkFunc(pass, fn, s)
			}
		}
	}

	return nil, nil
}

func checkFunc(pass *analysis.Pass, fn *ast.FuncDecl, s Settings) {
	if fn.Body == nil || fn.Name.Name == "init" || fn.Name.Name == "_" || fn.Type.TypeParams != nil {
		return
	}
	recvName := movable.RecvTypeName(fn)
	if exclude.Match(s.Exclude, recvName, fn.Name.Name) {
		return
	}
	// A private function with exactly one parameter is GID-195's.
	if !fn.Name.IsExported() && paramCount(fn) == 1 {
		return
	}
	owner, ok := ownerParam(pass, fn)
	if !ok {
		return
	}
	if fn.Recv != nil && (movable.UsesReceiver(pass, fn) || implementsLocalInterface(pass, fn, recvName)) {
		return
	}
	if movable.DependsOnPackage(pass, fn) {
		return
	}
	obj := owner.Obj()
	pkg := obj.Pkg()
	display := pkg.Name() + "." + obj.Name()
	if fn.Recv != nil {
		pass.Reportf(fn.Name.Pos(),
			"%s: method %q uses neither its receiver nor the package, it is pure logic over %s. "+
				"Fix: make it a public method of %s and drop the service that only carries it",
			ruleID, fn.Name.Name, display, display)

		return
	}
	pass.Reportf(fn.Name.Pos(),
		"%s: function %q uses nothing of the package, it is pure logic over %s. "+
			"Fix: make it a public method of %s",
		ruleID, fn.Name.Name, display, display)
}

// ownerParam — the first parameter of the form T or *T, T a model struct/enum,
// provided every parameter is plain data: an interface, func or channel
// anywhere in a parameter type (context.Context, a repository, a callback)
// means the function talks to the outside and cannot move.
func ownerParam(pass *analysis.Pass, fn *ast.FuncDecl) (*types.Named, bool) {
	params := fn.Type.Params
	if params == nil {
		return nil, false
	}
	var owner *types.Named
	for _, field := range params.List {
		t := pass.TypesInfo.TypeOf(field.Type)
		if t == nil || !isData(t) {
			return nil, false
		}
		if owner != nil {
			continue
		}
		if _, variadic := field.Type.(*ast.Ellipsis); variadic {
			continue
		}
		if named, ok := movable.ModelNamed(pass, field.Type); ok && ownModule(pass, named) {
			owner = named
		}
	}

	return owner, owner != nil
}

// ownModule reports whether the type is declared in the module under analysis:
// a method cannot be added to a type of a foreign module. Without module
// information (GOPATH mode, eval fixtures) the type is taken as own.
func ownModule(pass *analysis.Pass, named *types.Named) bool {
	if pass.Module == nil || pass.Module.Path == "" {
		return true
	}
	obj := named.Obj()
	pkg := obj.Pkg()

	path := pkg.Path()
	root := pass.Module.Path

	return path == root || strings.HasPrefix(path, root+"/")
}

// isData reports whether t carries values only: no interface, func or channel
// at any depth of pointers, slices, arrays and maps.
func isData(t types.Type) bool {
	alias := types.Unalias(t)
	switch u := alias.Underlying().(type) {
	case *types.Interface, *types.Signature, *types.Chan:
		return false
	case *types.Pointer:
		return isData(u.Elem())
	case *types.Slice:
		return isData(u.Elem())
	case *types.Array:
		return isData(u.Elem())
	case *types.Map:
		return isData(u.Key()) && isData(u.Elem())
	default:
		return true
	}
}

// implementsLocalInterface reports whether the method is required by an
// interface declared in the same package that its receiver type implements —
// then it cannot leave the receiver type.
func implementsLocalInterface(pass *analysis.Pass, fn *ast.FuncDecl, recvName string) bool {
	scope := pass.Pkg.Scope()
	recv, ok := scope.Lookup(recvName).(*types.TypeName)
	if !ok {
		return false
	}
	ptr := types.NewPointer(recv.Type())
	for _, name := range scope.Names() {
		tn, ok := scope.Lookup(name).(*types.TypeName)
		if !ok {
			continue
		}
		tnType := tn.Type()
		iface, ok := tnType.Underlying().(*types.Interface)
		if !ok || !types.Implements(ptr, iface) {
			continue
		}
		for m := range iface.Methods() {
			if m.Name() == fn.Name.Name {
				return true
			}
		}
	}

	return false
}

// paramCount — the number of parameters: each name counts, an unnamed one too.
func paramCount(fn *ast.FuncDecl) int {
	if fn.Type.Params == nil {
		return 0
	}
	n := 0
	for _, field := range fn.Type.Params.List {
		n += max(len(field.Names), 1)
	}

	return n
}

func inScope(pkgPath string) bool {
	for _, scope := range scopes {
		if pathseg.EndsWith(pkgPath, scope...) {
			return true
		}
	}

	return false
}
