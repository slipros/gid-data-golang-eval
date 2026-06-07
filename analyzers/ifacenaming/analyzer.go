// Package ifacenaming реализует правило GID-173 (iface-entity-prefix):
//
//   - GID-173 (gidifacenaming): интерфейсы зависимостей именуются с
//     префиксом сущности (`HelloRepository`, `HelloConnection`). Голое имя
//     роли (`Repository`, `Connection`, …) запрещено: по нему нельзя понять,
//     зависимостью какой сущности является интерфейс.
//
// Scope: пакеты в слоях /domain/service, /domain/usecase, /dal/repository,
// /server/**, /event/** (pathseg.Contains). Проверяются объявления
// interface-типов, чьё имя ТОЧНО совпадает с голой ролью из словаря.
// Сгенерированный код пропускается.
package ifacenaming

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-173"

// Analyzer — вариант с дефолтным словарём ролей.
var Analyzer = NewAnalyzer(Settings{})

// Settings — настройки линтера из .golangci.yml.
type Settings struct {
	// Names — словарь голых ролей (заменяет дефолтный список).
	Names []string `json:"names"`
}

// NewAnalyzer строит анализатор GID-173 из настроек линтера (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	names := resolveNames(s)
	roles := make(map[string]struct{}, len(names))
	for _, n := range names {
		roles[n] = struct{}{}
	}
	return &analysis.Analyzer{
		Name: "gidifacenaming",
		Doc:  ruleID + ": интерфейсы зависимостей именуются с префиксом сущности (например, HelloRepository)",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, roles)
		},
	}
}

func resolveNames(s Settings) []string {
	if len(s.Names) == 0 {
		return []string{
			"Repository",
			"Service",
			"Client",
			"Connection",
			"Producer",
			"Consumer",
			"Validator",
			"Storage",
			"Cache",
		}
	}
	return s.Names
}

func run(pass *analysis.Pass, roles map[string]struct{}) (any, error) {
	if !inScope(pass.Pkg.Path()) {
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
				if _, ok := ts.Type.(*ast.InterfaceType); !ok {
					continue // не интерфейс — правило не применяется
				}
				if _, ok := roles[ts.Name.Name]; !ok {
					continue // имя не совпадает с голой ролью точно
				}
				pass.Reportf(ts.Name.Pos(),
					"%s: интерфейс %q именуется с префиксом сущности (например, HelloRepository)",
					ruleID, ts.Name.Name)
			}
		}
	}
	return nil, nil
}

// inScope сообщает, относится ли пакет к слою, где действует правило.
func inScope(path string) bool {
	return pathseg.Contains(path, "domain", "service") ||
		pathseg.Contains(path, "domain", "usecase") ||
		pathseg.Contains(path, "dal", "repository") ||
		pathseg.Contains(path, "server") ||
		pathseg.Contains(path, "event")
}
