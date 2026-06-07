// Package createupdate реализует правило GID-112: методы, создающие
// сущность или обновляющие состояние (Create*/Update*), в repo и service
// возвращают только error. Если после создания нужны данные — вызывающий
// код получает их отдельным запросом.
//
// Исключения (бывает удобно сразу получить сущность):
//   - точечно: //nolint:gidcreateupdate
//   - централизованно: settings.exclude в .golangci.yml —
//     записи вида "CreateSession" (имя метода) или "Job.CreateJob"
//     (конкретный тип).
package createupdate

import (
	"go/ast"
	"go/types"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-112"

var verbs = []string{"Create", "Update"}

var scopes = [][]string{
	{"dal", "repository"},
	{"domain", "service"},
}

// Analyzer — вариант с настройками по умолчанию (без исключений).
var Analyzer = NewAnalyzer(Settings{})

// Settings — настройки линтера из .golangci.yml.
type Settings struct {
	// Exclude — методы-исключения: "Метод" или "Тип.Метод".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer строит анализатор GID-112 из настроек линтера (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidcreateupdate",
		Doc:  ruleID + ": методы Create*/Update* в repo и service возвращают только error",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s)
		},
	}
}

func run(pass *analysis.Pass, s Settings) (any, error) {
	if !inScope(pass.Pkg.Path()) {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || !fn.Name.IsExported() {
				continue
			}
			if !hasVerbPrefix(fn.Name.Name) {
				continue
			}
			if exclude.Match(s.Exclude, recvTypeName(fn), fn.Name.Name) {
				continue
			}
			checkResults(pass, fn)
		}
	}
	return nil, nil
}

func checkResults(pass *analysis.Pass, fn *ast.FuncDecl) {
	obj, ok := pass.TypesInfo.Defs[fn.Name].(*types.Func)
	if !ok {
		return
	}
	sig, ok := obj.Type().(*types.Signature)
	if !ok {
		return
	}
	results := sig.Results()
	if results.Len() == 0 {
		return
	}
	for v := range results.Variables() {
		if isError(v.Type()) {
			continue
		}
		pass.Reportf(fn.Name.Pos(),
			"%s: метод %q создаёт/обновляет состояние — возвращает только error, данные получают отдельным запросом "+
				"(исключения: nolint или settings.exclude)",
			ruleID, fn.Name.Name)
		return
	}
}

func inScope(pkgPath string) bool {
	for _, scope := range scopes {
		if pathseg.Contains(pkgPath, scope...) {
			return true
		}
	}
	return false
}

// hasVerbPrefix: имя начинается со слова Create/Update
// (CreateJob, Update — да; CreatedAt — нет).
func hasVerbPrefix(name string) bool {
	for _, verb := range verbs {
		if name == verb {
			return true
		}
		if len(name) > len(verb) && name[:len(verb)] == verb {
			r, _ := utf8.DecodeRuneInString(name[len(verb):])
			if unicode.IsUpper(r) || unicode.IsDigit(r) {
				return true
			}
		}
	}
	return false
}

func recvTypeName(fn *ast.FuncDecl) string {
	t := fn.Recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	if ident, ok := t.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

func isError(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj.Pkg() == nil && obj.Name() == "error"
}
