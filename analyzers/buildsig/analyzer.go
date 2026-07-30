// Package buildsig implements rule GID-212 (build-signature): the contract
// of repository build functions.
//
// Source: repo.md.
//
// Checks:
//
//  1. Result signature. In /dal/repository/build/** packages, exported
//     functions (FuncDecl without a receiver) of non-test files must return
//     EITHER (string, []any, error) — a single query (sql, args, err), OR
//     (*<...>.Batch, error) — a batch operation (matched by the name of the
//     named type Batch, any package). Any other result signature → diagnostic.
//     Unexported helper functions of the build package are not checked, and
//     neither are _test.go files: a Test/Benchmark/Fuzz function or a test
//     builder is not a build function.
//
//     settings.allow-results extends the contract with additional result
//     signatures for builders that do not produce SQL (a search-engine DSL
//     builder has no args []any): e.g. ["(string, error)", "string",
//     "[]string", "(*omd.SearchEntitiesWithQueryParams, error)"]. Empty by
//     default — the strict contract above.
//
//  2. Ban on the squirrel import. Importing github.com/Masterminds/squirrel is
//     allowed only in /dal/repository/build/** packages (including their
//     external test package build_test). In any other package a squirrel
//     import → diagnostic, in _test.go files as well.
//
// Signatures are recognized structurally via go/types (LoadModeTypesInfo).
// Generated code is skipped.
package buildsig

import (
	"go/ast"
	"go/types"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-212"

// Analyzer — the variant with default settings (the strict contract).
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// AllowResults — additional allowed result signatures of a build function,
	// beyond the built-in (string, []any, error) and (*batch.Batch, error).
	// Written the way they are declared in Go: "(string, error)", "string",
	// "[]string", "(*omd.SearchEntitiesWithQueryParams, error)" — the package
	// is named by its name, not by its import path. Whitespace is ignored,
	// interface{} and any are the same thing. Empty — the strict contract.
	AllowResults []string `json:"allow-results"`
}

// NewAnalyzer builds the GID-212 analyzer with the given settings.
func NewAnalyzer(cfg Settings) *analysis.Analyzer {
	allowed := make(map[string]struct{}, len(cfg.AllowResults))
	for _, sig := range cfg.AllowResults {
		allowed[normalizeResults(sig)] = struct{}{}
	}
	return &analysis.Analyzer{
		Name: "gidbuildsig",
		Doc: ruleID + ": build functions return (string, []any, error) or (*batch.Batch, error) " +
			"(settings.allow-results extends the contract); squirrel only in /dal/repository/build",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, allowed)
		},
	}
}

func run(pass *analysis.Pass, allowed map[string]struct{}) (any, error) {
	// The external test package of a build package (build_test) is a build
	// package too as far as the squirrel ban goes.
	pkgPath := strings.TrimSuffix(pass.Pkg.Path(), "_test")
	inBuild := pathseg.HasLayer(pkgPath, "dal", "repository", "build")

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}

		// Check 2: the squirrel import is allowed only in build packages.
		if !inBuild {
			checkSquirrelImports(pass, file)
		}

		// Check 1: the result signature of exported build functions. _test.go
		// files are out of scope: a Test/Benchmark/Fuzz function or a test
		// builder is not a build function.
		if inBuild && !isTestFile(pass, file) {
			checkBuildSignatures(pass, file, allowed)
		}
	}
	return nil, nil
}

// isTestFile reports whether the file is a _test.go one.
func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	return strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, "_test.go")
}

// checkSquirrelImports flags a squirrel import outside a build package.
func checkSquirrelImports(pass *analysis.Pass, file *ast.File) {
	const (
		squirrelPkg = "github.com/Masterminds/squirrel"
		msgSquirrel = ruleID + ": squirrel is allowed only in repository build packages (/dal/repository/build). Fix: move squirrel usage into /dal/repository/build"
	)
	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		if path == squirrelPkg {
			pass.Reportf(imp.Pos(), msgSquirrel)
		}
	}
}

