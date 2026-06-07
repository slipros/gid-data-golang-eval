// Package filterplace реализует правило GID-171 (filter-location):
//
//   - GID-171 (gidfilterplace): фильтры list-операций живут в своём месте
//     слоя. Entity-фильтры — в /dal/entity/filter, model-фильтры — в
//     model-слое (/domain/model и его подпакеты, например /domain/model/filter).
//
// Проверяются только объявления STRUCT-типов с именем-фильтром (Filter* или
// *Filter), чтобы не задевать FilterFunc, интерфейсы и алиасы. Имя считается
// фильтром, если слово Filter стоит на границе: префикс Filter + заглавная
// буква/конец (FilterJobs, Filter) либо суффикс Filter (JobsFilter). Filterable
// не флагается — после Filter идёт строчная буква, это другое слово.
//
// Источники: model.md, entity.md.
package filterplace

import (
	"go/ast"
	"regexp"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-171"

// filterName: префикс Filter перед заглавной буквой/цифрой или концом имени
// (Filter, FilterJobs, Filter2) либо суффикс Filter после строчной буквы/цифры
// или в начале имени (JobsFilter, Filter). Filterable не матчится — после
// Filter идёт строчная буква.
var filterName = regexp.MustCompile(`(^Filter([A-Z0-9].*)?$)|([a-z0-9]Filter$)`)

// Analyzer — правило GID-171: list-operation filters live in /dal/entity/filter (entity) or /domain/model (model). Fix: move the filter there.
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

	// Слой не dal и не domain — правило не применяется.
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
				// Только struct-типы: FilterFunc, интерфейсы и алиасы не трогаем.
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
