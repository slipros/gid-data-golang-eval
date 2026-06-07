// Package ifaceplace реализует правило GID-134 (interface-near-consumer):
// интерфейсы живут там, где используются.
//
// Проверка: если в полях структуры или в параметрах/результатах функции
// (метода) пакета используется именованный interface-тип, смотрим пакет,
// в котором этот интерфейс объявлен:
//
//   - тот же пакет — ОК (интерфейс определён рядом с потребителем);
//   - stdlib или внешняя библиотека — ОК. «Свой» пакет сервиса отличаем
//     от библиотеки по сегментам пути (pathseg): путь содержит слой-сегмент
//     (dal, domain, client, server, event, app, metric) — это наш пакет;
//     иначе — библиотека;
//   - интерфейс из model-слоя (/domain/model, включая подпакеты) — ОК, но
//     только если потребитель в слое /domain/service или /domain/usecase;
//     для остальных потребителей это нарушение;
//   - любой другой «свой» пакет — нарушение.
//
// Не задеваются: анонимные интерфейсы, error, any/interface{},
// generic-констрейнты. Сгенерированный код пропускается.
//
// LoadMode: нужен TypesInfo — определяем types.Interface и пакет
// объявления через Named.Obj().Pkg().
package ifaceplace

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-134"

// layerSegments — сегменты пути, по которым пакет опознаётся как «свой»
// слой сервиса (а не stdlib/внешняя библиотека).
var layerSegments = []string{
	"dal", "domain", "client", "server", "event", "app", "metric",
}

// Analyzer — правило GID-134: interfaces live where they are used; .
var Analyzer = &analysis.Analyzer{
	Name: "gidifaceplace",
	Doc: ruleID + ": interfaces live where they are used; " +
		"define the interface next to its consumer (exceptions: libraries and /domain/model for service/usecase)",
	Run: run,
}

func run(pass *analysis.Pass) (any, error) {
	consumerPkg := pass.Pkg
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				checkTypeDecl(pass, consumerPkg, d)
			case *ast.FuncDecl:
				checkFuncDecl(pass, consumerPkg, d)
			}
		}
	}
	return nil, nil
}

// checkTypeDecl проверяет поля struct-типов в объявлении типов.
func checkTypeDecl(pass *analysis.Pass, consumer *types.Package, gd *ast.GenDecl) {
	for _, spec := range gd.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok || st.Fields == nil {
			continue
		}
		for _, field := range st.Fields.List {
			checkExpr(pass, consumer, field.Type)
		}
	}
}

// checkFuncDecl проверяет параметры и результаты функции/метода.
func checkFuncDecl(pass *analysis.Pass, consumer *types.Package, fn *ast.FuncDecl) {
	if fn.Type == nil {
		return
	}
	checkFieldList(pass, consumer, fn.Type.Params)
	checkFieldList(pass, consumer, fn.Type.Results)
}

func checkFieldList(pass *analysis.Pass, consumer *types.Package, fl *ast.FieldList) {
	if fl == nil {
		return
	}
	for _, field := range fl.List {
		checkExpr(pass, consumer, field.Type)
	}
}

// checkExpr рассматривает типовое выражение позиции (поле/параметр/результат).
// Флагается только именованный interface-тип, объявленный в чужом «своём»
// пакете. Анонимные интерфейсы (ast.InterfaceType) сюда не попадают — у них
// нет *types.Named, значит и пакета объявления.
func checkExpr(pass *analysis.Pass, consumer *types.Package, expr ast.Expr) {
	tv, ok := pass.TypesInfo.Types[expr]
	if !ok {
		return
	}
	named, ok := tv.Type.(*types.Named)
	if !ok {
		return // анонимный интерфейс, базовый тип, не-именованный — пропуск
	}
	obj := named.Obj()
	if obj == nil {
		return
	}
	// error и прочие builtin-именованные типы: пакета нет.
	declPkg := obj.Pkg()
	if declPkg == nil {
		return
	}
	// Интересует только интерфейс.
	if _, isIface := named.Underlying().(*types.Interface); !isIface {
		return
	}

	declPath := declPkg.Path()
	// Тот же пакет — интерфейс определён рядом с потребителем.
	if declPkg == consumer {
		return
	}
	// Библиотека (stdlib / внешний модуль) — путь не содержит слой-сегментов.
	if !isOwnPackage(declPath) {
		return
	}
	// model-слой: разрешён только потребителям service/usecase.
	if pathseg.Contains(declPath, "domain", "model") && isServiceOrUsecase(consumer.Path()) {
		return
	}
	// Чужой «свой» пакет (или model-слой у не service/usecase) — нарушение.
	pass.Reportf(expr.Pos(),
		"%s: interface %s is declared in %s. Fix: define the interface next to its consumer "+
			"(exceptions: libraries and /domain/model for service/usecase)",
		ruleID, obj.Name(), declPath)
}

// isOwnPackage сообщает, что пакет — наш слой сервиса (а не библиотека):
// путь содержит хотя бы один слой-сегмент.
func isOwnPackage(path string) bool {
	for _, seg := range layerSegments {
		if pathseg.Contains(path, seg) {
			return true
		}
	}
	return false
}

// isServiceOrUsecase сообщает, что потребитель — слой domain/service
// или domain/usecase.
func isServiceOrUsecase(path string) bool {
	return pathseg.Contains(path, "domain", "service") ||
		pathseg.Contains(path, "domain", "usecase")
}
