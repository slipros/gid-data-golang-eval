// Package convnaming реализует правила конвертеров:
//
//   - GID-105: имена конвертеров в convert-пакетах — <Dst><Type>From<Src>
//     (EntityCreateSnapshotFromModel, ModelHelloOutFromEntity);
//   - GID-135: функции-конвертеры (паттерн ...From...) живут в
//     convert/-подпакете своего слоя, не в самих service/repo/handler.
//
// Исключение: ctx-helper'ы <Name>FromContext (GID-166) — не конвертеры.
package convnaming

import (
	"go/ast"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const (
	ruleNaming = "GID-105"
	rulePlace  = "GID-135"
)

// converterName: <Dst><Type>From<Src> — слова до и после From.
var converterName = regexp.MustCompile(`^[A-Z][A-Za-z0-9]*From[A-Z][A-Za-z0-9]*$`)

// scopes — слои, в которых конвертеры обязаны жить в convert/.
var scopes = [][]string{
	{"dal"},
	{"domain"},
	{"server"},
	{"event"},
}

// Analyzer — правило GID: см. Doc.
var Analyzer = &analysis.Analyzer{
	Name: "gidconvnaming",
	Doc:  ruleNaming + "/" + rulePlace + ": конвертеры — <Dst><Type>From<Src> в convert/-пакетах",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	inConvert := pathseg.EndsWith(pass.Pkg.Path(), "convert")
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !fn.Name.IsExported() {
				continue
			}
			if inConvert {
				checkConverterName(pass, fn)
				continue
			}
			checkConverterPlace(pass, fn)
		}
	}
	return nil, nil
}

// checkConverterName — GID-105: экспортируемые функции convert-пакета
// именуются <Dst><Type>From<Src>.
func checkConverterName(pass *analysis.Pass, fn *ast.FuncDecl) {
	if converterName.MatchString(fn.Name.Name) {
		return
	}
	pass.Reportf(fn.Name.Pos(),
		"%s: конвертер %q именуется <Dst><Type>From<Src> (например, EntityCreateSnapshotFromModel)",
		ruleNaming, fn.Name.Name)
}

// checkConverterPlace — GID-135: функция-конвертер вне convert-пакета.
func checkConverterPlace(pass *analysis.Pass, fn *ast.FuncDecl) {
	name := fn.Name.Name
	if !converterName.MatchString(name) || strings.HasSuffix(name, "FromContext") {
		return
	}
	if !inScope(pass.Pkg.Path()) {
		return
	}
	pass.Reportf(fn.Name.Pos(),
		"%s: конвертер %q живёт в convert/-подпакете своего слоя", rulePlace, name)
}

func inScope(pkgPath string) bool {
	for _, scope := range scopes {
		if pathseg.Contains(pkgPath, scope...) {
			return true
		}
	}
	return false
}
