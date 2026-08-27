// Package modelplace implements rule GID-270 (model-place): data types are
// declared only in /domain/model.
//
// Owner requirement 2026-08-27: «Не может convert возвращать какую-то бизнес
// сущность. Бизнес сущности объявляются только в пакете model» — «Все модели
// данных живут только в domain/model. Не могут они вот так просто
// объявляться в service, usecase, convert». The incident shape:
// consent-webhook-trigger declared the WebhookTriggerV2Build result struct in
// a convert package and AltCraftTriggerV2Build right in usecase.
//
// Two parts, one linter:
//
//   - Part A: a package whose FINAL path segment is convert (the GID-247
//     boundary, pathseg.EndsWith) declares no types at all — a converter
//     transforms foreign types and has no type vocabulary of its own. Only
//     struct declarations are judged: an interface or a named basic type in
//     convert is outside the rule's territory. A convert package nested under
//     /domain (usecase/convert, repository/convert) is judged as convert and
//     is NOT additionally judged by part B — the output the message names is
//     the same /domain/model.
//
//   - Part B: in /domain/service and /domain/usecase (matched by path
//     segments — internal/domain/... and pkg/<module>/domain/... alike), an
//     exported struct is a data model when ALL of it holds:
//
//     the type has no methods in this package — a struct with behavior is
//     the layer's entity and sits where it belongs;
//     the type is not returned by any New/new-prefixed function of the
//     package (a pointer result counts) — what a constructor assembles is
//     the layer's entity itself: the service, the usecase;
//     the name does not end with an options suffix (settings.suffixes,
//     default Options/Option/Config/Params/Settings) — the options
//     convention (GID-126) keeps its settings types next to the entity.
//
//     A data structure without behavior is the model's cargo; its place is
//     /domain/model.
//
// Not judged: unexported types (a private key or a local pair is a package
// implementation detail that does not leak), _test.go files (a struct a test
// declares is its own scaffolding, GID-250), generated code.
//
// Exceptions: //nolint:gidmodelplace, or settings.exclude (the "Struct" form,
// as in GID-260).
package modelplace

import (
	"go/ast"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const ruleID = "GID-270"

// defaultSuffixes — name suffixes of the options convention (GID-126): a type
// named by one of them is settings, not a data model.
var defaultSuffixes = []string{"Options", "Option", "Config", "Params", "Settings"}

// Analyzer — GID-270 with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings — settings of rule GID-270 from .golangci.yml.
type Settings struct {
	// Suffixes — name suffixes of the options convention. Replaces the
	// defaults when set.
	Suffixes []string `json:"suffixes"`
	// Exclude — type names exempt from the rule (the "Struct" form).
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-270 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	suffixes := s.Suffixes
	if len(suffixes) == 0 {
		suffixes = defaultSuffixes
	}

	return &analysis.Analyzer{
		Name: "gidmodelplace",
		Doc: ruleID + ": data types live in /domain/model — a convert package declares no types, " +
			"and an exported struct with no methods and no constructor in /domain/service or /domain/usecase " +
			"is a data model. Fix: declare the type in /domain/model",
		Requires: astwalk.Requires,
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, suffixes, s.Exclude)
		},
	}
}

func run(pass *analysis.Pass, suffixes, excl []string) (any, error) {
	pkgPath := pass.Pkg.Path()
	isConvert := pathseg.EndsWith(pkgPath, "convert")
	layer := domainLayer(pkgPath)
	if !isConvert && layer == "" {
		return nil, nil
	}

	skip := func(file *ast.File) bool {
		return ast.IsGenerated(file) || srcfile.IsTest(pass, file)
	}

	typeSpecs, funcDecls := packageDecls(pass, skip)

	// Part B needs the whole package picture before judging a struct: a
	// constructor or a method may sit below the type in the file.
	var withMethods, builtByCtor map[string]struct{}
	if layer != "" {
		withMethods, builtByCtor = entityNames(pass, funcDecls)
	}

	for _, ts := range typeSpecs {
		checkType(pass, ts, layer, isConvert, suffixes, excl, withMethods, builtByCtor)
	}

	return nil, nil
}

