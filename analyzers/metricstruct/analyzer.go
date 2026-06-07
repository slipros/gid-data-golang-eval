// Package metricstruct реализует правило GID-174 (gidmetricstruct):
// пакет метрик сервиса стандартизирован.
//
// Конвенция всех backend-go сервисов: путь пакета — /metric, имя пакета —
// metric, а метрики агрегируются в экспортируемом struct-агрегаторе
// Prometheus (поля — метрики по протоколам/подсистемам: HTTP, GRPC, Kafka,
// …) с методом Register.
//
// Правило применимо только к пакету, чей import-путь оканчивается сегментом
// metric или metrics (pathseg.EndsWith). Прочие пакеты не трогаем.
//
// Проверки:
//  1. путь оканчивается metrics → пакет назван неверно;
//  2. путь оканчивается metric, но типа Prometheus нет;
//  3. Prometheus есть, но без метода Register;
//  4. Prometheus объявлен, но не struct.
//
// Конвенция группировки (доп. проверки, только для пути .../metric):
//   - доп. метрики живут в отдельных файлах, группируясь функционально в
//     структурах (по одной группе на файл);
//   - prometheus.go занимается wiring'ом: тип Prometheus объявлен именно в
//     prometheus.go, его метод Register регистрирует группы — вызывает их
//     метод Register.
//
// Доп. проверки:
//  5. тип Prometheus объявлен не в prometheus.go;
//  6. в prometheus.go объявлены другие экспортируемые struct-типы;
//  7. в файле пакета metric (не prometheus.go) объявлено ≥2 экспортируемых
//     struct-типов — репорт на втором и последующих;
//  8. поле Prometheus, чей тип имеет метод Register, не зарегистрировано в
//     теле Prometheus.Register (нет вызова <поле>.Register(...)).
package metricstruct

import (
	"go/ast"
	"go/types"
	"path/filepath"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const (
	ruleID     = "GID-174"
	typeName   = "Prometheus"
	regMethod  = "Register"
	wiringFile = "prometheus.go"
)

// Analyzer — правило GID-174. Требует информацию о типах: метод Register
// определяется через types (учитывая value/pointer receiver).
var Analyzer = &analysis.Analyzer{
	Name: "gidmetricstruct",
	Doc:  ruleID + ": the metrics package is standardized: path/name metric, a Prometheus struct with a Register method. Fix: follow that layout",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	path := pass.Pkg.Path()

	// Неприменимость: пакет не является корнем metric/metrics-слоя.
	switch {
	case pathseg.EndsWith(path, "metrics"):
		// Проверка 1: пакет назван metrics вместо metric.
		reportOnPackageClause(pass,
			"%s: the metrics package must be named metric, not metrics. Fix: rename it to metric", ruleID)
		return nil, nil
	case pathseg.EndsWith(path, "metric"):
		// продолжаем проверки 2-4
	default:
		return nil, nil
	}

	ts, named := findPrometheus(pass)
	if named == nil {
		// Проверка 2: типа Prometheus нет.
		reportOnPackageClause(pass,
			"%s: the metric package must declare a metrics aggregator: struct %s with a %s method. Fix: add it",
			ruleID, typeName, regMethod)
		return nil, nil
	}

	// Проверка 4: Prometheus есть, но не struct.
	st, ok := named.Underlying().(*types.Struct)
	if !ok {
		pass.Reportf(ts.Name.Pos(),
			"%s: %s must be a metrics aggregator struct. Fix: make it a struct", ruleID, typeName)
		return nil, nil
	}

	// Проверка 5: Prometheus объявлен не в prometheus.go.
	if filepath.Base(pass.Fset.Position(ts.Name.Pos()).Filename) != wiringFile {
		pass.Reportf(ts.Name.Pos(),
			"%s: the %s aggregator must live in %s. Fix: move it there", ruleID, typeName, wiringFile)
	}

	// Проверки 6 и 7: группировка struct-типов по файлам.
	checkGrouping(pass)

	// Проверка 3: struct Prometheus без метода Register.
	if !hasRegisterMethod(named) {
		pass.Reportf(ts.Name.Pos(),
			"%s: struct %s must have a %s method. Fix: add it", ruleID, typeName, regMethod)
		return nil, nil
	}

	// Проверка 8: каждая группа-поле зарегистрирована в Prometheus.Register.
	checkRegisterWiring(pass, named, st)

	return nil, nil
}

// checkGrouping реализует проверки 6 и 7:
//   - 6: в prometheus.go не должно быть других экспортируемых struct-типов
//     (кроме Prometheus);
//   - 7: в прочих файлах пакета — не более одной экспортируемой struct-группы
//     на файл (репорт на второй и последующих).
func checkGrouping(pass *analysis.Pass) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		fname := filepath.Base(pass.Fset.Position(file.Pos()).Filename)
		if isTestFile(fname) {
			continue
		}
		isWiring := fname == wiringFile
		groupsInFile := 0
		for _, ts := range exportedStructTypes(file) {
			if isWiring {
				if ts.Name.Name == typeName {
					continue // сам агрегатор в prometheus.go — ок
				}
				pass.Reportf(ts.Name.Pos(),
					"%s: a metrics group must live in its own file; %s is wiring only. Fix: move the group out",
					ruleID, wiringFile)
				continue
			}
			groupsInFile++
			if groupsInFile >= 2 {
				pass.Reportf(ts.Name.Pos(),
					"%s: one functional metrics group per file. Fix: split groups into separate files", ruleID)
			}
		}
	}
}

