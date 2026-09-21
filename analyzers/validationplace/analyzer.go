// Package validationplace implements rule GID-278: transport request and
// command validation belongs to the ingress validate package, not the domain.
//
// The rule deliberately recognises only the structural convention that can be
// enforced without guessing business semantics:
//   - a /domain/model type whose name ends in Request or Command, or any type
//     in /domain/model/request, must not own an error-returning Validate method;
//   - /domain/service and /domain/usecase must not call such a method on a
//     model type from their own module.
//
// A domain entity with a Validate method is not judged: its method may enforce
// a business invariant. ValidateState and methods that do not return error are
// also outside the transport-validator convention. Generated and test files
// are skipped.
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
	Doc: ruleID + ": transport Request and Command validation belongs to the ingress validate package, " +
		"not /domain/model, /domain/service or /domain/usecase",
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
		typeName, modelPkg, ok := transportValidationReceiverPackage(fn)
		if !ok || pathseg.ModuleRoot(modelPkg) != pathseg.ModuleRoot(pass.Pkg.Path()) {
			return
		}
		pass.Reportf(call.Pos(),
			"%s: /domain/%s calls %s.Validate; transport requests and commands must be validated at ingress. "+
				"Fix: validate before conversion in the HTTP, gRPC or Kafka validate package and remove the domain call",
			ruleID, layer, typeName)
	})
}

func transportValidationReceiver(fn *types.Func) (string, bool) {
	name, _, ok := transportValidationReceiverPackage(fn)
	return name, ok
}

func transportValidationReceiverPackage(fn *types.Func) (name, pkgPath string, ok bool) {
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
	if pkg == nil || !pathseg.HasLayer(pkg.Path(), "domain", "model") || !transportModelType(pkg.Path(), obj.Name()) {
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
