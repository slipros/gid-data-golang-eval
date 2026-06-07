// Package opstruct реализует правило GID-210 (op-struct-fields):
//
//   - GID-210 (gidopstruct): операционные Create-структуры содержат
//     минимальный набор полей.
//
// Проверяются struct-типы, чьё имя матчит regexp `^Create[A-Z]` (CreateJob,
// CreateStageInput). Имя Create без следующей заглавной (а также CreatedBy,
// CreatedAt, CreatedSnapshot) под regexp НЕ попадает — это другие слова.
//
// Слой определяется по сегментам import-пути:
//
//   - model-слой (/domain/model и подпакеты): Create-структура НЕ содержит
//     полей ID, CreatedAt, UpdatedAt — они генерируются на уровне
//     service/convert (источник: model.md «модельные Create-структуры не
//     содержат ID и CreatedAt»).
//   - entity-слой (/dal/entity и подпакеты): Create-структура НЕ содержит поля
//     UpdatedAt — Create содержит только поля INSERT. При этом entity-Create
//     ЛЕГИТИМНО содержит ID и CreatedAt, поэтому они не флагаются.
//
// Embedded-поля не проверяются. Сгенерированные файлы пропускаются.
//
// Источники: model.md, entity.md.
package opstruct

import (
	"go/ast"
	"regexp"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-210"

// createName: имя операционной Create-структуры — префикс Create перед
// заглавной буквой (CreateJob, CreateStageInput). Create без продолжения,
// CreatedBy, CreatedAt, CreatedSnapshot и Update* под regexp не попадают.
var createName = regexp.MustCompile(`^Create[A-Z]`)

// model-слой запрещает эти поля в Create-структуре.
var modelForbidden = map[string]string{
	"ID":        "генерируется на уровне convert/БД",
	"CreatedAt": "проставляется на уровне service/convert",
	"UpdatedAt": "проставляется на уровне service/convert",
}

// entity-слой запрещает только UpdatedAt (ID и CreatedAt в entity-Create легитимны).
var entityForbidden = map[string]string{
	"UpdatedAt": "Create содержит только поля INSERT",
}

// Analyzer — правило GID-210: операционные Create-структуры содержат минимальный набор полей.
var Analyzer = &analysis.Analyzer{
	Name: "gidopstruct",
	Doc: ruleID + ": операционные Create-структуры содержат минимальный набор полей " +
		"(model: без ID/CreatedAt/UpdatedAt; entity: без UpdatedAt)",
	Run: run,
}

func run(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()

	inModel := pathseg.Contains(pkgPath, "domain", "model")
	inEntity := pathseg.Contains(pkgPath, "dal", "entity")

	var forbidden map[string]string
	switch {
	case inModel:
		forbidden = modelForbidden
	case inEntity:
		forbidden = entityForbidden
	default:
		// Не model- и не entity-слой — правило не применяется.
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
				if !createName.MatchString(ts.Name.Name) {
					continue
				}
				checkStruct(pass, ts.Name.Name, st, forbidden)
			}
		}
	}
	return nil, nil
}

func checkStruct(pass *analysis.Pass, typeName string, st *ast.StructType, forbidden map[string]string) {
	for _, field := range st.Fields.List {
		// Embedded-поля не проверяем.
		if len(field.Names) == 0 {
			continue
		}
		for _, name := range field.Names {
			reason, bad := forbidden[name.Name]
			if !bad {
				continue
			}
			pass.Reportf(name.Pos(),
				"%s: операционная структура %q не должна содержать поле %q (%s) — убери его из Create",
				ruleID, typeName, name.Name, reason)
		}
	}
}