// checkType judges one package-level type declaration against both parts.
func checkType(
	pass *analysis.Pass,
	ts *ast.TypeSpec,
	layer string,
	isConvert bool,
	suffixes, excl []string,
	withMethods, builtByCtor map[string]struct{},
) {
	if _, ok := ts.Type.(*ast.StructType); !ok {
		return // only struct declarations are judged
	}
	name := ts.Name.Name
	if !ts.Name.IsExported() || slices.Contains(excl, name) {
		return
	}

	if isConvert {
		pass.Reportf(ts.Name.Pos(),
			"%s: type %q is declared in a convert package — a converter transforms foreign types "+
				"and has no type vocabulary of its own. Fix: declare the type in /domain/model "+
				"(or //nolint:gidmodelplace when explicitly intended)",
			ruleID, name)
		return
	}

	if hasSuffix(name, suffixes) {
		return
	}
	if _, ok := withMethods[name]; ok {
		return
	}
	if _, ok := builtByCtor[name]; ok {
		return
	}

	pass.Reportf(ts.Name.Pos(),
		"%s: data struct %q is declared in /%s — it has no methods and is built by no constructor, "+
			"so it is a data model. Fix: move the type to /domain/model "+
			"(or //nolint:gidmodelplace when explicitly intended)",
		ruleID, name, layer)
}

// packageDecls collects the package-level type specs and function declarations
// of the package in one pruned walk: a function body holds no package-level
// declarations, so its subtree is skipped once the decl itself is recorded.
func packageDecls(pass *analysis.Pass, skip func(*ast.File) bool) ([]*ast.TypeSpec, []*ast.FuncDecl) {
	var typeSpecs []*ast.TypeSpec
	var funcDecls []*ast.FuncDecl

	astwalk.NodesPruning(pass,
		[]ast.Node{(*ast.TypeSpec)(nil), (*ast.FuncDecl)(nil)},
		skip,
		func(_ *ast.File, n ast.Node) bool {
			switch n := n.(type) {
			case *ast.TypeSpec:
				typeSpecs = append(typeSpecs, n)
			case *ast.FuncDecl:
				funcDecls = append(funcDecls, n)
				return false
			}

			return true
		})

	return typeSpecs, funcDecls
}

// entityNames returns the names of the package's own types that part B keeps
// where they are: withMethods — types with at least one method in the package,
// builtByCtor — types returned by at least one New/new-prefixed function.
func entityNames(pass *analysis.Pass, funcDecls []*ast.FuncDecl) (methods, built map[string]struct{}) {
	methods = map[string]struct{}{}
	built = map[string]struct{}{}

	for _, fn := range funcDecls {
		if fn.Recv != nil && len(fn.Recv.List) > 0 {
			if name, ok := localNamed(pass, fn.Recv.List[0].Type); ok {
				methods[name] = struct{}{}
			}

			continue
		}
		if !isCtorName(fn.Name.Name) || fn.Type.Results == nil {
			continue
		}
		for _, res := range fn.Type.Results.List {
			if name, ok := localNamed(pass, res.Type); ok {
				built[name] = struct{}{}
			}
		}
	}

	return methods, built
}

// isCtorName reports whether the function name carries the New/new prefix of
// the constructor convention (GID-104).
func isCtorName(name string) bool {
	return strings.HasPrefix(name, "New") || strings.HasPrefix(name, "new")
}

// localNamed resolves the type expression to the name of a type declared in
// this package — a constructor result is dereferenced through its pointer.
func localNamed(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	t := pass.TypesInfo.TypeOf(expr)
	if t == nil {
		return "", false
	}
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return "", false
	}
	obj := named.Obj()
	if obj == nil || obj.Pkg() != pass.Pkg {
		return "", false
	}

	return obj.Name(), true
}

// domainLayer returns "domain/service" or "domain/usecase" for a package of
// the corresponding domain layer, or "" outside of them.
func domainLayer(pkgPath string) string {
	for _, sub := range []string{"service", "usecase"} {
		if pathseg.HasLayer(pkgPath, "domain", sub) {
			return "domain/" + sub
		}
	}

	return ""
}

// hasSuffix reports whether name ends with one of the options-convention suffixes.
func hasSuffix(name string, suffixes []string) bool {
	for _, s := range suffixes {
		if s != "" && strings.HasSuffix(name, s) {
			return true
		}
	}

	return false
}
