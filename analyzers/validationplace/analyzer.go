// Package validationplace implements rule GID-278: transport request and
// command validation belongs to the ingress validate package, not the domain.
//
// The rule deliberately recognises only structural boundaries:
//   - a /domain/model type whose name ends in Request or Command, or any type
//     in /domain/model/request, must not own an error-returning Validate method;
//   - /domain/service and /domain/usecase must not call an error-returning
//     Validate method on any model type from their own module.
//
// A domain entity may still declare Validate for a business invariant, but the
// domain execution layers cannot use that generic method as an input-validation
// boundary. ValidateState and methods that do not return error remain outside
// the convention. Generated and test files are skipped.
package validationplace

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const ruleID = "GID-278"

// Analyzer rejects transport-validation ownership in the domain layer.
var Analyzer = &analysis.Analyzer{
	Name: "gidvalidationplace",
	Doc: ruleID + ": transport validation belongs to the ingress validate package; " +
		"domain services and usecases do not validate domain models",
	Requires: astwalk.Requires,
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()
	isModel := pathseg.HasLayer(pkgPath, "domain", "model")
	layer := executionLayer(pkgPath)
	if !isModel && layer == "" {
		return nil, nil
	}

	if isModel {
		for _, file := range pass.Files {
			if ast.IsGenerated(file) || srcfile.IsTest(pass, file) {
				continue
			}
			checkModelMethods(pass, file)
		}
	}
	if layer != "" {
		checkDomainCalls(pass, layer)
	}

	return nil, nil
}

func checkModelMethods(pass *analysis.Pass, file *ast.File) {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || fn.Name.Name != "Validate" {
			continue
		}
		obj, ok := pass.TypesInfo.Defs[fn.Name].(*types.Func)
		if !ok {
			continue
		}
		typeName, ok := transportValidationReceiver(obj)
		if !ok {
			continue
		}
		pass.Reportf(fn.Name.Pos(),
			"%s: Validate on domain model type %s owns transport validation. "+
				"Fix: move validation to the ingress validate package and pass a validated model to the domain",
			ruleID, typeName)
	}
}

func checkDomainCalls(pass *analysis.Pass, layer string) {
	skip := func(file *ast.File) bool {
		return ast.IsGenerated(file) || srcfile.IsTest(pass, file)
	}
	astwalk.NodesOf[*ast.CallExpr](pass, skip, func(_ *ast.File, call *ast.CallExpr) {
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Validate" {
			return
		}
		selection := pass.TypesInfo.Selections[sel]
		if selection == nil {
			return
		}
		fn, ok := selection.Obj().(*types.Func)
		if !ok {
			return
		}
		typeName, modelPkg, ok := domainModelValidationReceiverPackage(fn)
		if !ok || pathseg.ModuleRoot(modelPkg) != pathseg.ModuleRoot(pass.Pkg.Path()) {
			return
		}
		pass.Reportf(call.Pos(),
			"%s: /domain/%s calls %s.Validate on a domain model; validation must run at ingress. "+
				"Fix: validate before conversion in the HTTP, gRPC or Kafka validate package and remove the domain call",
			ruleID, layer, typeName)
	})
}

func transportValidationReceiver(fn *types.Func) (string, bool) {
	name, _, ok := transportValidationReceiverPackage(fn)
	return name, ok
}

func transportValidationReceiverPackage(fn *types.Func) (name, pkgPath string, ok bool) {
	name, pkgPath, ok = domainModelValidationReceiverPackage(fn)
	if !ok || !transportModelType(pkgPath, name) {
		return "", "", false
	}
	return name, pkgPath, true
}

func domainModelValidationReceiverPackage(fn *types.Func) (name, pkgPath string, ok bool) {
	sig, ok := fn.Type().(*types.Signature)
	if !ok || sig.Recv() == nil || !hasErrorResult(sig) {
		return "", "", false
	}
	receiver := sig.Recv()
	named, ok := namedType(receiver.Type())
	if !ok {
		return "", "", false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	if pkg == nil || !pathseg.HasLayer(pkg.Path(), "domain", "model") {
		return "", "", false
	}
	return obj.Name(), pkg.Path(), true
}

func namedType(typ types.Type) (*types.Named, bool) {
	typ = types.Unalias(typ)
	if ptr, ok := typ.(*types.Pointer); ok {
		typ = types.Unalias(ptr.Elem())
	}
	named, ok := typ.(*types.Named)
	return named, ok
}

func hasErrorResult(sig *types.Signature) bool {
	results := sig.Results()
	errorObject := types.Universe.Lookup("error")
	for i := range results.Len() {
		result := results.At(i)
		if types.Identical(result.Type(), errorObject.Type()) {
			return true
		}
	}
	return false
}

func transportModelType(pkgPath, name string) bool {
	return pathseg.EndsWith(pkgPath, "request") ||
		strings.HasSuffix(name, "Request") ||
		strings.HasSuffix(name, "Command")
}

func executionLayer(pkgPath string) string {
	switch {
	case pathseg.HasLayer(pkgPath, "domain", "service"):
		return "service"
	case pathseg.HasLayer(pkgPath, "domain", "usecase"):
		return "usecase"
	default:
		return ""
	}
}
