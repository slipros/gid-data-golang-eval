// Package validatorshape реализует правило GID-213 (validator-shape):
// форма валидатора. Валидатор — структура с методом
// Validate(ctx context.Context, req *T) error, имя которой совпадает с
// именем операции.
//
// Scope: пакеты со слоем validate (segment "validate" в import-пути).
// Каждый ЭКСПОРТИРУЕМЫЙ struct-тип (кроме имён с суффиксом Options) обязан
// иметь метод Validate (pointer- или value-receiver), у которого:
//   - первый параметр имеет тип context.Context;
//   - единственный результат имеет тип error.
//
// Достаточно проверить первый параметр (ctx) и единственный результат
// (error): тип запроса req может быть любым, число параметров после ctx не
// ограничивается (см. validatorshape.feature, граничный класс).
//
// Сгенерированный код (ast.IsGenerated) пропускается. LoadMode — TypesInfo:
// нужны go/types, чтобы распознать context.Context, error и методы типа.
//
// Исключения:
//   - точечно: //nolint:gidvalidatorshape
//   - централизованно: settings.exclude в .golangci.yml — имена типов,
//     которые не считаются валидаторами (например "HealthCheck").
//
// Источник: validator.md.
package validatorshape

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-213"

// Analyzer — вариант с настройками по умолчанию (без исключений).
var Analyzer = NewAnalyzer(Settings{})

// Settings — настройки линтера из .golangci.yml.
type Settings struct {
	// Exclude — имена struct-типов, которые не считаются валидаторами
	// и не обязаны иметь метод Validate.
	Exclude []string `json:"exclude"`
}

// NewAnalyzer строит анализатор GID-213 из настроек линтера (.golangci.yml).
func NewAnalyzer(cfg Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidvalidatorshape",
		Doc:  ruleID + ": a validator is a struct with a Validate(ctx context.Context, req *T) error method. Fix: add that method",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, cfg)
		},
	}
}

func run(pass *analysis.Pass, cfg Settings) (any, error) {
	// Scope: только пакеты слоя validate.
	if !pathseg.Contains(pass.Pkg.Path(), "validate") {
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
				// Только экспортируемые struct-типы.
				if !ts.Name.IsExported() {
					continue
				}
				if _, ok := ts.Type.(*ast.StructType); !ok {
					continue
				}
				// Типы-настройки (*Options) не валидаторы.
				if strings.HasSuffix(ts.Name.Name, "Options") {
					continue
				}
				// Централизованные исключения по имени типа.
				if exclude.Match(cfg.Exclude, ts.Name.Name, ts.Name.Name) {
					continue
				}
				checkValidator(pass, ts)
			}
		}
	}
	return nil, nil
}

func checkValidator(pass *analysis.Pass, ts *ast.TypeSpec) {
	obj, ok := pass.TypesInfo.Defs[ts.Name].(*types.TypeName)
	if !ok {
		return
	}
	named, ok := obj.Type().(*types.Named)
	if !ok {
		return
	}
	if hasValidate(named) {
		return
	}
	pass.Reportf(ts.Name.Pos(),
		"%s: validator %q must have a Validate(ctx context.Context, req *T) error method. Fix: add it",
		ruleID, ts.Name.Name)
}

// hasValidate сообщает, есть ли у типа (или у указателя на него) метод
// Validate с первым параметром context.Context и единственным результатом
// error.
func hasValidate(named *types.Named) bool {
	// Метод ищем и на T, и на *T: pointer-receiver не попадает в methodset T.
	mset := types.NewMethodSet(types.NewPointer(named))
	for i := 0; i < mset.Len(); i++ {
		sel := mset.At(i)
		fn, ok := sel.Obj().(*types.Func)
		if !ok || fn.Name() != "Validate" {
			continue
		}
		sig, ok := fn.Type().(*types.Signature)
		if !ok {
			continue
		}
		return validateShape(sig)
	}
	return false
}

func validateShape(sig *types.Signature) bool {
	// Первый параметр — context.Context.
	params := sig.Params()
	if params.Len() < 1 {
		return false
	}
	first := params.At(0)
	if !isContext(first.Type()) {
		return false
	}
	// Единственный результат — error.
	results := sig.Results()
	if results.Len() != 1 {
		return false
	}
	res := results.At(0)
	return isError(res.Type())
}

func isContext(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	return pkg != nil && pkg.Path() == "context" && obj.Name() == "Context"
}

func isError(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj.Pkg() == nil && obj.Name() == "error"
}
