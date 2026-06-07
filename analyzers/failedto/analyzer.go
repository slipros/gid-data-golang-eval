// Package failedto реализует правило GID-184 (линтер gidfailedto):
// сообщение об ошибке описывает операцию, а не факт провала
// (Uber: error wrapping). Строковый литерал-сообщение в вызовах
// errors.Wrap/Wrapf/WithMessage/WithMessagef/Errorf/New пакета
// github.com/pkg/errors не должно начинаться (регистронезависимо) с
// запрещённого префикса (failed to, failed, unable to, error, couldn't,
// could not, can't, cannot и т.п.).
//
// Вместо "failed to select user" → "select user": при разворачивании
// цепочки сообщение читается как описание операций.
//
// pkg/errors определяется по import-пути github.com/pkg/errors через
// TypesInfo. Не-литеральное сообщение (переменная, конкатенация с
// переменной) не проверяется. Сгенерированный код (ast.IsGenerated)
// пропускается.
package failedto

import (
	"go/ast"
	"go/constant"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

const ruleID = "GID-184"

// errFuncs — функции pkg/errors, у которых аргумент-сообщение проверяется.
// Значение — индекс аргумента-сообщения в вызове.
var errFuncs = map[string]int{
	"Wrap":         1, // Wrap(err, message)
	"Wrapf":        1, // Wrapf(err, format, ...)
	"WithMessage":  1, // WithMessage(err, message)
	"WithMessagef": 1, // WithMessagef(err, format, ...)
	"Errorf":       0, // Errorf(format, ...)
	"New":          0, // New(message)
}

// defaultPrefixes — запрещённые префиксы сообщения (регистронезависимо).
var defaultPrefixes = []string{
	"failed to",
	"failed",
	"unable to",
	"error",
	"couldn't",
	"could not",
	"can't",
	"cannot",
}

// Analyzer — вариант с настройками по умолчанию.
var Analyzer = NewAnalyzer(Settings{})

// Settings — настройки линтера из .golangci.yml.
type Settings struct {
	// Prefixes — запрещённые префиксы сообщения. Если задан — замещает
	// дефолтный список целиком.
	Prefixes []string `json:"prefixes"`
}

// NewAnalyzer строит анализатор GID-184 с указанными настройками.
func NewAnalyzer(s Settings) *analysis.Analyzer {
	prefixes := defaultPrefixes
	if len(s.Prefixes) > 0 {
		prefixes = s.Prefixes
	}
	lower := make([]string, len(prefixes))
	for i, p := range prefixes {
		lower[i] = strings.ToLower(p)
	}
	return &analysis.Analyzer{
		Name: "gidfailedto",
		Doc:  ruleID + ": сообщение об ошибке описывает операцию, а не факт провала",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, lower)
		},
	}
}

func run(pass *analysis.Pass, prefixes []string) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			name := pkgErrorsCallName(pass, call)
			if name == "" {
				return true
			}
			idx, ok := errFuncs[name]
			if !ok || idx >= len(call.Args) {
				return true
			}
			msg, ok := stringLiteral(pass, call.Args[idx])
			if !ok {
				return true
			}
			if prefix, hit := matchPrefix(msg, prefixes); hit {
				pass.Reportf(call.Args[idx].Pos(),
					"%s: сообщение ошибки начинается с %q — опишите операцию: вместо \"failed to select user\" → \"select user\"",
					ruleID, prefix)
			}
			return true
		})
	}
	return nil, nil
}

// matchPrefix сообщает, начинается ли сообщение (регистронезависимо) с
// запрещённого префикса как со слова (за префиксом — конец строки или не
// буква/цифра, чтобы "failure" не матчилось на "failed"... на "fail" —
// здесь префиксов "fail" нет, но граница слова защищает от подстрок).
func matchPrefix(msg string, prefixes []string) (string, bool) {
	low := strings.ToLower(strings.TrimSpace(msg))
	for _, p := range prefixes {
		if !strings.HasPrefix(low, p) {
			continue
		}
		rest := low[len(p):]
		if rest == "" {
			return p, true
		}
		r := rune(rest[0])
		if !isWordRune(r) {
			return p, true
		}
	}
	return "", false
}

func isWordRune(r rune) bool {
	return r == '_' ||
		(r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9')
}

// stringLiteral возвращает значение строкового литерала (включая
// нетипизированную константу-строку). Переменная или конкатенация с
// переменной → (_, false): такие сообщения не проверяются.
func stringLiteral(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	tv, ok := pass.TypesInfo.Types[expr]
	if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
		return "", false
	}
	// Только литералы и константы — переменные значения не имеют.
	return constant.StringVal(tv.Value), true
}

// pkgErrorsCallName возвращает имя функции github.com/pkg/errors,
// если call — её вызов; иначе "".
func pkgErrorsCallName(pass *analysis.Pass, call *ast.CallExpr) string {
	const pkgErrorsPath = "github.com/pkg/errors"
	fn := typeutil.Callee(pass.TypesInfo, call)
	f, ok := fn.(*types.Func)
	if !ok || f.Pkg() == nil {
		return ""
	}
	pkg := f.Pkg()
	if pkg.Path() != pkgErrorsPath {
		return ""
	}
	return f.Name()
}
