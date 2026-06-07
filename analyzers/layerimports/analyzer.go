// Package layerimports реализует правила направления зависимостей между
// слоями Clean Architecture.
//
// GID-132 (layer-imports):
//   - /dal/** не импортирует /domain/** — repository работает только с entity;
//   - /domain/model не импортирует /dal/** — model чистый;
//   - /domain/usecase не импортирует /dal/** — usecase работает только
//     с model, с DAL общается через сервисы;
//   - /domain/service не импортирует /dal/repository — зависимость от
//     репозитория описывается интерфейсом рядом с потребителем.
//     Импорт /dal/entity сервису разрешён: он конвертирует model <-> entity.
//
// GID-170 (no-event-import):
//   - /domain/** не импортирует /event/**;
//   - /dal/** не импортирует /event/** — event-слой (kafka producer/consumer,
//     DTO) зависит от domain/model и конвертирует model <-> DTO, не наоборот.
//
// GID-172 (client-no-entity):
//   - /client/** не импортирует /dal/** — у клиента свои типы, он ничего
//     не знает о entity/repository.
package layerimports

import (
	"go/ast"
	"strconv"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

var layerRules = []layerRule{
	{
		id:     "GID-132",
		scope:  []string{"dal"},
		banned: [][]string{{"domain"}},
		reason: "dal-слой работает только с entity, domain-типы ему недоступны",
	},
	{
		id:     "GID-132",
		scope:  []string{"domain", "model"},
		banned: [][]string{{"dal"}},
		reason: "model не зависит от dal-слоя",
	},
	{
		id:     "GID-132",
		scope:  []string{"domain", "usecase"},
		banned: [][]string{{"dal"}},
		reason: "usecase работает только с model, с DAL общается через сервисы",
	},
	{
		id:     "GID-132",
		scope:  []string{"domain", "service"},
		banned: [][]string{{"dal", "repository"}},
		reason: "сервис зависит от репозитория через интерфейс рядом с потребителем",
	},
	{
		id:     "GID-170",
		scope:  []string{"domain"},
		banned: [][]string{{"event"}},
		reason: "domain не зависит от event-слоя: event конвертирует model <-> DTO, не наоборот",
	},
	{
		id:     "GID-170",
		scope:  []string{"dal"},
		banned: [][]string{{"event"}},
		reason: "dal не зависит от event-слоя: event конвертирует model <-> DTO, не наоборот",
	},
	{
		id:     "GID-172",
		scope:  []string{"client"},
		banned: [][]string{{"dal"}},
		reason: "у клиента свои типы, он ничего не знает о entity/repository из dal-слоя",
	},
}

// Analyzer — правила GID-132/170/172: направление импортов между слоями.
var Analyzer = &analysis.Analyzer{
	Name: "gidlayerimports",
	Doc: "GID-132/GID-170/GID-172: направление зависимостей между слоями " +
		"(dal -> entity, domain -> model; domain/dal не импортируют event; " +
		"client не импортирует dal)",
	Run: run,
}

// layerRule: пакетам в scope запрещены импорты banned. id — ID правила,
// под которым рапортуется нарушение.
type layerRule struct {
	id     string
	scope  []string
	banned [][]string
	reason string
}

func run(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()
	//nolint:gidallptr // плагин не зависит от внутренней библиотеки gdhelper
	for _, rule := range layerRules {
		if !pathseg.Contains(pkgPath, rule.scope...) {
			continue
		}
		for _, file := range pass.Files {
			if ast.IsGenerated(file) {
				continue
			}
			checkImports(pass, &rule, file)
		}
	}
	return nil, nil
}

func checkImports(pass *analysis.Pass, rule *layerRule, file *ast.File) {
	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		for _, banned := range rule.banned {
			if pathseg.Contains(path, banned...) {
				pass.Reportf(imp.Pos(),
					"%s: пакету %q запрещён импорт %q — %s",
					rule.id, pass.Pkg.Path(), path, rule.reason)
			}
		}
	}
}
