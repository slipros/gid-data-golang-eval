// Package opstruct implements rule GID-210 (op-struct-fields):
//
//   - GID-210 (gidopstruct): operational Create structs hold a minimal set
//     of fields.
//
// Struct types whose name matches the regexp `^Create[A-Z]` are checked
// (CreateJob, CreateStageInput). The name Create without a following capital
// (as well as CreatedBy, CreatedAt, CreatedSnapshot) does NOT match the
// regexp — those are different words.
//
// The layer is determined by import-path segments:
//
//   - model layer (/domain/model and subpackages): a Create struct does NOT
//     contain the fields ID, CreatedAt, UpdatedAt — they are generated at the
//     service/convert level (source: model.md "model Create structs do not
//     contain ID and CreatedAt").
//   - entity layer (/dal/entity and subpackages): a Create struct does NOT
//     contain the UpdatedAt field — Create holds only INSERT fields. At the
//     same time an entity Create LEGITIMATELY contains ID and CreatedAt, so
//     they are not flagged.
//
// Embedded fields are not checked. Generated files are skipped.
//
// Sources: model.md, entity.md.
package opstruct

import (
	"go/ast"
	"regexp"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-210"

// createName: the name of an operational Create struct — the Create prefix
// followed by a capital letter (CreateJob, CreateStageInput). Create with no
// continuation, CreatedBy, CreatedAt, CreatedSnapshot and Update* do not match.
var createName = regexp.MustCompile(`^Create[A-Z]`)

// The model layer forbids these fields in a Create struct.
var modelForbidden = map[string]string{
	"ID":        "generated at the convert/DB level",
	"CreatedAt": "set at the service/convert level",
	"UpdatedAt": "set at the service/convert level",
}

// The entity layer forbids only UpdatedAt (ID and CreatedAt are legitimate in an entity Create).
var entityForbidden = map[string]string{
	"UpdatedAt": "Create holds only INSERT fields",
}

// Analyzer — rule GID-210: operational Create structs hold a minimal set of fields.
var Analyzer = &analysis.Analyzer{
	Name: "gidopstruct",
	Doc: ruleID + ": operational Create structs hold a minimal set of fields " +
		"(model: no ID/CreatedAt/UpdatedAt; entity: no UpdatedAt). Fix: drop those fields",
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
		// Neither the model nor the entity layer — the rule does not apply.
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
		// Embedded fields are not checked.
		if len(field.Names) == 0 {
			continue
		}
		for _, name := range field.Names {
			reason, bad := forbidden[name.Name]
			if !bad {
				continue
			}
			pass.Reportf(name.Pos(),
				"%s: operational struct %q must not contain field %q (%s). Fix: remove it from Create",
				ruleID, typeName, name.Name, reason)
		}
	}
}
