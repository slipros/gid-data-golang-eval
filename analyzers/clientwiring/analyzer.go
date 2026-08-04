// Package clientwiring implements rule GID-255: a package in /client/** that
// declares no client is wiring, not a layer — it belongs in the composition
// root.
//
// The case that produced the rule (ad-cabinet-connector, 2026-08-04):
// internal/client/resourceregistry holds exactly two functions —
// NewConnection, which assembles a gRPC connection out of options, a logger
// and metrics and returns a type of the grpc library, and NewLoggingDecider,
// its logging policy. There is no client type, no method, no model of its
// own, and the only consumer is internal/app/api. Assembling a transport
// object once at start-up out of options and dependencies IS the composition
// root's job; a package that does only that is the composition root spelled
// in the wrong directory.
//
// A package with a constructor but no method is also the reliable symptom of
// the layer being used wrongly: a real client (client.md) owns a type whose
// methods call the external API, owns its own models, and converts them —
// the consumer never sees the transport. When that type is missing, the
// domain ends up talking to the foreign service's DTOs directly, and the
// "client" contributes a directory and an import, nothing else. Such a
// package moves into /app/** as is.
//
// Detect (per package): the package lies under /client/** in a module laid
// out as a service, declares at least one function and NOT ONE method (a
// FuncDecl with a receiver). Reported once, on the package clause.
//
// Not flagged: a package with methods (a real client, however thin); a
// package of pure types/constants with no functions at all (client models);
// the leaf subpackages convert/dto/mock/mocks, which are legitimately
// function-only; generated code and _test.go, which are ignored when
// deciding. A flat library module is skipped entirely (internal/modlayout) —
// libs/grpc.git/client is a library's own client package, not a service layer.
package clientwiring

import (
	"go/ast"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/modlayout"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-255"

// skipLeaves — leaf package names under /client/** that are function-only by
// design: converters, DTO helpers and generated mocks.
var skipLeaves = []string{"convert", "dto", "mock", "mocks"}

// Analyzer — rule GID-255.
var Analyzer = NewAnalyzer()

// NewAnalyzer builds the GID-255 analyzer.
func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidclientwiring",
		Doc: ruleID + ": a package in /client/** with constructors but no method declares no client — it is " +
			"wiring. Fix: move it into the composition root /app/**",
		Run: run,
	}
}

func run(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()
	if !pathseg.HasLayer(pkgPath, "client") || isSkippedLeaf(pkgPath) {
		return nil, nil
	}
	if !modlayout.IsServiceModule(pass) {
		return nil, nil
	}

	hasFunc := false
	var reportAt *ast.File
	for _, file := range pass.Files {
		if ast.IsGenerated(file) || isTestFile(pass, file) {
			continue
		}
		if reportAt == nil {
			reportAt = file
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if fn.Recv != nil {
				// A method — the package owns a client type. Nothing to report.
				return nil, nil
			}
			hasFunc = true
		}
	}
	if !hasFunc || reportAt == nil {
		return nil, nil
	}

	pass.Reportf(reportAt.Package,
		"%s: package %q is in the client layer but declares no client — it has functions and not one method, "+
			"so nothing here calls an external API on a type of its own. A package that only assembles a "+
			"connection out of options, a logger and metrics is the composition root spelled in the wrong "+
			"directory, and it leaves the consumer talking to the foreign service's DTOs. "+
			"Fix: move it as is into /app/** (wiring), or give the layer a real client — a type whose methods "+
			"call the API and return the client's own models (client.md)",
		ruleID, pass.Pkg.Name())
	return nil, nil
}

// isSkippedLeaf reports whether the package's last path segment is one of the
// function-only leaves under /client/**.
func isSkippedLeaf(pkgPath string) bool {
	idx := strings.LastIndex(pkgPath, "/")
	return slices.Contains(skipLeaves, pkgPath[idx+1:])
}

// isTestFile reports whether the file is a _test.go file.
func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	return strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, "_test.go")
}
