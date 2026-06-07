// Package optsnaming реализует правило GID-126: конвенции имени и дефолтов
// Options-паттерна.
//
//   - struct-тип с именем РОВНО Options вне app-слоя — нарушение: тип
//     настроек именуется с префиксом сущности (JobOptions), не голым Options.
//     В app-слое голый Options — норма (композиция GRPCOptions/KafkaOptions).
//   - package-level var типа <X>Options (включая указатель), имя которой не
//     начинается с Default — нарушение: дефолты живут в переменной Default<X>Options.
//     Var в app-слое тоже проверяется (дефолты и там Default*).
package optsnaming

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-126"

// Analyzer — правило GID-126: имя Options-типа с префиксом сущности,
// дефолты — переменная Default<X>Options.
var Analyzer = &analysis.Analyzer{
	Name: "gidoptsnaming",
	Doc:  ruleID + ": тип настроек — с префиксом сущности; дефолты — переменная Default<X>Options",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	inApp := pathseg.Contains(pass.Pkg.Path(), "app")
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			switch gd.Tok {
			case token.TYPE:
				if !inApp {
					checkTypeNames(pass, gd)
				}
			case token.VAR:
				checkDefaultNames(pass, gd)
			}
		}
	}
	return nil, nil
}

// checkTypeNames: struct-тип с именем ровно Options вне app-слоя — нарушение.
// Не-struct типы (alias на Options, interface) не задеваются.
func checkTypeNames(pass *analysis.Pass, gd *ast.GenDecl) {
	for _, spec := range gd.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		if ts.Name.Name != "Options" {
			continue
		}
		if ts.Assign.IsValid() {
			continue // alias (type Options = X) — не задеваем
		}
		if _, ok := ts.Type.(*ast.StructType); !ok {
			continue // только struct-типы
		}
		pass.Reportf(ts.Name.Pos(),
			"%s: тип настроек — с префиксом сущности: JobOptions, не голый Options", ruleID)
	}
}

// checkDefaultNames: package-level var типа <X>Options (включая указатель),
// имя которой не начинается с Default — нарушение. Локальные переменные сюда
// не попадают (проверяются только top-level GenDecl с Tok==var).
func checkDefaultNames(pass *analysis.Pass, gd *ast.GenDecl) {
	for _, spec := range gd.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range vs.Names {
			if name.Name == "_" {
				continue
			}
			obj := pass.TypesInfo.Defs[name]
			if obj == nil {
				continue
			}
			if !isOptionsType(obj.Type()) {
				continue
			}
			if strings.HasPrefix(name.Name, "Default") {
				continue
			}
			pass.Reportf(name.Pos(),
				"%s: дефолты Options — переменная Default<X>Options", ruleID)
		}
	}
}

// isOptionsType сообщает, является ли тип именованным <X>Options
// (с префиксом сущности) — по значению или по указателю. Голый Options
// без префикса не считается (это сам тип настроек, не его дефолт).
func isOptionsType(t types.Type) bool {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	name := obj.Name()
	return strings.HasSuffix(name, "Options") && name != "Options"
}
