// Package modelmethod implements rule GID-195: a private function in
// service/usecase that takes a single value of a model type and does not
// depend on its own package is model behaviour: it belongs as a public
// method of that type in the model layer.
//
// The same applies to a private method of a service/usecase struct that does
// not use its receiver. Non-movable candidates are not flagged: a method
// that uses its receiver; a function referencing package-level symbols of
// its own package (including package types in the results).
package modelmethod

import (
	"go/ast"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-195"

// scopes — the layers where the rule applies. Layer roots only (pathseg.EndsWith):
// subpackages like convert/ etc. are not affected.
var scopes = [][]string{
	{"domain", "service"},
	{"domain", "usecase"},
}

// Analyzer — rule GID-195 with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings — settings of rule GID-195 from .golangci.yml.
type Settings struct {
	// Exclude — exclusions: "Function" or "Type.Method".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-195 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidmodelmethod",
		Doc: ruleID + ": a private service/usecase function over a single model value is " +
			"model behaviour; expose it as a public method of that type. Fix: move it onto the model",
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
		if ast.IsGenerated(file) || isTestFile(pass, file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			checkFunc(pass, fn, s)
		}
	}
	return nil, nil
}

func checkFunc(pass *analysis.Pass, fn *ast.FuncDecl, s Settings) {
	if fn.Name.IsExported() || fn.Name.Name == "init" || fn.Type.TypeParams != nil {
		return
	}
	if exclude.Match(s.Exclude, recvTypeName(fn), fn.Name.Name) {
		return
	}
	param, ok := singleModelParam(pass, fn)
	if !ok {
		return
	}
	// A method that uses its receiver is not movable — it legitimately
	// belongs to its struct.
	if fn.Recv != nil && usesReceiver(pass, fn) {
		return
	}
	// Dependence on package-level symbols of its own package (including
	// package types in the signature) — the function cannot be moved to model.
	if dependsOnPackage(pass, fn) {
		return
	}
	paramObj := param.Obj()
	paramPkg := paramObj.Pkg()
	display := paramPkg.Name() + "." + paramObj.Name()
	if fn.Recv != nil {
		pass.Reportf(fn.Name.Pos(),
			"%s: method %q ignores its receiver and works only with the %s value. "+
				"Fix: this is model behaviour, make it a public method of that type",
			ruleID, fn.Name.Name, display)
		return
	}
	pass.Reportf(fn.Name.Pos(),
		"%s: private function %q works only with the %s value. "+
			"Fix: this is model behaviour, make it a public method of that type",
		ruleID, fn.Name.Name, display)
}

// singleModelParam — the function's single parameter of the form T or *T,
// where T is a named type of the model layer (struct, enum, etc., not an interface).
func singleModelParam(pass *analysis.Pass, fn *ast.FuncDecl) (*types.Named, bool) {
	params := fn.Type.Params
	if params == nil || len(params.List) != 1 {
		return nil, false
	}
	field := params.List[0]
	// func f(a, b *model.T) — two values; variadic — a slice of values.
	if len(field.Names) > 1 {
		return nil, false
	}
	if _, ok := field.Type.(*ast.Ellipsis); ok {
		return nil, false
	}
	t := pass.TypesInfo.TypeOf(field.Type)
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return nil, false
	}
	// A method cannot be added to an interface — it does not "own" behaviour.
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

// usesReceiver reports whether the method body accesses the receiver.
func usesReceiver(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	if len(fn.Recv.List) == 0 || len(fn.Recv.List[0].Names) == 0 {
		return false // unnamed receiver
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

// dependsOnPackage reports whether the function (signature and body) refers
// to package-level symbols of its own package — such a function is not movable.
func dependsOnPackage(pass *analysis.Pass, fn *ast.FuncDecl) bool {
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

func inScope(pkgPath string) bool {
	for _, scope := range scopes {
		if pathseg.EndsWith(pkgPath, scope...) {
			return true
		}
	}
	return false
}

func recvTypeName(fn *ast.FuncDecl) string {
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

func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	return strings.HasSuffix(filepath.Base(pass.Fset.Position(file.Pos()).Filename), "_test.go")
}
