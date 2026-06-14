// Package validatorlib implements rule GID-164: all incoming data
// (http and grpc requests, kafka events) is validated via
// github.com/raoptimus/validator.go/v2.
//
//   - validate packages (server/*/handler/validate, kafka/consumer/validate)
//     must use validator.go;
//   - third-party validation libraries are forbidden everywhere.
//
// Exceptions are possible:
//   - targeted: //nolint:gidvalidator (on the package or import line)
//   - centralized: settings.exclude — packages exempt from the requirement
//     (a full import path or a segment suffix).
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

// foreignValidators — third-party validation libraries (by prefix).
var foreignValidators = []string{
	"github.com/go-playground/validator",
	"github.com/go-ozzo/ozzo-validation",
	"github.com/asaskevich/govalidator",
	"gopkg.in/go-playground/validator",
}

// Analyzer — the variant without exclusions.
var Analyzer = NewAnalyzer(Settings{})

// Settings — the linter settings from .golangci.yml.
type Settings struct {
	// Exclude — validate packages exempt from the requirement:
	// a full import path or a segment suffix.
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-164 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidvalidator",
		Doc:  ruleID + ": incoming data is validated via " + validatorLib,
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

// checkForeignValidators: third-party validators are forbidden in any package.
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
						"%s: third-party validation library %q is forbidden. Fix: use %s",
						ruleID, path, validatorLib)
				}
			}
		}
	}
}

// checkValidatorUsed: a validate package must import validator.go.
func checkValidatorUsed(pass *analysis.Pass) {
	for _, file := range pass.Files {
		for _, imp := range file.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			if path == validatorLib || strings.HasPrefix(path, validatorLib+"/") {
				return // validator.go is used
			}
		}
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		pass.Reportf(file.Name.Pos(),
			"%s: validate package %q must use %s. Fix: import it (exceptions: nolint or settings.exclude)",
			ruleID, pass.Pkg.Path(), validatorLib)
		return // one diagnostic per package is enough
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
