// Package serviceorch implements rule GID-260: a domain service does not
// orchestrate — it is one entity and the repository of that entity. Composing
// several data sources, and holding them together in a transaction, is what a
// usecase is for.
//
// Two markers, both read off the service struct in /domain/service:
//
//  1. more than one repository field (a named type whose name ends with a
//     configured suffix, settings.suffixes, default "Repository") — two sources
//     in one service is composition;
//  2. a transaction field — a func value whose last parameter is itself a func
//     returning error and which returns error (model.InTransactionFunc,
//     GID-175). A service that opens a transaction does so to keep several
//     writes atomic; the writes it coordinates are somebody else's.
//
// This is what GID-236 (a repository of a FOREIGN entity) could not say on its
// own: the incident service (2026-08-06, resource-registry
// pkg/integration/push/firebase/domain/service/integration.go) carried
// tx + coreRepo + repo and answered GID-236 with a //nolint stating that
// "orchestration IS the essence of this service" — which is precisely the
// verdict, addressed to the wrong layer. The correct shape: a core integration
// service over the core repository, a firebase service over the firebase
// repository, and a usecase in the module composing the two under the
// transaction.
//
// A _test.go file is not judged: tests live in the same package (GID-250), and
// a harness bundling several doubles is composition of the test, not of a
// service (same relaxation as GID-148/236).
//
// Exceptions: //nolint:gidserviceorch, or settings.exclude
// ("Struct" as a whole | "Struct.Field").
package serviceorch

import (
	"go/ast"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const ruleID = "GID-260"

var defaultSuffixes = []string{"Repository"}

// Analyzer — rule GID-260 with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings — settings of rule GID-260 from .golangci.yml.
type Settings struct {
	// Suffixes — type-name suffixes that mark a repository dependency.
	// Defaults to ["Repository"] when empty.
	Suffixes []string `json:"suffixes"`
	// Exclude — exclusions: "Struct" (as a whole) or "Struct.Field".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-260 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	suffixes := s.Suffixes
	if len(suffixes) == 0 {
		suffixes = defaultSuffixes
	}

	return &analysis.Analyzer{
		Name: "gidserviceorch",
		Doc: ruleID + ": a service does not orchestrate — no second repository and no transaction " +
			"in its fields. Fix: move the composition to a usecase",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, suffixes, s.Exclude)
		},
	}
}

func run(pass *analysis.Pass, suffixes, excl []string) (any, error) {
	if !pathseg.EndsWith(pass.Pkg.Path(), "domain", "service") {
		return nil, nil
	}

	for _, file := range pass.Files {
		if ast.IsGenerated(file) || srcfile.IsTest(pass, file) {
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
				checkServiceStruct(pass, suffixes, excl, ts, st)
			}
		}
	}

	return nil, nil
}

func checkServiceStruct(
	pass *analysis.Pass,
	suffixes, excl []string,
	ts *ast.TypeSpec,
	st *ast.StructType,
) {
	owner := ts.Name.Name
	if strings.HasSuffix(owner, "Options") || slices.Contains(excl, owner) {
		return
	}

	var repos []string
	for _, field := range st.Fields.List {
		t := pass.TypesInfo.TypeOf(field.Type)
		if t == nil {
			continue
		}
		if name, ok := repositoryName(t, suffixes); ok && !exclude.Match(excl, owner, fieldName(field, name)) {
			repos = append(repos, name)
			continue
		}
		if isTransactionFunc(t) && !exclude.Match(excl, owner, fieldName(field, "")) {
			pass.Reportf(field.Pos(),
				"%s: service %q holds a transaction — a service does not coordinate several writes. "+
					"Fix: keep the transaction in a usecase, which calls the services it composes "+
					"(or //nolint:gidserviceorch when explicitly intended)",
				ruleID, owner)
		}
	}

	if len(repos) > 1 {
		pass.Reportf(ts.Name.Pos(),
			"%s: service %q depends on %d repositories (%s) — a service is one entity and its "+
				"repository. Fix: split it into a service per entity and compose them in a usecase "+
				"(or //nolint:gidserviceorch when explicitly intended)",
			ruleID, owner, len(repos), strings.Join(repos, ", "))
	}
}

// repositoryName returns the name of a repository-typed field: a named type
// (through a pointer too) whose name ends with one of the suffixes. Unlike
// GID-236 the interface need not be declared in the same package — a service
// wired to a concrete repository counts as a source just the same.
func repositoryName(t types.Type, suffixes []string) (string, bool) {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return "", false
	}
	obj := named.Obj()
	name := obj.Name()
	for _, suffix := range suffixes {
		if suffix != "" && strings.HasSuffix(name, suffix) {
			return name, true
		}
	}

	return "", false
}

// isTransactionFunc reports whether the field type is a transaction runner:
// a func value returning exactly error whose last parameter is itself a func
// returning exactly error (model.InTransactionFunc — func(ctx context.Context,
// fn func(ctx context.Context) error) error). Matched by signature, so a
// project's own name for the type is handled.
func isTransactionFunc(t types.Type) bool {
	sig, ok := t.Underlying().(*types.Signature)
	if !ok || !returnsOnlyError(sig) {
		return false
	}
	params := sig.Params()
	if params.Len() == 0 {
		return false
	}
	last := params.At(params.Len() - 1)
	lastType := last.Type()
	inner, ok := lastType.Underlying().(*types.Signature)

	return ok && returnsOnlyError(inner)
}

// returnsOnlyError reports whether the signature has exactly one result and it
// is the error interface.
func returnsOnlyError(sig *types.Signature) bool {
	results := sig.Results()
	if results.Len() != 1 {
		return false
	}
	result := results.At(0)
	named, ok := result.Type().(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()

	return obj.Name() == "error" && obj.Pkg() == nil
}

// fieldName returns the field's name for the exclusion list; an embedded field
// is identified by its type name.
func fieldName(field *ast.Field, typeName string) string {
	if len(field.Names) > 0 {
		return field.Names[0].Name
	}

	return typeName
}
