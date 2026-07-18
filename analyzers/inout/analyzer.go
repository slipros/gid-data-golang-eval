// Package inout implements rule GID-111: input data (model/entity
// structs) is passed by pointer, output is returned by value.
// Applies on the repo, service, usecase and handler layers.
//
// In /client/** packages the client has no domain/model or dal/entity of its
// own, so the same rule applies to the client's own same-module named
// structs (its request/response types) instead — see clientStructName.
//
// Exceptions:
//   - pointwise: //nolint:gidinout
//   - centrally: settings.exclude in .golangci.yml —
//     entries of the form "Method" or "Type.Method".
package inout

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-111"

// scopes — layers (anchored to the module root) whose exported methods carry
// domain/model or dal/entity data. Handler packages are handled separately in
// inScope: they are leaves, not a module-root layer.
var scopes = [][]string{
	{"dal", "repository"},
	{"domain", "service"},
	{"domain", "usecase"},
}

// clientScope — the client layer tree. Unlike the scopes above, its
// input/output data is not domain/model or dal/entity but the client's own
// same-module named structs (see clientStructName).
var clientScope = []string{"client"}

// layerTypeTrees — package trees whose structs are treated as layer data
// for the repo/service/usecase/handler scopes.
var layerTypeTrees = [][]string{
	{"domain", "model"},
	{"dal", "entity"},
}

// Analyzer is the variant with default settings (no exceptions).
var Analyzer = NewAnalyzer(Settings{})

// Settings holds the linter settings from .golangci.yml.
type Settings struct {
	// Exclude — methods to exempt: "Method" or "Type.Method".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-111 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidinout",
		Doc:  ruleID + ": input model/entity (or, in /client/**, the client's own) structs by pointer, output by value. Fix: take *T for input, return T for output",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s)
		},
	}
}

func run(pass *analysis.Pass, s Settings) (any, error) {
	pkgPath := pass.Pkg.Path()
	isClient := pathseg.HasLayer(pkgPath, clientScope...)
	if !isClient && !inScope(pkgPath) {
		return nil, nil
	}
	structName := layerStructName
	if isClient {
		structName = clientStructName(pkgPath)
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || !fn.Name.IsExported() {
				continue
			}
			if exclude.Match(s.Exclude, recvTypeName(fn), fn.Name.Name) {
				continue
			}
			checkSignature(pass, fn, structName)
		}
	}
	return nil, nil
}

func checkSignature(pass *analysis.Pass, fn *ast.FuncDecl, structName func(types.Type) (string, bool)) {
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			if name, ok := structName(pass.TypesInfo.TypeOf(field.Type)); ok {
				pass.Reportf(field.Pos(),
					"%s: input data must be passed by pointer. Fix: use *%s", ruleID, name)
			}
		}
	}
	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			t := pass.TypesInfo.TypeOf(field.Type)
			ptr, ok := t.(*types.Pointer)
			if !ok {
				continue
			}
			if name, ok := structName(ptr.Elem()); ok {
				pass.Reportf(field.Pos(),
					"%s: output data must be returned by value. Fix: use %s", ruleID, name)
			}
		}
	}
}

// layerStructName returns the type name if it is a struct from
// /domain/model or /dal/entity (by value, without a pointer).
func layerStructName(t types.Type) (string, bool) {
	named, ok := t.(*types.Named)
	if !ok {
		return "", false
	}
	if _, ok := named.Underlying().(*types.Struct); !ok {
		return "", false
	}
	obj := named.Obj()
	if obj.Pkg() == nil {
		return "", false
	}
	pkg := obj.Pkg()
	for _, tree := range layerTypeTrees {
		if pathseg.HasLayer(pkg.Path(), tree...) {
			return pkg.Name() + "." + obj.Name(), true
		}
	}
	return "", false
}

// clientStructName builds a struct-name check for the client scope: a
// client has no dedicated model/entity tree, so any same-module named
// struct (the client's own request/response type, wherever it is declared)
// counts as its input/output data. A struct from a foreign module
// (a third-party or generated dependency, e.g. a protobuf stub) does not.
// Same-module membership is decided by the module root (pathseg.ModuleRoot),
// not by the first path segment — for real import paths every
// github.com/<org>/<repo> package shares the segment "github.com".
func clientStructName(pkgPath string) func(types.Type) (string, bool) {
	modRoot := pathseg.ModuleRoot(pkgPath)
	return func(t types.Type) (string, bool) {
		named, ok := t.(*types.Named)
		if !ok {
			return "", false
		}
		if _, ok := named.Underlying().(*types.Struct); !ok {
			return "", false
		}
		obj := named.Obj()
		if obj.Pkg() == nil {
			return "", false
		}
		pkg := obj.Pkg()
		if pathseg.ModuleRoot(pkg.Path()) != modRoot {
			return "", false
		}
		return pkg.Name() + "." + obj.Name(), true
	}
}

func inScope(pkgPath string) bool {
	// handler packages are leaves (server/grpc/handler, server/http/handler),
	// matched by the trailing segment rather than anchored to the module root.
	if pathseg.EndsWith(pkgPath, "handler") {
		return true
	}
	for _, scope := range scopes {
		if pathseg.HasLayer(pkgPath, scope...) {
			return true
		}
	}
	return false
}

func recvTypeName(fn *ast.FuncDecl) string {
	t := fn.Recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	if ident, ok := t.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}
