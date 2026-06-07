// Package utilpkg реализует правило GID-187 (no-util-package): имена пакетов
// вроде util/utils/common/helper/helpers/shared/misc/lib/base запрещены.
// Такой пакет — свалка без зоны ответственности: по имени нельзя понять, что
// он предоставляет. Пакет называют по тому, что в нём лежит (например, parse,
// retry, money).
//
// Проверяется имя пакета (pass.Pkg.Name(), что совпадает с последним сегментом
// пути) — регистронезависимо. Суффикс _test у тестового пакета нормализуется
// (utils_test → utils). Один репорт на пакет: на package-клаузе первого
// (не сгенерированного) файла. Словарь имён настраивается settings.names
// и полностью замещает дефолтный список.
//
// LoadMode — Syntax: типы не нужны, хватает имени пакета и AST.
package utilpkg

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-187"

// Analyzer — вариант с дефолтным чёрным списком имён.
var Analyzer = NewAnalyzer(Settings{})

// Settings — настройки линтера из .golangci.yml.
type Settings struct {
	// Names — чёрный список имён пакетов (заменяет дефолтный список).
	Names []string `json:"names"`
}

// NewAnalyzer строит анализатор GID-187 из настроек линтера (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	names := resolveNames(s)
	blacklist := make(map[string]struct{}, len(names))
	for _, n := range names {
		blacklist[strings.ToLower(n)] = struct{}{}
	}
	return &analysis.Analyzer{
		Name: "gidutilpkg",
		Doc: ruleID + ": forbid junk-drawer packages (util, utils, common, helper, …). " +
			"Fix: name the package after what it provides",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, blacklist)
		},
	}
}

func resolveNames(s Settings) []string {
	if len(s.Names) == 0 {
		return []string{
			"util",
			"utils",
			"common",
			"helper",
			"helpers",
			"shared",
			"misc",
			"lib",
			"base",
		}
	}
	return s.Names
}

func run(pass *analysis.Pass, blacklist map[string]struct{}) (any, error) {
	name := normalize(pass.Pkg.Name())
	if _, banned := blacklist[name]; !banned {
		return nil, nil
	}

	// Один репорт на пакет — на package-клаузе первого не сгенерированного файла.
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		pass.Reportf(file.Name.Pos(),
			"%s: package %q is a junk drawer with no responsibility. Fix: name the package after what it provides",
			ruleID, pass.Pkg.Name())
		return nil, nil
	}
	return nil, nil
}

// normalize приводит имя пакета к нижнему регистру и снимает суффикс _test
// у тестовых пакетов (utils_test → utils).
func normalize(name string) string {
	name = strings.ToLower(name)
	return strings.TrimSuffix(name, "_test")
}
