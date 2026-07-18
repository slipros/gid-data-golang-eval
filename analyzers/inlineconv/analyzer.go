// Package inlineconv implements the GID-215 rule (no-inline-entity-literal):
// model ↔ entity conversion lives only in convert packages.
//
// Source: service.md "Conversion is always performed via the convert package".
//
// Scope: domain-layer packages (pathseg.HasLayer(pkgPath, "domain")), EXCEPT
// leaf packages named convert (pathseg.EndsWith(pkgPath, "convert") — that is
// where conversion should live).
//
// What is forbidden: a composite literal with ≥1 element whose named type
// (struct or named slice) is declared in an entity-layer package
// (pathseg.HasLayer(type's package, "dal", "entity") — including the
// filter/enum subpackages). Inline-filling an entity outside a convert package
// means conversion is smeared across the domain layer.
//
// What is NOT forbidden:
//   - an empty literal (entity.Snapshot{} — zero value);
//   - a model-type literal (model in domain is normal);
//   - an entity literal inside the service's convert package.
//
// Only the outermost entity literal is flagged: literals nested inside an
// already-flagged one are not reported again. Maps/slices of entity types
// (map[K]entity.X, []entity.X) are not flagged on their own — what gets flagged
// is the literal of the named entity type among their elements (it will be the
// outermost one).
//
// LoadMode — TypesInfo (types are needed to determine the named type's package).
// _test.go and generated files (ast.IsGenerated) are skipped.
// Targeted disabling: //nolint:gidinlineconv.
package inlineconv

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-215"

// Analyzer — the GID-215 rule: inline-filling an entity type in the domain
// layer is forbidden; conversion lives in a convert package.
var Analyzer = &analysis.Analyzer{
	Name: "gidinlineconv",
	Doc:  ruleID + ": inline-filling an entity type in the domain layer is forbidden; conversion lives in a convert package. Fix: move it to a convert function",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()

	// Rule zone: the domain layer (anchored to the module root — a package
	// nested under a different layer, e.g. .../server/grpc/domain/..., is NOT
	// the domain layer), but not convert packages (a leaf package, matched by
	// its own name — that is where conversion lives).
	if !pathseg.HasLayer(pkgPath, "domain") || pathseg.EndsWith(pkgPath, "convert") {
		return nil, nil
	}

	for _, file := range pass.Files {
		tokenFile := pass.Fset.File(file.Pos())
		if ast.IsGenerated(file) || strings.HasSuffix(tokenFile.Name(), "_test.go") {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}
			name, isEntity := entityLitName(pass, lit)
			if !isEntity {
				return true // descend inside — the violation may be deeper.
			}
			pass.Reportf(lit.Pos(),
				"%s: inline-filling the entity type %s in the domain layer is forbidden. "+
					"Fix: put conversion in a convert package (<Dst><Type>From<Src>)",
				ruleID, name)
			// The outermost entity literal is flagged — do not descend,
			// to avoid reporting nested entity literals again.
			return false
		})
	}
	return nil, nil
}

// entityLitName reports whether lit is a non-empty composite literal of a
// named entity type (struct or named slice from /dal/entity), and returns its
// display name (pkg.Type).
func entityLitName(pass *analysis.Pass, lit *ast.CompositeLit) (string, bool) {
	if len(lit.Elts) == 0 {
		return "", false // empty literal — zero value, allowed.
	}
	t := pass.TypesInfo.TypeOf(lit)
	if t == nil {
		return "", false
	}
	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return "", false // anonymous struct/slice, map[...], []... — not a named type.
	}
	// Only a struct or a named slice — maps/arrays as the literal's own type
	// are not counted (their named elements are handled by separate literals).
	switch named.Underlying().(type) {
	case *types.Struct, *types.Slice:
	default:
		return "", false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	// The entity layer is anchored to the module root: a type declared in a
	// package nested under a different layer (e.g. .../server/api/dal/entity)
	// is NOT the module's entity layer, so pathseg.HasLayer (not Contains) is
	// used here — it also covers the filter/enum subpackages (a trailing
	// suffix after the matched dal/entity prefix).
	if pkg == nil || !pathseg.HasLayer(pkg.Path(), "dal", "entity") {
		return "", false
	}
	return pkg.Name() + "." + obj.Name(), true
}
