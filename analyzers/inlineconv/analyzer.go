// Package inlineconv реализует правило GID-215 (no-inline-entity-literal):
// конвертация model ↔ entity живёт только в convert-пакетах.
//
// Источник: service.md «Конвертация всегда выполняется через пакет convert».
//
// Scope: пакеты domain-слоя (pathseg.Contains(pkgPath, "domain")), КРОМЕ
// пакетов с сегментом convert (там конвертация и должна жить).
//
// Что запрещено: composite literal с ≥1 элементом, чей именованный тип
// (struct или именованный слайс) объявлен в пакете entity-слоя
// (pathseg.Contains(пакет типа, "dal", "entity") — включая подпакеты
// filter/enum). Инлайн-заполнение entity вне convert-пакета означает, что
// конвертация размазана по domain-слою.
//
// Что НЕ запрещено:
//   - пустой литерал (entity.Snapshot{} — zero value);
//   - литерал model-типа (model в domain — норма);
//   - entity-литерал внутри convert-пакета сервиса.
//
// Флагается только внешний (outermost) entity-литерал: вложенные внутри уже
// зафлаганного повторно не репортятся. Карты/слайсы entity-типов (map[K]entity.X,
// []entity.X) сами по себе не флагаются — флагается именно литерал именованного
// entity-типа среди их элементов (он будет внешним).
//
// LoadMode — TypesInfo (нужны типы, чтобы определить пакет именованного типа).
// _test.go и сгенерированные файлы (ast.IsGenerated) пропускаются.
// Точечное отключение: //nolint:gidinlineconv.
package inlineconv

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-215"

// Analyzer — правило GID-215: инлайн-заполнение entity-типа в domain-слое
// запрещено, конвертация живёт в convert-пакете.
var Analyzer = &analysis.Analyzer{
	Name: "gidinlineconv",
	Doc:  ruleID + ": inline-filling an entity type in the domain layer is forbidden; conversion lives in a convert package. Fix: move it to a convert function",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()

	// Зона правила: domain-слой, но не convert-пакеты (там и живёт конвертация).
	if !pathseg.Contains(pkgPath, "domain") || pathseg.Contains(pkgPath, "convert") {
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
				return true // спускаемся внутрь — нарушение может быть глубже.
			}
			pass.Reportf(lit.Pos(),
				"%s: inline-filling the entity type %s in the domain layer is forbidden. "+
					"Fix: put conversion in a convert package (<Dst><Type>From<Src>)",
				ruleID, name)
			// Внешний entity-литерал зафлаган — внутрь не спускаемся,
			// чтобы не репортить вложенные entity-литералы повторно.
			return false
		})
	}
	return nil, nil
}

// entityLitName сообщает, является ли lit непустым composite-литералом
// именованного entity-типа (struct или именованный слайс из /dal/entity),
// и возвращает его отображаемое имя (pkg.Type).
func entityLitName(pass *analysis.Pass, lit *ast.CompositeLit) (string, bool) {
	if len(lit.Elts) == 0 {
		return "", false // пустой литерал — zero value, разрешён.
	}
	t := pass.TypesInfo.TypeOf(lit)
	if t == nil {
		return "", false
	}
	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return "", false // анонимные struct/slice, map[...], []... — не именованный тип.
	}
	// Только struct или именованный слайс — карты/массивы как сам тип литерала
	// не считаем (их именованные элементы обрабатываются отдельными литералами).
	switch named.Underlying().(type) {
	case *types.Struct, *types.Slice:
	default:
		return "", false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	if pkg == nil || !pathseg.Contains(pkg.Path(), "dal", "entity") {
		return "", false
	}
	return pkg.Name() + "." + obj.Name(), true
}
