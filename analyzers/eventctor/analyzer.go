// Package eventctor реализует правило GID-216 (event-ctor-deps): зависимости
// конструкторов event-слоя (источник: event.md).
//
//   - GID-216 (gideventctor): в consumer-слое конструктор обязан принимать
//     logger (*logrus.Logger или *logrus.Entry) — consumer собирает Entry с
//     полями broker/consumer; в producer-слое конструктор logger не принимает —
//     ошибки пробрасываются вызывающему коду.
//
// Scope определяется по сегментам import-пути: consumer — пакет с сегментами
// event и consumer, producer — с сегментами event и producer. Подпакеты с
// сегментами validate и convert исключаются: там валидаторы и конвертеры,
// а не consumer'ы/producer'ы.
//
// Конструктор — экспортируемая функция ^New[A-Z], возвращающая указатель на
// struct-тип, объявленный В ТОМ ЖЕ ПАКЕТЕ. Это автоматически исключает
// schema-функции вида New<X>Schema, возвращающие *registry.Schema чужого
// пакета.
//
// Исключения: имена конструкторов в settings.exclude (.golangci.yml) либо
// точечно //nolint:gideventctor.
package eventctor

import (
	"go/ast"
	"go/types"
	"regexp"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/lgr"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-216"

// Режимы проверки пакета event-слоя (тип scope объявлен ниже).
const (
	scopeNone scope = iota
	scopeConsumer
	scopeProducer
)

// ctorName — конструктор: экспортируемое имя вида New + заглавная буква.
var ctorName = regexp.MustCompile(`^New[A-Z]`)

// Analyzer — вариант с настройками по умолчанию (без исключений).
var Analyzer = NewAnalyzer(Settings{})

// Settings — настройки линтера из .golangci.yml.
type Settings struct {
	// Exclude — имена конструкторов-исключений (например "NewOrderConsumer").
	Exclude []string `json:"exclude"`
}

// NewAnalyzer строит анализатор GID-216 из настроек линтера (.golangci.yml).
func NewAnalyzer(cfg Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gideventctor",
		Doc:  ruleID + ": конструкторы consumer'ов принимают logrus-logger, конструкторы producer'ов — нет",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, cfg)
		},
	}
}

// scope — режим проверки для пакета event-слоя.
type scope int

func pkgScope(pkgPath string) scope {
	if !pathseg.Contains(pkgPath, "event") {
		return scopeNone
	}
	// validate/convert — это валидаторы и конвертеры, не consumer'ы/producer'ы.
	if pathseg.Contains(pkgPath, "validate") || pathseg.Contains(pkgPath, "convert") {
		return scopeNone
	}
	switch {
	case pathseg.Contains(pkgPath, "consumer"):
		return scopeConsumer
	case pathseg.Contains(pkgPath, "producer"):
		return scopeProducer
	default:
		return scopeNone
	}
}

func run(pass *analysis.Pass, cfg Settings) (any, error) {
	sc := pkgScope(pass.Pkg.Path())
	if sc == scopeNone {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !fn.Name.IsExported() {
				continue
			}
			if !ctorName.MatchString(fn.Name.Name) {
				continue
			}
			if exclude.Match(cfg.Exclude, "", fn.Name.Name) {
				continue
			}
			obj, ok := pass.TypesInfo.Defs[fn.Name].(*types.Func)
			if !ok {
				continue
			}
			sig, ok := obj.Type().(*types.Signature)
			if !ok {
				continue
			}
			if !returnsLocalStructPtr(pass.Pkg, sig) {
				continue
			}
			check(pass, sc, fn, sig)
		}
	}
	return nil, nil
}

func check(pass *analysis.Pass, sc scope, fn *ast.FuncDecl, sig *types.Signature) {
	has := hasLoggerParam(sig)
	switch sc {
	case scopeConsumer:
		if !has {
			pass.Reportf(fn.Name.Pos(),
				"%s: consumer принимает *logrus.Logger и собирает Entry с полями broker/consumer (см. event.md)",
				ruleID)
		}
	case scopeProducer:
		if has {
			pass.Reportf(fn.Name.Pos(),
				"%s: producer не принимает logger — ошибки пробрасываются вызывающему коду; "+
					"осознанное исключение — //nolint:gideventctor",
				ruleID)
		}
	}
}

// returnsLocalStructPtr сообщает, возвращает ли сигнатура (первым результатом)
// указатель на struct-тип, объявленный в текущем пакете.
func returnsLocalStructPtr(pkg *types.Package, sig *types.Signature) bool {
	results := sig.Results()
	if results.Len() == 0 {
		return false
	}
	first := results.At(0)
	ptr, ok := first.Type().(*types.Pointer)
	if !ok {
		return false
	}
	named, ok := types.Unalias(ptr.Elem()).(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	if obj.Pkg() != pkg {
		return false
	}
	_, ok = named.Underlying().(*types.Struct)
	return ok
}

// hasLoggerParam сообщает, есть ли среди параметров logrus-тип
// (*logrus.Logger или *logrus.Entry).
func hasLoggerParam(sig *types.Signature) bool {
	params := sig.Params()
	for i := 0; i < params.Len(); i++ {
		param := params.At(i)
		if lgr.IsType(param.Type()) {
			return true
		}
	}
	return false
}
