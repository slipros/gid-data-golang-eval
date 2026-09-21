// Package protoconvplace implements rule GID-277: substantial conversion between
// transport messages and domain models lives in a leaf convert package.
//
// GID-215 already catches a domain package that directly fills an entity
// literal. This rule closes the complementary adapter gap: a gRPC handler or
// event adapter can otherwise hide a protobuf/model converter in an ordinary
// helper. GID-224 separately forbids transport and event code from importing
// DAL entities at all.
//
// A function is judged when all of the following hold:
//   - it is in /server/grpc/service/handler or /event, but not in a leaf convert
//     package;
//   - its parameters and results cross the domain-model and generated-protobuf
//     representation families;
//   - its body constructs a result-family composite literal with at least two
//     elements. The size threshold leaves one-field transport status/outcome
//     wrappers in handlers.
//
// Generated files and _test.go files are skipped. The gidprotoconvplace identifier
// supports targeted suppression.
package protoconvplace

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const (
	ruleID             = "GID-277"
	familyNone  family = 0
	familyModel family = 1 << iota
	familyProto
)

// Analyzer — rule GID-277: substantial cross-representation conversion lives
// in a leaf convert package.
var Analyzer = &analysis.Analyzer{
	Name:     "gidprotoconvplace",
	Doc:      ruleID + ": substantial protobuf/model conversion belongs in a leaf convert package",
	Requires: astwalk.Requires,
	Run:      run,
}

type family uint8

func run(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()
	if pathseg.EndsWith(pkgPath, "convert") || !judgedLayer(pkgPath) {
		return nil, nil
	}

	astwalk.NodesOf(pass, func(file *ast.File) bool {
		return ast.IsGenerated(file) || srcfile.IsTest(pass, file)
	}, func(_ *ast.File, fn *ast.FuncDecl) {
		if fn.Body == nil {
			return
		}

		from := fieldFamilies(pass, fn.Type.Params)
		to := fieldFamilies(pass, fn.Type.Results)
		if !crossesRepresentations(from, to) || !buildsSubstantialResult(pass, fn.Body, to) {
			return
		}

		pass.Reportf(fn.Name.Pos(),
			"%s: function %q performs substantial cross-representation conversion outside a convert package. "+
				"Fix: move the field mapping to a leaf convert package and call it from %q",
			ruleID, fn.Name.Name, pass.Pkg.Name())
	})

	return nil, nil
}

func judgedLayer(pkgPath string) bool {
	isGRPCHandler := pathseg.HasLayer(pkgPath, "server", "grpc", "service", "handler") &&
		pathseg.EndsWith(pkgPath, "handler")

	return isGRPCHandler || pathseg.HasLayer(pkgPath, "event")
}

func fieldFamilies(pass *analysis.Pass, fields *ast.FieldList) family {
	if fields == nil {
		return familyNone
	}

	var out family
	for _, field := range fields.List {
		out |= typeFamily(pass.TypesInfo.TypeOf(field.Type))
	}

	return out
}

func typeFamily(t types.Type) family {
	if t == nil {
		return familyNone
	}

	switch typed := types.Unalias(t).(type) {
	case *types.Pointer:
		return typeFamily(typed.Elem())
	case *types.Slice:
		return typeFamily(typed.Elem())
	case *types.Array:
		return typeFamily(typed.Elem())
	case *types.Map:
		return typeFamily(typed.Elem())
	case *types.Named:
		obj := typed.Obj()
		if obj.Pkg() == nil {
			return familyNone
		}
		pkg := obj.Pkg()
		pkgPath := pkg.Path()
		switch {
		case pathseg.HasLayer(pkgPath, "domain", "model"):
			return familyModel
		case isProtoMessage(typed):
			return familyProto
		}
	}

	return familyNone
}

func isProtoMessage(named *types.Named) bool {
	methods := types.NewMethodSet(types.NewPointer(named))

	return methods.Lookup(nil, "ProtoReflect") != nil
}

func crossesRepresentations(from, to family) bool {
	if from == familyNone || to == familyNone {
		return false
	}

	return from&^to != 0 || to&^from != 0
}

func buildsSubstantialResult(pass *analysis.Pass, body *ast.BlockStmt, results family) bool {
	found := false
	ast.Inspect(body, func(node ast.Node) bool {
		if found {
			return false
		}
		lit, ok := node.(*ast.CompositeLit)
		if !ok || len(lit.Elts) < 2 {
			return true
		}
		if typeFamily(pass.TypesInfo.TypeOf(lit))&results == 0 {
			return true
		}
		found = true

		return false
	})

	return found
}
