// Package servicesingle implements rule GID-148: a domain service is devoted
// to one entity and does not depend on other services. Orchestrating the
// business logic of several entities is the job of a usecase, which may use
// several services.
//
// The deterministic check: in the root of /domain/service a struct field
// whose type is another struct from the same package (except *Options) means
// a service-on-service dependency.
package servicesingle

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-148"

// Analyzer — rule GID-148: a service must not depend on another service; entity orchestration happens in usecase. Fix: move orchestration to usecase.
var Analyzer = &analysis.Analyzer{
	Name: "gidservicesingle",
	Doc:  ruleID + ": a service must not depend on another service; entity orchestration happens in usecase. Fix: move orchestration to usecase",
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

// samePackageStruct returns the type name if the field type is a struct
// (or a pointer to a struct) from the same package and not an Options type.
func samePackageStruct(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	t := pass.TypesInfo.TypeOf(expr)
	if t == nil {
		return "", false
	}
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return "", false
	}
	namedObj := named.Obj()
	if namedObj.Pkg() != pass.Pkg {
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
