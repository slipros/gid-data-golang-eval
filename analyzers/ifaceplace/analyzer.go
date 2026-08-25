// Package ifaceplace implements rules GID-134 (interface-near-consumer) and
// GID-269 (no-inline-interface-field).
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
// GID-134 leaves anonymous interfaces untouched. GID-269 reports a non-empty
// anonymous interface used directly as a struct field type. Empty interface{}
// fields, error, any/interface{}, and generic constraints are untouched.
// Generated code is skipped.
//
// LoadMode: TypesInfo is needed — we detect types.Interface and the
// declaring package via Named.Obj().Pkg().
//
// Test files (GID-250 puts tests in the same package as the code under test,
// see "Test files" at the end of RULES.md). A wiring test that boots a real
// handler has no say in the type it names: it calls a production constructor
// (handler.NewCreate(validator, service CreateService)) and so must declare
// its own helper's parameters with that exact consumer-side interface, even
// though the helper itself lives in a foreign package (incident 2026-08-06,
// resource-registry: yandex_audience_wire_test.go typed a server-startup
// helper's parameters as handler.CreateService/handler.IntegrationService/…
// only because NewCreate demands it — the test had no other legal type to
// write). That is a *use* of the interface in a parameter/result list
// (checkFuncDecl) — skipped in _test.go files via srcfile.IsTest.
//
// A struct type *declared* in a _test.go file is a different case: nothing
// forces its field types. A test picking a foreign "own"-package interface
// for a struct field chose to — it was never handed that type by a call it
// has to match. checkTypeDecl (struct fields) therefore keeps checking
// _test.go files exactly like production ones; only checkFuncDecl
// (parameters/results) skips them.
package ifaceplace

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const (
	ruleID                = "GID-134"
	inlineInterfaceRuleID = "GID-269"
)

// layerSegments — path segments by which a package is recognized as the
// service's "own" layer (rather than stdlib/an external library).
var layerSegments = []string{
	"dal", "domain", "client", "server", "event", "app", "metric",
}

// Analyzer checks interface placement rules GID-134 and GID-269.
var Analyzer = &analysis.Analyzer{
	Name: "gidifaceplace",
	Doc: ruleID + ": interfaces live where they are used; " +
		"define the interface next to its consumer (exceptions: libraries, /domain/model for service/usecase, " +
		"and a _test.go helper's parameters/results dictated by a production constructor); " +
		inlineInterfaceRuleID + ": struct fields use named interfaces instead of inline declarations",
	Run: run,
}

func run(pass *analysis.Pass) (any, error) {
	consumerPkg := pass.Pkg
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		// A _test.go file cannot be exempted wholesale: struct fields it
		// declares are a free choice (checked as usual), while a helper's
		// parameters/results are dictated by the production constructor it
		// wires up and must be skipped. See the package doc.
		isTest := srcfile.IsTest(pass, file)
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				checkTypeDecl(pass, consumerPkg, d)
			case *ast.FuncDecl:
				if isTest {
					continue
				}
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
			if iface, ok := field.Type.(*ast.InterfaceType); ok && iface.Methods != nil && len(iface.Methods.List) > 0 {
				pass.Reportf(field.Type.Pos(),
					"%s: anonymous interface is declared in a struct field. "+
						"Fix: declare a named interface next to the struct and use it as the field type",
					inlineInterfaceRuleID)
			}
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
	if pathseg.HasLayer(declPath, "domain", "model") && isServiceOrUsecase(consumer.Path()) {
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
		if pathseg.HasLayer(path, seg) {
			return true
		}
	}
	return false
}

// isServiceOrUsecase reports that the consumer is the domain/service
// or domain/usecase layer.
func isServiceOrUsecase(path string) bool {
	return pathseg.HasLayer(path, "domain", "service") ||
		pathseg.HasLayer(path, "domain", "usecase")
}
