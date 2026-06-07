// Package dbtags реализует правила про db-теги полей структур по слоям:
//
//   - GID-125 (giddbtags): поля entity-структур (DAL) имеют тег маппинга
//     на колонки БД. По умолчанию это db-тег; список допустимых тегов
//     настраивается — например, библиотека для ClickHouse использует тег ch.
//   - GID-168 (gidmodeltags): в /domain/** запрещены db-теги у полей
//     структур. model — чистый бизнес-объект, маппинг на колонки БД живёт
//     в entity (DAL). Список тегов маппинга настраивается тем же Settings
//     (по умолчанию ["db"]); прочие теги (json и пр.) не трогаем.
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

// Analyzer — вариант с дефолтным тегом db.
var Analyzer = NewAnalyzer(Settings{})

// ModelAnalyzer — вариант с дефолтным тегом db.
var ModelAnalyzer = NewModelAnalyzer(Settings{})

// Settings — настройки линтера из .golangci.yml.
type Settings struct {
	// Tags — допустимые теги маппинга (заменяют дефолтный ["db"]).
	Tags []string `json:"tags"`
}

// NewAnalyzer строит анализатор GID-125 из настроек линтера (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	tags := resolveTags(s)
	return &analysis.Analyzer{
		Name: "giddbtags",
		Doc:  ruleID + ": поля entity-структур имеют тег маппинга (" + strings.Join(tags, "/") + ")",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, tags)
		},
	}
}

// NewModelAnalyzer создаёт анализатор GID-168: запрет db-тегов у полей
// структур в /domain/**.
func NewModelAnalyzer(s Settings) *analysis.Analyzer {
	tags := resolveTags(s)
	return &analysis.Analyzer{
		Name: "gidmodeltags",
		Doc:  modelRuleID + ": в /domain/** запрещены теги маппинга на БД (" + strings.Join(tags, "/") + ")",
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
			continue // embedded и приватные поля не маппятся напрямую
		}
		if hasMappingTag(field, tags) {
			continue
		}
		pass.Reportf(field.Pos(),
			"%s: поле %s.%s без тега маппинга (%s) — соответствие entity колонкам БД явное",
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
			continue // нет тега маппинга — поле не нарушает правило
		}
		pass.Reportf(field.Pos(),
			"%s: поле %s.%s с тегом %q в domain-слое — маппинг на БД живёт в /dal/entity",
			modelRuleID, name, fieldName(field), tag)
	}
}

// fieldName возвращает имя поля; для embedded-поля — имя встроенного типа.
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

// mappingTag возвращает первый из tags, присутствующий у поля, либо "".
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
