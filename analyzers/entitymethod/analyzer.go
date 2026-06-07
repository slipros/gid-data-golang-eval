// Package entitymethod реализует правило GID-114: экспортируемые методы
// структур в корневых пакетах слоёв /dal/repository и /domain/service
// именуются от сущности.
//
// Три проверки:
//  1. префикс List запрещён — множественное число вместо него (Jobs, не ListJobs);
//  2. суффикс ByID запрещён — Job(ctx, id) вместо JobByID
//     (только точный суффикс ByID; ByStageID и прочие By<Field>ID разрешены —
//     это уточнение выборки, не получение по первичному ключу);
//  3. имя метода обязано содержать имя сущности — имя типа-ресивера
//     как CamelCase-подстроку (Job → Job, Jobs, CreateJob, JobsByStageID).
//
// Проверка 3 применяется только к ресиверам с осмысленным именем сущности
// (len > 2); однобуквенные/служебные имена не проверяются. Методы-глаголы
// без имени сущности (Close, Ping, Flush) попадут под проверку 3 — они
// легитимны редко и выключаются через exclude/nolint.
//
// Scope — только корневые пакеты слоя (pathseg.EndsWith); подпакеты
// convert/build не задеваются. Конструкторы New* — это функции, а не
// методы, и сюда не попадают.
//
// Исключения:
//   - точечно: //nolint:gidentitymethod
//   - централизованно: settings.exclude в .golangci.yml —
//     записи вида "Close" (имя метода) или "Job.Close" (конкретный тип).
package entitymethod

import (
	"go/ast"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-114"

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

// NewAnalyzer строит анализатор GID-114 из настроек линтера (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidentitymethod",
		Doc: ruleID + ": методы repo/service именуются от сущности — " +
			"без префикса List, без суффикса ByID, с именем сущности",
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
			recv := recvTypeName(fn)
			name := fn.Name.Name
			if exclude.Match(s.Exclude, recv, name) {
				continue
			}
			checkName(pass, fn, recv, name)
		}
	}
	return nil, nil
}

func checkName(pass *analysis.Pass, fn *ast.FuncDecl, recv, name string) {
	// Проверка 1: префикс List запрещён.
	if hasWordPrefix(name, "List") {
		pass.Reportf(fn.Name.Pos(),
			"%s: без префикса List — множественное число: Jobs вместо ListJobs",
			ruleID)
		return
	}
	// Проверка 2: точный суффикс ByID запрещён (ByStageID и прочие разрешены).
	if hasExactByIDSuffix(name) {
		pass.Reportf(fn.Name.Pos(),
			"%s: без суффикса ByID — Job(ctx, id) вместо JobByID",
			ruleID)
		return
	}
	// Проверка 3: имя метода обязано содержать имя сущности (имя ресивера)
	// как CamelCase-подстроку. Только для осмысленных имён сущности:
	// имена длиной <= 2 (T, ID, и т.п.) считаем служебными и не проверяем.
	const minEntityLen = 2
	if len(recv) <= minEntityLen {
		return
	}
	if !containsEntity(name, recv) {
		pass.Reportf(fn.Name.Pos(),
			"%s: имя метода %q должно содержать имя сущности %q "+
				"(Job, Jobs, CreateJob, JobsByStageID; исключения: nolint или settings.exclude)",
			ruleID, name, recv)
	}
}

func inScope(pkgPath string) bool {
	for _, scope := range scopes {
		if pathseg.EndsWith(pkgPath, scope...) {
			return true
		}
	}
	return false
}

// hasWordPrefix: имя начинается со слова word по границе CamelCase
// (List, ListJobs — да; Listen — нет, т.к. следующая руна не заглавная).
func hasWordPrefix(name, word string) bool {
	if name == word {
		return true
	}
	if len(name) <= len(word) || name[:len(word)] != word {
		return false
	}
	r, _ := utf8.DecodeRuneInString(name[len(word):])
	return unicode.IsUpper(r) || unicode.IsDigit(r)
}

// hasExactByIDSuffix: имя оканчивается ровно на "ByID" по границе слова.
// JobByID — да; JobsByStageID — нет (перед ID не "By"); ByID самостоятельно — нет
// (это не имя сущности с суффиксом, но и не валидно — отловит проверка 3).
func hasExactByIDSuffix(name string) bool {
	const suffix = "ByID"
	if !strings.HasSuffix(name, suffix) {
		return false
	}
	return len(name) > len(suffix)
}

// containsEntity: имя метода содержит entity как CamelCase-подстроку.
// Граница слова — начало имени, либо предыдущая руна нижнего регистра
// перед заглавной первой руной entity (CreateJob: ...e|Job).
func containsEntity(name, entity string) bool {
	for idx := strings.Index(name, entity); idx >= 0; idx = nextIndex(name, entity, idx) {
		if isWordBoundary(name, idx) {
			return true
		}
	}
	return false
}

func nextIndex(name, entity string, prev int) int {
	rest := strings.Index(name[prev+1:], entity)
	if rest < 0 {
		return -1
	}
	return prev + 1 + rest
}

// isWordBoundary: позиция idx начинает CamelCase-слово.
// Истина, если idx == 0 или предыдущая руна не заглавная
// (граница camelCase: lowerUpper). Это отсекает совпадения внутри слова.
func isWordBoundary(name string, idx int) bool {
	if idx == 0 {
		return true
	}
	prev, _ := utf8.DecodeLastRuneInString(name[:idx])
	return !unicode.IsUpper(prev)
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
