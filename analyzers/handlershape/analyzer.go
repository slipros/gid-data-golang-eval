// Package handlershape implements rule GID-230 (grpc-handler-shape):
// the shape of a gRPC/HTTP handler and of the gRPC service struct in the
// transport layer (server.md).
//
// Scope (via pathseg, never strings.Contains):
//   - handler packages: import path contains segment "server" AND segment
//     "grpc" AND ends with segment "handler"
//     (internal/server/grpc/<svc>/handler). Subpackages handler/convert and
//     handler/validate are out of scope. HTTP handler packages
//     (internal/server/http/handler) are NOT in scope: their handlers follow
//     the data-response.go shape (Handle(*http.Request, *dataresponse.Factory)
//     *response.DataResponse) and are governed by GID-162/GID-163, not the
//     gRPC handler shape enforced here.
//   - service packages: any other package under segment "server" — only the
//     gRPC service struct check applies there.
//
// Checks in a handler package, for every EXPORTED struct type (names with
// the Options suffix are skipped):
//  1. The struct must have a Handle method (pointer- or value-receiver)
//     whose first parameter is context.Context and whose last result is
//     error. The request/response types are not constrained: proto types
//     differ per RPC and Handle(ctx) error is accepted (boundary).
//  2. Every named, non-embedded field with a NAMED interface type must use
//     an interface called <StructName>Validator or <StructName>Service —
//     server.md: a handler depends on exactly these two interfaces.
//     Anonymous interface literals and embedded interfaces are skipped to
//     avoid false positives (see handlershape.feature, non-applicability).
//
// Check in a service package: a struct embedding a type named
// Unimplemented*Server is the gRPC service struct; all its named fields
// must be exported and carry the Handler suffix (handlers are injected as
// public fields via the composition root). Other embedded fields are
// skipped.
//
// Generated code (ast.IsGenerated) is skipped. LoadMode — TypesInfo:
// go/types is needed to recognize context.Context, error, method sets and
// interface-typed fields.
//
// Exclusions:
//   - inline: //nolint:gidhandlershape
//   - centralized: settings.exclude in .golangci.yml — struct type names
//     that are neither handlers nor gRPC service structs (e.g. "HealthCheck").
//
// Source: server.md.
package handlershape

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-230"

// Analyzer — variant with default settings (no exclusions).
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — struct type names that are not treated as handlers or
	// gRPC service structs and are not checked.
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-230 analyzer from linter settings (.golangci.yml).
func NewAnalyzer(cfg Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidhandlershape",
		Doc:  ruleID + ": a handler has Handle(ctx context.Context, req *T) (*R, error) and depends on <Handler>Validator/<Handler>Service; the gRPC service struct exposes handlers as exported *Handler fields. Fix: follow the server.md handler template",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, cfg)
		},
	}
}

func run(pass *analysis.Pass, cfg Settings) (any, error) {
	pkgPath := pass.Pkg.Path()
	// Scope: only the transport layer (segment "server").
	if !pathseg.Contains(pkgPath, "server") {
		return nil, nil
	}
	// gRPC handler packages end with the "handler" segment AND live under a
	// "grpc" segment (internal/server/grpc/<svc>/handler). handler/convert and
	// handler/validate do not end with "handler" and are out of scope. HTTP
	// handler packages (internal/server/http/handler) are intentionally
	// excluded: their handlers follow the data-response.go shape
	// Handle(*http.Request, *dataresponse.Factory) *response.DataResponse and
	// are governed by GID-162/GID-163, not the gRPC handler shape enforced here.
	isHandlerPkg := pathseg.EndsWith(pkgPath, "handler") && pathseg.Contains(pkgPath, "grpc")
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
				// Only exported struct types; *Options are settings types.
				if !ts.Name.IsExported() || strings.HasSuffix(ts.Name.Name, "Options") {
					continue
				}
				// Centralized exclusions by type name.
				if exclude.Match(cfg.Exclude, ts.Name.Name, ts.Name.Name) {
					continue
				}
				switch {
				case isHandlerPkg:
					checkHandler(pass, ts, st)
				case embedsUnimplemented(st):
					checkServiceStruct(pass, ts, st)
				}
			}
		}
	}
	return nil, nil
}

