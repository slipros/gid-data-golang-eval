// Package dbtags implements the per-layer rules about db tags on struct fields:
//
//   - GID-125 (giddbtags): fields of entity structs (DAL) have a tag mapping
//     them to DB columns. By default this is the db tag; the list of allowed
//     tags is configurable — e.g. the ClickHouse library uses the ch tag.
//   - GID-168 (gidmodeltags): db tags on struct fields are forbidden in
//     /domain/**. A model is a pure business object; mapping to DB columns lives
//     in entity (DAL). The list of mapping tags is configured by the same Settings
//     (default ["db"]); other tags (json, etc.) are left alone.
package dbtags

import (
	"go/ast"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const (
	ruleID      = "GID-125"
	modelRuleID = "GID-168"
)

// Analyzer — the variant with the default db tag.
var Analyzer = NewAnalyzer(Settings{})

// ModelAnalyzer — the variant with the default db tag.
var ModelAnalyzer = NewModelAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Tags — allowed mapping tags (they replace the default ["db"]).
	Tags []string `json:"tags"`
}

// NewAnalyzer builds the GID-125 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	tags := resolveTags(s)
	return &analysis.Analyzer{
		Name: "giddbtags",
		Doc:  ruleID + ": entity struct fields must have a mapping tag (" + strings.Join(tags, "/") + ")",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, tags)
		},
	}
}

// NewModelAnalyzer creates the GID-168 analyzer: a ban on db tags on struct
// fields in /domain/**.
func NewModelAnalyzer(s Settings) *analysis.Analyzer {
	tags := resolveTags(s)
	return &analysis.Analyzer{
		Name: "gidmodeltags",
		Doc:  modelRuleID + ": db mapping tags are forbidden in /domain/** (" + strings.Join(tags, "/") + ")",
		Run: func(pass *analysis.Pass) (any, error) {
			return runModel(pass, tags)
		},
	}
}

func resolveTags(s Settings) []string {
	if len(s.Tags) == 0 {
		return []string{"db"}
	}
	return s.Tags
}

func run(pass *analysis.Pass, tags []string) (any, error) {
	if !pathseg.EndsWith(pass.Pkg.Path(), "dal", "entity") {
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
				checkStruct(pass, ts.Name.Name, st, tags)
			}
		}
	}
	return nil, nil
}

func checkStruct(pass *analysis.Pass, name string, st *ast.StructType, tags []string) {
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 || !field.Names[0].IsExported() {
			continue // embedded and private fields are not mapped directly
		}
		if hasMappingTag(field, tags) {
			continue
		}
		pass.Reportf(field.Pos(),
			"%s: field %s.%s has no mapping tag (%s). Fix: add a tag so entity-to-column mapping is explicit",
			ruleID, name, field.Names[0].Name, strings.Join(tags, "/"))
	}
}

func runModel(pass *analysis.Pass, tags []string) (any, error) {
	if !pathseg.Contains(pass.Pkg.Path(), "domain") {
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
				checkModelStruct(pass, ts.Name.Name, st, tags)
			}
		}
	}
	return nil, nil
}

func checkModelStruct(pass *analysis.Pass, name string, st *ast.StructType, tags []string) {
	for _, field := range st.Fields.List {
		tag := mappingTag(field, tags)
		if tag == "" {
			continue // no mapping tag — the field does not violate the rule
		}
		pass.Reportf(field.Pos(),
			"%s: field %s.%s has a %q tag in the domain layer. Fix: keep db mapping in /dal/entity",
			modelRuleID, name, fieldName(field), tag)
	}
}

// fieldName returns the field name; for an embedded field — the embedded type name.
func fieldName(field *ast.Field) string {
	if len(field.Names) > 0 {
		return field.Names[0].Name
	}
	switch t := field.Type.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name
		}
	case *ast.SelectorExpr:
		return t.Sel.Name
	}
	return "<embedded>"
}

// mappingTag returns the first of tags present on the field, or "".
func mappingTag(field *ast.Field, tags []string) string {
	if field.Tag == nil {
		return ""
	}
	raw, err := strconv.Unquote(field.Tag.Value)
	if err != nil {
		return ""
	}
	st := reflect.StructTag(raw)
	for _, tag := range tags {
		if _, ok := st.Lookup(tag); ok {
			return tag
		}
	}
	return ""
}

func hasMappingTag(field *ast.Field, tags []string) bool {
	if field.Tag == nil {
		return false
	}
	raw, err := strconv.Unquote(field.Tag.Value)
	if err != nil {
		return false
	}
	st := reflect.StructTag(raw)
	for _, tag := range tags {
		if st.Get(tag) != "" {
			return true
		}
	}
	return false
}