// checkBuildSignatures checks the result of exported functions without a receiver.
func checkBuildSignatures(pass *analysis.Pass, file *ast.File, allowed map[string]struct{}) {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		// Methods (with a receiver) and unexported helpers are not checked.
		if fn.Recv != nil || !fn.Name.IsExported() {
			continue
		}
		obj, ok := pass.TypesInfo.Defs[fn.Name].(*types.Func)
		if !ok {
			continue
		}
		sig, ok := obj.Type().(*types.Signature)
		if !ok {
			continue
		}
		if isSingleQuerySig(sig) || isBatchSig(sig) || isAllowedSig(sig, allowed) {
			continue
		}
		const msgSignature = ruleID +
			": a build function must return (sql string, args []any, err error) or (*batch.Batch, error). Fix: adjust the signature"
		pass.Reportf(fn.Name.Pos(), msgSignature)
	}
}

// isAllowedSig — the result matches one of the signatures from
// settings.allow-results. Packages of named types are compared by package name
// (omd.SearchEntitiesWithQueryParams), not by import path: that is how the
// signature is written in the source, and that is what the setting spells out.
func isAllowedSig(sig *types.Signature, allowed map[string]struct{}) bool {
	if len(allowed) == 0 {
		return false
	}
	_, ok := allowed[resultsString(sig)]
	return ok
}

// resultsString renders the function's result list in the normalized form used
// for comparison with settings.allow-results.
func resultsString(sig *types.Signature) string {
	res := sig.Results()
	parts := make([]string, 0, res.Len())
	for v := range res.Variables() {
		parts = append(parts, types.TypeString(v.Type(), pkgNameQualifier))
	}
	return normalizeResults("(" + strings.Join(parts, ",") + ")")
}

// pkgNameQualifier renders a named type's package by its name, not by its
// import path: gitlab.gid.team/…/omd.Params → omd.Params.
func pkgNameQualifier(pkg *types.Package) string {
	return pkg.Name()
}

// normalizeResults brings a result list to a comparable form: no whitespace, no
// outer parentheses, interface{} spelled as any. So "(string, error)",
// "( string,error )" and "string,error" are the same setting.
func normalizeResults(sig string) string {
	s := strings.Join(strings.Fields(sig), "")
	s = strings.TrimPrefix(s, "(")
	s = strings.TrimSuffix(s, ")")
	return strings.ReplaceAll(s, "interface{}", "any")
}

// isSingleQuerySig — result (string, []any, error).
func isSingleQuerySig(sig *types.Signature) bool {
	res := sig.Results()
	if res.Len() != 3 {
		return false
	}
	sqlRes, argsRes, errRes := res.At(0), res.At(1), res.At(2)
	if !isString(sqlRes.Type()) {
		return false
	}
	if !isSliceOfAny(argsRes.Type()) {
		return false
	}
	return isError(errRes.Type())
}

// isBatchSig — result (*<...>.Batch, error): a pointer to a named type
// with the name Batch (any package).
func isBatchSig(sig *types.Signature) bool {
	const batchType = "Batch"
	res := sig.Results()
	if res.Len() != 2 {
		return false
	}
	batchRes, errRes := res.At(0), res.At(1)
	if !isPtrToNamed(batchRes.Type(), batchType) {
		return false
	}
	return isError(errRes.Type())
}

func isString(t types.Type) bool {
	b, ok := t.Underlying().(*types.Basic)
	return ok && b.Kind() == types.String
}

// isSliceOfAny — []any (a slice with the empty interface as the element).
func isSliceOfAny(t types.Type) bool {
	sl, ok := t.Underlying().(*types.Slice)
	if !ok {
		return false
	}
	elem := sl.Elem()
	iface, ok := elem.Underlying().(*types.Interface)
	return ok && iface.NumMethods() == 0
}

func isError(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj != nil && obj.Pkg() == nil && obj.Name() == "error"
}

// isPtrToNamed — a pointer to a named type with the given name.
func isPtrToNamed(t types.Type, name string) bool {
	ptr, ok := t.(*types.Pointer)
	if !ok {
		return false
	}
	named, ok := ptr.Elem().(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj != nil && obj.Name() == name
}
