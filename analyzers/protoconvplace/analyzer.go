// Package protoconvplace implements rule GID-277: substantial conversion between
// protobuf messages and domain or DAL representations lives in its boundary-owned
// convert package.
//
// GID-215 already catches a domain package that directly fills an entity
// literal. This rule closes the complementary adapter gap: a gRPC handler,
// Kafka adapter, domain service, or DAL repository wrapping a gRPC client can
// otherwise hide a protobuf converter in an ordinary helper. GID-224 separately forbids
// transport and event code from importing DAL entities at all.
//
// A function is judged when all of the following hold:
//   - it is under /server/grpc/service, /event, /domain/service, or
//     /dal/repository;
//   - it is outside the permitted conversion package: inbound gRPC conversion
//     belongs in /server/grpc/service/handler/convert, Kafka conversion belongs
//     in the matching /event/kafka/{producer,consumer}/convert package,
//     outbound domain gRPC-client conversion belongs in /domain/service/convert,
//     and protobuf/entity conversion belongs in /dal/repository/convert;
//   - its parameters and results cross generated-protobuf and domain-model or
//     DAL-entity representation families;
//   - its body constructs a result-family struct literal with at least two
//     elements. The size threshold leaves one-field transport status/outcome
//     wrappers in handlers; a slice or map literal only collects values and is
//     not a field mapping.
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

const ruleID = "GID-277"

const (
	familyNone  family = 0
	familyModel family = 1 << (iota - 1)
	familyEntity
	familyProto
)

// Analyzer — rule GID-277: substantial cross-representation conversion lives
// in its boundary-owned convert package.
var Analyzer = &analysis.Analyzer{
	Name:     "gidprotoconvplace",
	Doc:      ruleID + ": substantial protobuf/model conversion belongs in its boundary-owned convert package",
	Requires: astwalk.Requires,
	Run:      run,
}

type family uint8

func run(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()
	if !judgedLayer(pkgPath) || allowedConversionPackage(pkgPath) {
		return nil, nil
	}

	destination := conversionDestination(pkgPath)

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
			"%s: function %q performs substantial cross-representation conversion outside %s. "+
				"Fix: move the field mapping to %s and call it from %q",
			ruleID, fn.Name.Name, destination, destination, pass.Pkg.Name())
	})

	return nil, nil
}

func judgedLayer(pkgPath string) bool {
	return isGRPCService(pkgPath) || isDomainService(pkgPath) || isDALRepository(pkgPath) ||
		pathseg.HasLayer(pkgPath, "event")
}

func allowedConversionPackage(pkgPath string) bool {
	switch {
	case isGRPCService(pkgPath):
		return exactLayerPackage(pkgPath, "server", "grpc", "service", "handler", "convert")
	case isDomainService(pkgPath):
		return exactLayerPackage(pkgPath, "domain", "service", "convert")
	case isDALRepository(pkgPath):
		return exactLayerPackage(pkgPath, "dal", "repository", "convert")
	default:
		return exactLayerPackage(pkgPath, "event", "kafka", "producer", "convert") ||
			exactLayerPackage(pkgPath, "event", "kafka", "consumer", "convert")
	}
}

func exactLayerPackage(pkgPath string, segments ...string) bool {
	return len(pathseg.LayerSegments(pkgPath)) == len(segments) &&
		pathseg.HasLayer(pkgPath, segments...)
}

func conversionDestination(pkgPath string) string {
	if isGRPCService(pkgPath) {
		return "/server/grpc/service/handler/convert"
	}
	if isDomainService(pkgPath) {
		return "/domain/service/convert"
	}
	if isDALRepository(pkgPath) {
		return "/dal/repository/convert"
	}

	switch {
	case pathseg.HasLayer(pkgPath, "event", "kafka", "producer"):
		return "/event/kafka/producer/convert"
	case pathseg.HasLayer(pkgPath, "event", "kafka", "consumer"):
		return "/event/kafka/consumer/convert"
	default:
		return "/event/kafka/producer/convert or /event/kafka/consumer/convert"
	}
}

func isGRPCService(pkgPath string) bool {
	return pathseg.HasLayer(pkgPath, "server", "grpc", "service")
}

func isDomainService(pkgPath string) bool {
	return pathseg.HasLayer(pkgPath, "domain", "service")
}

func isDALRepository(pkgPath string) bool {
	return pathseg.HasLayer(pkgPath, "dal", "repository")
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
		case pathseg.HasLayer(pkgPath, "dal", "entity"):
			return familyEntity
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
	const nonProto = familyModel | familyEntity

	return from&familyProto != 0 && to&nonProto != 0 ||
		to&familyProto != 0 && from&nonProto != 0
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
		litType := pass.TypesInfo.TypeOf(lit)
		if !isStructLiteral(litType) || typeFamily(litType)&results == 0 {
			return true
		}
		found = true

		return false
	})

	return found
}

// isStructLiteral reports whether a composite literal of type t builds a
// struct — that is a field mapping. A slice or map literal only collects
// values and maps nothing, however many elements it lists. An elided literal
// inside []*T{{…}} is recorded with the pointer type, so it is dereferenced.
func isStructLiteral(t types.Type) bool {
	if t == nil {
		return false
	}
	if ptr, ok := types.Unalias(t).(*types.Pointer); ok {
		t = ptr.Elem()
	}
	_, ok := t.Underlying().(*types.Struct)

	return ok
}
