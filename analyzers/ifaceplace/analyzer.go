// Package ifaceplace implements rule GID-134 (interface-near-consumer):
// interfaces live where they are used.
//
// The check: if a named interface type is used in struct fields or in the
// parameters/results of a function (method) of the package, we look at the
// package where that interface is declared:
//
//   - the same package — OK (the interface is defined next to the consumer);
//   - stdlib or an external library — OK. The service's "own" package is
//     told apart from a library by path segments (pathseg): if the path
//     contains a layer segment (dal, domain, client, server, event, app,
//     metric) — it is our package; otherwise — a library;
//   - an interface from the model layer (/domain/model, including
//     subpackages) — OK, but only if the consumer is in the /domain/service
//     or /domain/usecase layer; for other consumers it is a violation;
//   - any other "own" package — a violation.
//
// Untouched: anonymous interfaces, error, any/interface{},
// generic constraints. Generated code is skipped.
//
// LoadMode: TypesInfo is needed — we detect types.Interface and the
// declaring package via Named.Obj().Pkg().
package ifaceplace

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-134"

// layerSegments — path segments by which a package is recognized as the
// service's "own" layer (rather than stdlib/an external library).
var layerSegments = []string{
	"dal", "domain", "client", "server", "event", "app", "metric",
}

// Analyzer — rule GID-134: interfaces live where they are used.
var Analyzer = &analysis.Analyzer{
	Name: "gidifaceplace",
	Doc: ruleID + ": interfaces live where they are used; " +
		"define the interface next to its consumer (exceptions: libraries and /domain/model for service/usecase)",
	Run: run,
}

func run(pass *analysis.Pass) (any, error) {
	consumerPkg := pass.Pkg
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				checkTypeDecl(pass, consumerPkg, d)
			case *ast.FuncDecl:
				checkFuncDecl(pass, consumerPkg, d)
			}
		}
	}
	return nil, nil
}

// checkTypeDecl checks the fields of struct types in a type declaration.
func checkTypeDecl(pass *analysis.Pass, consumer *types.Package, gd *ast.GenDecl) {
	for _, spec := range gd.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok || st.Fields == nil {
			continue
		}
		for _, field := range st.Fields.List {
			checkExpr(pass, consumer, field.Type)
		}
	}
}

// checkFuncDecl checks the parameters and results of a function/method.
func checkFuncDecl(pass *analysis.Pass, consumer *types.Package, fn *ast.FuncDecl) {
	if fn.Type == nil {
		return
	}
	checkFieldList(pass, consumer, fn.Type.Params)
	checkFieldList(pass, consumer, fn.Type.Results)
}

func checkFieldList(pass *analysis.Pass, consumer *types.Package, fl *ast.FieldList) {
	if fl == nil {
		return
	}
	for _, field := range fl.List {
		checkExpr(pass, consumer, field.Type)
	}
}

// checkExpr examines the type expression of a position (field/parameter/result).
// Only a named interface type declared in someone else's "own" package is
// flagged. Anonymous interfaces (ast.InterfaceType) do not get here — they
// have no *types.Named, hence no declaring package.
func checkExpr(pass *analysis.Pass, consumer *types.Package, expr ast.Expr) {
	tv, ok := pass.TypesInfo.Types[expr]
	if !ok {
		return
	}
	named, ok := tv.Type.(*types.Named)
	if !ok {
		return // an anonymous interface, a basic type, non-named — skipped
	}
	obj := named.Obj()
	if obj == nil {
		return
	}
	// error and other builtin named types: no package.
	declPkg := obj.Pkg()
	if declPkg == nil {
		return
	}
	// Only interfaces are of interest.
	if _, isIface := named.Underlying().(*types.Interface); !isIface {
		return
	}

	declPath := declPkg.Path()
	// The same package — the interface is defined next to the consumer.
	if declPkg == consumer {
		return
	}
	// A library (stdlib / an external module) — the path has no layer segments.
	if !isOwnPackage(declPath) {
		return
	}
	// The model layer: allowed only for service/usecase consumers.
	if pathseg.Contains(declPath, "domain", "model") && isServiceOrUsecase(consumer.Path()) {
		return
	}
	// Someone else's "own" package (or the model layer for non service/usecase) — a violation.
	pass.Reportf(expr.Pos(),
		"%s: interface %s is declared in %s. Fix: define the interface next to its consumer "+
			"(exceptions: libraries and /domain/model for service/usecase)",
		ruleID, obj.Name(), declPath)
}

// isOwnPackage reports that the package is our service layer (not a library):
// the path contains at least one layer segment.
func isOwnPackage(path string) bool {
	for _, seg := range layerSegments {
		if pathseg.Contains(path, seg) {
			return true
		}
	}
	return false
}

// isServiceOrUsecase reports that the consumer is the domain/service
// or domain/usecase layer.
func isServiceOrUsecase(path string) bool {
	return pathseg.Contains(path, "domain", "service") ||
		pathseg.Contains(path, "domain", "usecase")
}
