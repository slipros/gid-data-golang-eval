// Package servicesingle implements rule GID-148: a domain service is devoted
// to one entity and does not depend on other services. Orchestrating the
// business logic of several entities is the job of a usecase, which may use
// several services.
//
// The deterministic check: in the root of /domain/service a struct field
// whose type is another domain service/usecase struct (except *Options) means
// a service-on-service dependency. Constructing a different domain service
// inside a method (literal, new, or New* factory) is also a dependency.
// Production wiring also checks concrete service/usecase implementations passed
// through interfaces to a service constructor or literal.
//
// A _test.go file is not judged: tests live in the same package (GID-250), and
// a test harness — a capture struct holding the service under test, a fixture
// bundling several doubles — is composition of the test, not a dependency of a
// service on a service.
package servicesingle

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/dependencywiring"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const ruleID = "GID-148"

// Analyzer — rule GID-148: a service must not depend on another service; entity orchestration happens in usecase. Fix: move orchestration to usecase.
var Analyzer = &analysis.Analyzer{
	Name:     "gidservicesingle",
	Doc:      ruleID + ": a service must not depend on another service; entity orchestration happens in usecase. Fix: move orchestration to usecase",
	Run:      run,
	Requires: astwalk.Requires,
}

func run(pass *analysis.Pass) (any, error) {
	for _, edge := range dependencywiring.Edges(pass) {
		dep := edge.Dependency.Obj()
		if strings.HasSuffix(dep.Name(), "Options") {
			continue
		}
		if dependencywiring.InRootLayer(edge.Dependency, "domain", "service") || dependencywiring.InRootLayer(edge.Dependency, "domain", "usecase") {
			owner := edge.Service.Obj()
			pass.Reportf(edge.Pos, "%s: service %q receives domain dependency %q. Fix: move service orchestration to usecase", ruleID, owner.Name(), dep.Name())
		}
	}

	if !pathseg.EndsWith(pass.Pkg.Path(), "domain", "service") {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) || srcfile.IsTest(pass, file) {
			continue
		}
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv != nil && fn.Body != nil {
				checkConstruction(pass, fn)
			}
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				checkServiceStruct(pass, ts.Name.Name, st)
			}
		}
	}
	return nil, nil
}

func checkServiceStruct(pass *analysis.Pass, owner string, st *ast.StructType) {
	for _, field := range st.Fields.List {
		dep, ok := samePackageStruct(pass, field.Type)
		if !ok {
			continue
		}
		pass.Reportf(field.Pos(),
			"%s: service %q depends on service %q. Fix: a service serves one entity, "+
				"orchestrate multiple services in usecase",
			ruleID, owner, dep)
	}
}

// samePackageStruct recognises the local service and concrete domain dependencies,
// including those declared in another module. Options are not services.
func samePackageStruct(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	t := pass.TypesInfo.TypeOf(expr)
	if t == nil {
		return "", false
	}
	t = types.Unalias(t)
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return "", false
	}
	namedObj := named.Obj()
	if namedObj.Pkg() != pass.Pkg && !dependencywiring.InRootLayer(named, "domain", "service") && !dependencywiring.InRootLayer(named, "domain", "usecase") {
		return "", false
	}
	if _, ok := named.Underlying().(*types.Struct); !ok {
		return "", false
	}
	name := namedObj.Name()
	if strings.HasSuffix(name, "Options") {
		return "", false
	}
	return name, true
}

func checkConstruction(pass *analysis.Pass, fn *ast.FuncDecl) {
	owner, ok := samePackageStruct(pass, fn.Recv.List[0].Type)
	if !ok {
		return
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		var expr ast.Expr
		switch node := n.(type) {
		case *ast.CompositeLit:
			expr = node
		case *ast.CallExpr:
			callee := typeutil.Callee(pass.TypesInfo, node)
			switch c := callee.(type) {
			case *types.Builtin:
				if c.Name() != "new" {
					return true
				}
			case *types.Func:
				if !strings.HasPrefix(c.Name(), "New") {
					return true
				}
			default:
				return true
			}
			expr = node
		default:
			return true
		}
		dep, ok := samePackageStruct(pass, expr)
		if ok && dep != owner {
			pass.Reportf(expr.Pos(), "%s: service %q constructs service %q inside a method. Fix: move service construction and orchestration to usecase", ruleID, owner, dep)
		}
		return true
	})
}
