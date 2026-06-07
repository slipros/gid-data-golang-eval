// Package validatorlib реализует правило GID-164: любые входящие данные
// (http- и grpc-запросы, kafka-события) валидируются через
// github.com/raoptimus/validator.go/v2.
//
//   - validate-пакеты (server/*/handler/validate, kafka/consumer/validate)
//     обязаны использовать validator.go;
//   - сторонние валидационные библиотеки запрещены везде.
//
// Исключения возможны:
//   - точечно: //nolint:gidvalidator (на строке package или импорта)
//   - централизованно: settings.exclude — пакеты, освобождённые от
//     требования (полный import-путь или суффикс сегментов).
package validatorlib

import (
	"go/ast"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const (
	ruleID       = "GID-164"
	validatorLib = "github.com/raoptimus/validator.go/v2"
)

// foreignValidators — сторонние валидационные библиотеки (по префиксу).
var foreignValidators = []string{
	"github.com/go-playground/validator",
	"github.com/go-ozzo/ozzo-validation",
	"github.com/asaskevich/govalidator",
	"gopkg.in/go-playground/validator",
}

// Analyzer — вариант без исключений.
var Analyzer = NewAnalyzer(Settings{})

// Settings — настройки линтера из .golangci.yml.
type Settings struct {
	// Exclude — validate-пакеты, освобождённые от требования:
	// полный import-путь или суффикс сегментов.
	Exclude []string `json:"exclude"`
}

// NewAnalyzer строит анализатор GID-164 из настроек линтера (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidvalidator",
		Doc:  ruleID + ": входящие данные валидируются через " + validatorLib,
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s)
		},
	}
}

func run(pass *analysis.Pass, s Settings) (any, error) {
	checkForeignValidators(pass)
	if pathseg.EndsWith(pass.Pkg.Path(), "validate") && !excludedPkg(pass.Pkg.Path(), s.Exclude) {
		checkValidatorUsed(pass)
	}
	return nil, nil
}

// checkForeignValidators: сторонние валидаторы запрещены в любом пакете.
func checkForeignValidators(pass *analysis.Pass) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, imp := range file.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			for _, lib := range foreignValidators {
				if path == lib || strings.HasPrefix(path, lib+"/") {
					pass.Reportf(imp.Pos(),
						"%s: сторонняя валидационная библиотека %q запрещена — используйте %s",
						ruleID, path, validatorLib)
				}
			}
		}
	}
}

// checkValidatorUsed: validate-пакет обязан импортировать validator.go.
func checkValidatorUsed(pass *analysis.Pass) {
	for _, file := range pass.Files {
		for _, imp := range file.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			if path == validatorLib || strings.HasPrefix(path, validatorLib+"/") {
				return // validator.go используется
			}
		}
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		pass.Reportf(file.Name.Pos(),
			"%s: validate-пакет %q обязан использовать %s (исключения: nolint или settings.exclude)",
			ruleID, pass.Pkg.Path(), validatorLib)
		return // одной диагностики на пакет достаточно
	}
}

func excludedPkg(pkgPath string, exclude []string) bool {
	for _, e := range exclude {
		if pkgPath == e || strings.HasSuffix(pkgPath, "/"+e) {
			return true
		}
	}
	return false
}