// checkHandler enforces the handler shape: a Handle method with
// context.Context first / error last, and dependency interfaces named
// <StructName>Validator / <StructName>Service.
func checkHandler(pass *analysis.Pass, ts *ast.TypeSpec, st *ast.StructType) {
	if !hasHandle(pass, ts) {
		pass.Reportf(ts.Name.Pos(),
			"%s: handler %q must have a Handle method with context.Context as the first param and error as the last result. Fix: func (h *%s) Handle(ctx context.Context, req *rpc.Request) (*rpc.Response, error)",
			ruleID, ts.Name.Name, ts.Name.Name)
	}
	wantValidator := ts.Name.Name + "Validator"
	wantService := ts.Name.Name + "Service"
	for _, field := range st.Fields.List {
		// Embedded fields (interfaces or structs) are skipped.
		if len(field.Names) == 0 {
			continue
		}
		named := namedInterface(pass.TypesInfo.TypeOf(field.Type))
		if named == nil {
			continue
		}
		obj := named.Obj()
		got := obj.Name()
		if got == wantValidator || got == wantService {
			continue
		}
		for _, name := range field.Names {
			pass.Reportf(name.Pos(),
				"%s: handler %q interface dependency %q must be named %q or %q. Fix: type %s interface{ Validate(ctx context.Context, req *rpc.Request) error }",
				ruleID, ts.Name.Name, got, wantValidator, wantService, wantValidator)
		}
	}
}

// checkServiceStruct enforces the gRPC service struct shape: handlers are
// exported fields with the Handler suffix.
func checkServiceStruct(pass *analysis.Pass, ts *ast.TypeSpec, st *ast.StructType) {
	for _, field := range st.Fields.List {
		// Embedded fields (Unimplemented*Server and friends) are skipped.
		if len(field.Names) == 0 {
			continue
		}
		for _, name := range field.Names {
			if name.IsExported() && strings.HasSuffix(name.Name, "Handler") {
				continue
			}
			pass.Reportf(name.Pos(),
				"%s: gRPC service %q must expose handlers as exported fields with the Handler suffix, got %q. Fix: DocumentsHandler *handler.Documents",
				ruleID, ts.Name.Name, name.Name)
		}
	}
}

// embedsUnimplemented reports whether the struct embeds a type named
// Unimplemented*Server (the protobuf forward-compatibility server).
func embedsUnimplemented(st *ast.StructType) bool {
	for _, field := range st.Fields.List {
		if len(field.Names) != 0 {
			continue
		}
		name := embeddedTypeName(field.Type)
		if strings.HasPrefix(name, "Unimplemented") && strings.HasSuffix(name, "Server") {
			return true
		}
	}
	return false
}

// embeddedTypeName extracts the bare type name of an embedded field:
// pb.UnimplementedFooServer, *pb.UnimplementedFooServer or a local ident.
func embeddedTypeName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.StarExpr:
		return embeddedTypeName(e.X)
	case *ast.SelectorExpr:
		return e.Sel.Name
	case *ast.Ident:
		return e.Name
	}
	return ""
}

// hasHandle reports whether the type (or a pointer to it) has a Handle
// method with context.Context as the first parameter and error as the last
// result.
func hasHandle(pass *analysis.Pass, ts *ast.TypeSpec) bool {
	obj, ok := pass.TypesInfo.Defs[ts.Name].(*types.TypeName)
	if !ok {
		return false
	}
	named, ok := obj.Type().(*types.Named)
	if !ok {
		return false
	}
	// Look on both T and *T: pointer-receiver methods are not in T's set.
	mset := types.NewMethodSet(types.NewPointer(named))
	for i := 0; i < mset.Len(); i++ {
		sel := mset.At(i)
		fn, ok := sel.Obj().(*types.Func)
		if !ok || fn.Name() != "Handle" {
			continue
		}
		sig, ok := fn.Type().(*types.Signature)
		if !ok {
			continue
		}
		return handleShape(sig)
	}
	return false
}

// handleShape: first param context.Context, last result error. Request and
// response types are intentionally unconstrained (proto types differ per RPC).
func handleShape(sig *types.Signature) bool {
	params := sig.Params()
	if params.Len() < 1 {
		return false
	}
	firstParam := params.At(0)
	if !isContext(firstParam.Type()) {
		return false
	}
	results := sig.Results()
	if results.Len() < 1 {
		return false
	}
	lastResult := results.At(results.Len() - 1)
	return isError(lastResult.Type())
}

// namedInterface returns the named type if t is a named interface type,
// nil otherwise (anonymous interface literals are skipped on purpose).
func namedInterface(t types.Type) *types.Named {
	named, ok := t.(*types.Named)
	if !ok {
		return nil
	}
	if _, ok := named.Underlying().(*types.Interface); !ok {
		return nil
	}
	return named
}

func isContext(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	return pkg != nil && pkg.Path() == "context" && obj.Name() == "Context"
}

func isError(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj.Pkg() == nil && obj.Name() == "error"
}
