// Package serviceorch implements rule GID-260: a domain service does not
// orchestrate — it is one entity and the repository of that entity. Composing
// several data sources, and holding them together in a transaction, is what a
// usecase is for.
//
// A service may hold a transaction function to keep operations on its entity
// atomic. Only multiple repository dependencies imply orchestration here.
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
		Doc: ruleID + ": a service does not orchestrate — no second repository " +
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

// fieldName returns the field's name for the exclusion list; an embedded field
// is identified by its type name.
func fieldName(field *ast.Field, typeName string) string {
	if len(field.Names) > 0 {
		return field.Names[0].Name
	}

	return typeName
}
