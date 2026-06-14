// Package filterplace implements rule GID-171 (filter-location):
//
//   - GID-171 (gidfilterplace): list-operation filters live in their layer's
//     place. Entity filters — in /dal/entity/filter, model filters — in the
//     model layer (/domain/model and its subpackages, e.g. /domain/model/filter).
//
// Only declarations of STRUCT types with a filter name (Filter* or
// *Filter) are checked, so as not to touch FilterFunc, interfaces and aliases.
// A name counts as a filter when the word Filter sits on a boundary: the prefix
// Filter + a capital letter/end (FilterJobs, Filter) or the suffix Filter
// (JobsFilter). Filterable is not flagged — Filter is followed by a lowercase
// letter, which makes it a different word.
//
// Sources: model.md, entity.md.
package filterplace

import (
	"go/ast"
	"regexp"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-171"

// filterName: the prefix Filter before a capital letter/digit or the end of the
// name (Filter, FilterJobs, Filter2), or the suffix Filter after a lowercase
// letter/digit or at the start of the name (JobsFilter, Filter). Filterable is
// not matched — Filter is followed by a lowercase letter.
var filterName = regexp.MustCompile(`(^Filter([A-Z0-9].*)?$)|([a-z0-9]Filter$)`)

// Analyzer — rule GID-171: list-operation filters live in /dal/entity/filter (entity) or /domain/model (model). Fix: move the filter there.
var Analyzer = &analysis.Analyzer{
	Name: "gidfilterplace",
	Doc:  ruleID + ": list-operation filters live in /dal/entity/filter (entity) or /domain/model (model). Fix: move the filter there",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()

	inDAL := pathseg.Contains(pkgPath, "dal")
	inEntityFilter := pathseg.Contains(pkgPath, "dal", "entity", "filter")
	inDomain := pathseg.Contains(pkgPath, "domain")
	inModel := pathseg.Contains(pkgPath, "domain", "model")

	// Neither the dal nor the domain layer — the rule does not apply.
	dalViolating := inDAL && !inEntityFilter
	domainViolating := inDomain && !inModel
	if !dalViolating && !domainViolating {
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
				// Struct types only: FilterFunc, interfaces and aliases are untouched.
				if _, ok := ts.Type.(*ast.StructType); !ok {
					continue
				}
				if !filterName.MatchString(ts.Name.Name) {
					continue
				}
				if dalViolating {
					pass.Reportf(ts.Name.Pos(),
						"%s: filter %q must live in /dal/entity/filter. Fix: move it there", ruleID, ts.Name.Name)
				} else {
					pass.Reportf(ts.Name.Pos(),
						"%s: filter %q must live in /domain/model. Fix: move it there", ruleID, ts.Name.Name)
				}
			}
		}
	}
	return nil, nil
}
