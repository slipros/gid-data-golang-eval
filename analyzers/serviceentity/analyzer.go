// Package serviceentity implements rule GID-236: a service works with
// exactly one entity — by default it calls the repository of its own
// entity only, unless explicitly stated otherwise (service.md).
//
// The deterministic check complements GID-148 (servicesingle: a service
// must not depend on another service): here we catch a service that
// injects a repository interface of a FOREIGN entity. For every struct
// declared at the root of /domain/service (except *Options types), for
// every field whose type is a named interface declared in the same
// package (GID-134 — interfaces live at the consumer) with a name ending
// in one of the configured suffixes (settings.suffixes, default
// "Repository"): the entity is that name with the suffix stripped; if
// the entity does not match the struct's own name, it is a violation.
//
// Per-project relaxation — settings.suffixes, settings.exclude
// ("Struct" as a whole | "Struct.Field"). Pointwise — //nolint:gidserviceentity
// when a cross-entity call is explicitly intended (service.md: "unless
// explicitly stated otherwise").
package serviceentity

import (
	"go/ast"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-236"

var defaultSuffixes = []string{"Repository"}

// Analyzer — rule GID-236 with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings — settings of rule GID-236 from .golangci.yml.
type Settings struct {
	// Suffixes — interface-name suffixes that mark a repository dependency.
	// Defaults to ["Repository"] when empty.
	Suffixes []string `json:"suffixes"`
	// Exclude — exclusions: "Struct" (as a whole) or "Struct.Field".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-236 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	suffixes := s.Suffixes
	if len(suffixes) == 0 {
		suffixes = defaultSuffixes
	}
	return &analysis.Analyzer{
		Name: "gidserviceentity",
		Doc: ruleID + ": a service works with exactly one entity; by default it calls the " +
			"repository of its own entity. Fix: orchestrate several entities in usecase, " +
			"or //nolint:gidserviceentity when explicitly intended",
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
				checkServiceStruct(pass, suffixes, excl, ts.Name.Name, st)
			}
		}
	}
	return nil, nil
}

func checkServiceStruct(pass *analysis.Pass, suffixes, excl []string, owner string, st *ast.StructType) {
	if strings.HasSuffix(owner, "Options") || slices.Contains(excl, owner) {
		return
	}
	for _, field := range st.Fields.List {
		ifaceName, ok := samePackageInterface(pass, field.Type)
		if !ok {
			continue
		}
		suffix, ok := matchingSuffix(ifaceName, suffixes)
		if !ok {
			continue
		}
		entity := strings.TrimSuffix(ifaceName, suffix)
		if entity == owner {
			continue
		}
		if exclude.Match(excl, owner, fieldName(field, ifaceName)) {
			continue
		}
		pass.Reportf(field.Pos(),
			"%s: service %q uses repository %q of another entity. Fix: a service works with "+
				"exactly one entity — orchestrate several entities in usecase (or "+
				"//nolint:gidserviceentity when explicitly intended)",
			ruleID, owner, ifaceName)
	}
}

// samePackageInterface returns the type name if the field type is a named
// interface declared in the same package (GID-134: interfaces live next to
// the consumer). Interfaces are never referenced through a pointer.
func samePackageInterface(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	t := pass.TypesInfo.TypeOf(expr)
	if t == nil {
		return "", false
	}
	named, ok := t.(*types.Named)
	if !ok {
		return "", false
	}
	namedObj := named.Obj()
	if namedObj.Pkg() != pass.Pkg {
		return "", false
	}
	if _, ok := named.Underlying().(*types.Interface); !ok {
		return "", false
	}
	return namedObj.Name(), true
}

// matchingSuffix returns the first configured suffix the interface name ends with.
func matchingSuffix(name string, suffixes []string) (string, bool) {
	for _, suf := range suffixes {
		if strings.HasSuffix(name, suf) {
			return suf, true
		}
	}
	return "", false
}

// fieldName — the field's own name for settings.exclude matching ("Struct.Field").
// An embedded field has no name of its own; the interface's type name stands in for it.
func fieldName(field *ast.Field, ifaceName string) string {
	if len(field.Names) == 0 {
		return ifaceName
	}
	return field.Names[0].Name
}
