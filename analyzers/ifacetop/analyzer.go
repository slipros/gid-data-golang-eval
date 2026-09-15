// Package ifacetop implements rule GID-276: interfaces open the file. In the
// layers that declare their dependencies as interfaces — /domain/service,
// /domain/usecase, /client, /dal/repository — every top-level interface type of
// a file is declared above every other type and every function: the reader
// sees what the entity depends on before the entity itself.
//
// import, const and var are not boundaries: GID-130 already puts them above
// types and functions. An interface is recognised by type information, so an
// alias or a defined type over an interface (type Reader = io.Reader) counts.
//
// A _test.go file is not judged: its composition — a double, the fixture it is
// wired into — is the test's own, as in GID-157.
package ifacetop

import (
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const ruleID = "GID-276"

// Analyzer — rule GID-276: interfaces are declared at the top of the file, above other types and functions.
var Analyzer = &analysis.Analyzer{
	Name: "gidifacetop",
	Doc:  ruleID + ": in service, usecase, client and repository layers interfaces are declared at the top of the file, right after import, const and var. Fix: move type HelloRepository interface above type Hello struct",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if !inScope(pass.Pkg.Path()) {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) || srcfile.IsTest(pass, file) {
			continue
		}
		checkFile(pass, file)
	}
	return nil, nil
}

func checkFile(pass *analysis.Pass, file *ast.File) {
	var boundary ast.Node
	for _, n := range typesAndFuncs(file) {
		ts, isType := n.(*ast.TypeSpec)
		switch {
		case isType && isInterface(pass, ts):
			if boundary != nil {
				report(pass, ts, boundary)
			}
		case boundary == nil:
			boundary = n
		}
	}
}

// typesAndFuncs lists the file's top-level type specs and functions in source
// order; a grouped type ( … ) block contributes each of its specs.
func typesAndFuncs(file *ast.File) []ast.Node {
	var out []ast.Node
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			out = append(out, d)
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					out = append(out, ts)
				}
			}
		}
	}
	return out
}

func report(pass *analysis.Pass, iface *ast.TypeSpec, boundary ast.Node) {
	head := declHead(boundary)
	pass.Reportf(iface.Name.Pos(),
		"%s: interface %s is declared below %s (line %d); interfaces open the file, right after import, const and var. Fix: move type %s interface above %s",
		ruleID, iface.Name.Name, head, pass.Fset.Position(boundary.Pos()).Line, iface.Name.Name, head)
}

func isInterface(pass *analysis.Pass, ts *ast.TypeSpec) bool {
	obj := pass.TypesInfo.Defs[ts.Name]
	if obj == nil {
		_, ok := ts.Type.(*ast.InterfaceType)
		return ok
	}
	typ := obj.Type()
	_, ok := typ.Underlying().(*types.Interface)
	return ok
}

// declHead names a declaration the way it opens in the source: type Hello,
// func NewHello, func (*Hello) Create.
func declHead(n ast.Node) string {
	switch d := n.(type) {
	case *ast.TypeSpec:
		return "type " + d.Name.Name
	case *ast.FuncDecl:
		if d.Recv == nil || len(d.Recv.List) == 0 {
			return "func " + d.Name.Name
		}
		return fmt.Sprintf("func (%s) %s", types.ExprString(d.Recv.List[0].Type), d.Name.Name)
	}
	return ""
}

// inScope reports whether the package belongs to a layer that declares its
// dependencies as interfaces. The layer is anchored to the module root
// (pathseg.HasLayer), so a nested .../app/client is not the client layer.
func inScope(path string) bool {
	return pathseg.HasLayer(path, "domain", "service") ||
		pathseg.HasLayer(path, "domain", "usecase") ||
		pathseg.HasLayer(path, "client") ||
		pathseg.HasLayer(path, "dal", "repository")
}