// checkRegisterWiring реализует проверку 8: поле Prometheus, чей тип (или
// указатель на него) имеет метод Register, обязано быть зарегистрировано
// внутри тела Prometheus.Register вызовом <поле>.Register(...).
func checkRegisterWiring(pass *analysis.Pass, named *types.Named, st *types.Struct) {
	body, recv := registerMethodBody(pass, named)
	if body == nil {
		return // метод Register объявлен в другом пакете/без тела — не репортим
	}
	called := registeredFields(body, recv)

	// Детерминированный порядок: по индексу полей структуры.
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		if !fieldTypeHasRegister(f.Type()) {
			continue
		}
		if _, ok := called[f.Name()]; ok {
			continue
		}
		pass.Reportf(f.Pos(),
			"%s: %s.%s registers group %s. Fix: call its %s",
			ruleID, typeName, regMethod, f.Name(), regMethod)
	}
}

// findPrometheus ищет объявление типа Prometheus в пакете и его *types.Named.
func findPrometheus(pass *analysis.Pass) (*ast.TypeSpec, *types.Named) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.Name.Name != typeName {
					continue
				}
				obj, ok := pass.TypesInfo.Defs[ts.Name].(*types.TypeName)
				if !ok {
					continue
				}
				named, ok := obj.Type().(*types.Named)
				if !ok {
					continue
				}
				return ts, named
			}
		}
	}
	return nil, nil
}

// hasRegisterMethod сообщает, есть ли у типа метод Register (любая сигнатура,
// value или pointer receiver). Pointer-набор методов покрывает оба случая.
func hasRegisterMethod(named *types.Named) bool {
	return lookupRegister(named) != nil
}

// lookupRegister возвращает метод Register типа (через pointer-набор) либо nil.
func lookupRegister(named *types.Named) *types.Func {
	mset := types.NewMethodSet(types.NewPointer(named))
	obj := named.Obj()
	sel := mset.Lookup(obj.Pkg(), regMethod)
	if sel == nil {
		return nil
	}
	fn, ok := sel.Obj().(*types.Func)
	if !ok {
		return nil
	}
	return fn
}

// fieldTypeHasRegister сообщает, имеет ли тип поля (или указатель на него)
// метод Register. Тип поля приводится к *types.Named.
func fieldTypeHasRegister(t types.Type) bool {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	return hasRegisterMethod(named)
}

// exportedStructTypes возвращает TypeSpec'и экспортируемых struct-типов файла
// в порядке объявления (детерминированно).
func exportedStructTypes(file *ast.File) []*ast.TypeSpec {
	var out []*ast.TypeSpec
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || !ts.Name.IsExported() {
				continue
			}
			if _, ok := ts.Type.(*ast.StructType); !ok {
				continue
			}
			out = append(out, ts)
		}
	}
	return out
}

// registerMethodBody находит AST-тело метода Register у Prometheus и имя его
// ресивера (для распознавания вызовов вида p.Field.Register(...)). Возвращает
// (nil, "") если метод не найден среди файлов пакета или у него нет тела.
func registerMethodBody(pass *analysis.Pass, named *types.Named) (body *ast.BlockStmt, recv string) {
	fn := lookupRegister(named)
	if fn == nil {
		return nil, ""
	}
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Name.Name != regMethod || fd.Recv == nil || fd.Body == nil {
				continue
			}
			obj, ok := pass.TypesInfo.Defs[fd.Name].(*types.Func)
			if !ok || obj != fn {
				continue
			}
			recv := ""
			if len(fd.Recv.List) > 0 && len(fd.Recv.List[0].Names) > 0 {
				recv = fd.Recv.List[0].Names[0].Name
			}
			return fd.Body, recv
		}
	}
	return nil, ""
}

// registeredFields собирает имена полей, для которых в теле Register есть вызов
// <field>.Register(...). Распознаются формы p.Field.Register(...) (через
// ресивер recv) и Field.Register(...) (напрямую).
func registeredFields(body *ast.BlockStmt, recv string) map[string]struct{} {
	out := map[string]struct{}{}
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		// верхний селектор: X.Register
		topSel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || topSel.Sel.Name != regMethod {
			return true
		}
		switch x := topSel.X.(type) {
		case *ast.SelectorExpr: // p.Field.Register(...)
			if id, ok := x.X.(*ast.Ident); ok && (recv == "" || id.Name == recv) {
				out[x.Sel.Name] = struct{}{}
			}
		case *ast.Ident: // Field.Register(...) напрямую
			out[x.Name] = struct{}{}
		}
		return true
	})
	return out
}

// isTestFile сообщает, является ли файл _test.go.
func isTestFile(name string) bool {
	const suffix = "_test.go"
	return len(name) > len(suffix) && name[len(name)-len(suffix):] == suffix
}

// reportOnPackageClause репортит на package clause не-generated файла с
// наименьшим именем — детерминированно вне зависимости от порядка pass.Files.
func reportOnPackageClause(pass *analysis.Pass, format string, args ...any) {
	var target *ast.File
	var targetName string
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		tokenFile := pass.Fset.File(file.Pos())
		name := tokenFile.Name()
		if target == nil || name < targetName {
			target, targetName = file, name
		}
	}
	if target == nil {
		return
	}
	pass.Reportf(target.Name.Pos(), format, args...)
}
