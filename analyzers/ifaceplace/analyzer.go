// Package ifaceplace implements rules GID-134 (interface-near-consumer),
// GID-269 (no-inline-interface-field) and GID-271 (iface-file-single-consumer).
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
// chose to — it was never handed that type by a call it has to match.
// checkTypeDecl (struct fields) therefore keeps checking _test.go files
// exactly like production ones; only checkFuncDecl (parameters/results)
// skips them.
//
// GID-271 judges the FILE, not the package: a file whose top-level
// declarations are only interfaces (imports don't count) is a "port file".
// GID-134 stays silent on it — the interfaces do live at the consumer
// PACKAGE — while the question the owner asked (2026-08-27,
// consent-webhook-trigger webhook_trigger_v2_ports.go: two interfaces
// consumed by exactly one struct) is about the file. A port file whose
// interfaces have exactly one consumer struct in the package (a field of the
// interface type, directly or through a pointer; different structs summed
// over all interfaces of the file) is a violation: declare the interface in
// the file of that one consumer. Two or more consumers is the convention the
// owner allowed — a dependencies.go shared by several entities. Zero
// consumers is not judged: the interface may be used only in function
// signatures or exist as the package's public contract. Generated files and
// _test.go files are not judged (GID-250); consumers are counted over named
// structs of non-generated, non-test files of the same package.
package ifaceplace

import (
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"slices"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const (
	ruleID                = "GID-134"
	inlineInterfaceRuleID = "GID-269"
	portFileRuleID        = "GID-271"
)

// layerSegments — path segments by which a package is recognized as the
// service's "own" layer (rather than stdlib/an external library).
var layerSegments = []string{
	"dal", "domain", "client", "server", "event", "app", "metric",
}

// Analyzer — the analyzer with default settings (no exclusions).
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — port files excluded from GID-271, by file name
	// ("dependencies.go" — the whole file) or by interface name ("Notifier"
	// — the interface is dropped from the file's judged set).
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-134/GID-269/GID-271 analyzer from the linter
// settings (.golangci.yml).
func NewAnalyzer(cfg Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidifaceplace",
		Doc: ruleID + ": interfaces live where they are used; " +
			"define the interface next to its consumer (exceptions: libraries, /domain/model for service/usecase, " +
			"and a _test.go helper's parameters/results dictated by a production constructor); " +
			inlineInterfaceRuleID + ": struct fields use named interfaces instead of inline declarations; " +
			portFileRuleID + ": a file of only interfaces with a single consumer struct in the package " +
			"moves its interfaces into the consumer's file",
		Requires: astwalk.Requires,
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, cfg)
		},
	}
}

func run(pass *analysis.Pass, cfg Settings) (any, error) {
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
	checkPortFiles(pass, cfg)
	return nil, nil
}

// portFile — a file whose top-level declarations are only interface types
// (imports are not declarations of substance). spec points at the first
// interface declaration — the diagnostic position.
type portFile struct {
	spec     *ast.TypeSpec
	fileName string
	ifaces   []*types.TypeName
}

// checkPortFiles implements GID-271: a port file whose interfaces are used by
// exactly one struct of the package is reported; two or more consumers — the
// dependencies.go convention — and zero consumers (the interface lives only
// in signatures or as a public contract) stay silent.
func checkPortFiles(pass *analysis.Pass, cfg Settings) {
	var (
		ports = map[*types.TypeName]*portFile{}
		order []*portFile
	)
	for _, file := range pass.Files {
		if ast.IsGenerated(file) || srcfile.IsTest(pass, file) {
			continue
		}
		pf := collectPortFile(pass, file, cfg.Exclude)
		if pf == nil {
			continue
		}
		order = append(order, pf)
		for _, obj := range pf.ifaces {
			ports[obj] = pf
		}
	}
	if len(order) == 0 {
		return
	}

	// Consumers: named structs of the package with a field typed by one of
	// the port interfaces (directly or through a pointer). Test files are
	// skipped: a double or a fixture picking up the interface is test
	// composition, not the production consumer the rule judges.
	consumers := make(map[*portFile]map[*ast.TypeSpec]string, len(order))
	skip := func(file *ast.File) bool {
		return ast.IsGenerated(file) || srcfile.IsTest(pass, file)
	}
	astwalk.NodesOf(pass, skip, func(file *ast.File, ts *ast.TypeSpec) {
		st, ok := ts.Type.(*ast.StructType)
		if !ok || st.Fields == nil {
			return
		}
		for _, field := range st.Fields.List {
			obj := namedInterfaceType(pass, field.Type)
			if obj == nil {
				continue
			}
			pf, ok := ports[obj]
			if !ok {
				continue
			}
			if consumers[pf] == nil {
				consumers[pf] = make(map[*ast.TypeSpec]string)
			}
			consumers[pf][ts] = fileNameOf(pass, file)
		}
	})

	for _, pf := range order {
		set := consumers[pf]
		if len(set) != 1 {
			continue
		}
		var (
			consumer     *ast.TypeSpec
			consumerFile string
		)
		for ts, fileName := range set {
			consumer = ts
			consumerFile = fileName
		}
		pass.Reportf(pf.spec.Pos(),
			"%s: the file declares only interfaces, and exactly one struct (%s) in the package uses them. "+
				"Fix: move the interface declaration to the file of its consumer (%s)",
			portFileRuleID, consumer.Name, consumerFile)
	}
}

// collectPortFile returns the port-file description for file, or nil when the
// file is not a port file (it declares anything besides interfaces) or none
// of its interfaces survive the exclusions.
func collectPortFile(pass *analysis.Pass, file *ast.File, excludeList []string) *portFile {
	fileName := fileNameOf(pass, file)
	if slices.Contains(excludeList, fileName) {
		return nil
	}
	var (
		ifaces []*types.TypeName
		first  *ast.TypeSpec
	)
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			return nil // a function or method — the file is not interface-only
		}
		if gd.Tok == token.IMPORT {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				return nil
			}
			if _, ok := ts.Type.(*ast.InterfaceType); !ok {
				return nil // a struct, a named type, an alias — not a port file
			}
			obj, isTypeName := pass.TypesInfo.Defs[ts.Name].(*types.TypeName)
			if !isTypeName || slices.Contains(excludeList, obj.Name()) {
				continue
			}
			if ifaces == nil {
				first = ts
			}
			ifaces = append(ifaces, obj)
		}
	}
	if len(ifaces) == 0 {
		return nil
	}
	return &portFile{spec: first, fileName: fileName, ifaces: ifaces}
}

// fileNameOf — the base name of the source file a node came from.
func fileNameOf(pass *analysis.Pass, file *ast.File) string {
	tokenFile := pass.Fset.File(file.Pos())
	if tokenFile == nil {
		return ""
	}
	return filepath.Base(tokenFile.Name())
}

// namedInterfaceType resolves the type expression of a struct field to the
// object of a named interface type, through a pointer if needed. Anything
// else (a basic type, an anonymous interface, error) returns nil.
func namedInterfaceType(pass *analysis.Pass, expr ast.Expr) *types.TypeName {
	tv, ok := pass.TypesInfo.Types[expr]
	if !ok {
		return nil
	}
	t := tv.Type
	if pt, ok := t.(*types.Pointer); ok {
		t = pt.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return nil
	}
	if _, isIface := named.Underlying().(*types.Interface); !isIface {
		return nil
	}
	return named.Obj()
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
