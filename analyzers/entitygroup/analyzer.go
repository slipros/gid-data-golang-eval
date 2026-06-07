// Package entitygroup реализует правило GID-157: код сущности — единый
// блок. Все функции сущности лежат в файле её объявления, в порядке
// type -> конструктор New<Entity> -> методы. Функции разных сущностей
// не перемешиваются в одну кучу.
package entitygroup

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-157"

const (
	kindType declKind = iota
	kindCtor
	kindMethod
)

// Analyzer — правило GID-157: код сущности — единый блок: type, конструктор, методы.
var Analyzer = &analysis.Analyzer{
	Name: "gidentitygroup",
	Doc:  ruleID + ": код сущности — единый блок: type, конструктор, методы; без перемешивания сущностей",
	Run:  run,
}

// ownedDecl — декларация, принадлежащая сущности.
type ownedDecl struct {
	entity string
	kind   declKind
	name   *ast.Ident
}

type declKind int

func run(pass *analysis.Pass) (any, error) {
	typeFile := structFiles(pass)
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		checkFile(pass, file, typeFile)
	}
	return nil, nil
}

func checkFile(pass *analysis.Pass, file *ast.File, typeFile map[string]*ast.File) {
	owned := ownedDecls(file, typeFile)

	typeIdx := map[string]int{}
	ctorIdx := map[string]int{}
	//nolint:gidallptr // плагин не зависит от внутренней библиотеки gdhelper
	for i, d := range owned {
		switch d.kind {
		case kindType:
			typeIdx[d.entity] = i
		case kindCtor:
			ctorIdx[d.entity] = i
		}
	}

	// Методы и конструктор — в файле объявления сущности.
	//nolint:gidallptr // плагин не зависит от внутренней библиотеки gdhelper
	for _, d := range owned {
		if d.kind == kindType {
			continue
		}
		declFile, ok := typeFile[d.entity]
		if ok && declFile != file {
			pass.Reportf(d.name.Pos(),
				"%s: %q принадлежит сущности %q — код сущности живёт в файле её объявления",
				ruleID, d.name.Name, d.entity)
		}
	}

	// Порядок внутри файла: type -> конструктор -> методы.
	//nolint:gidallptr // плагин не зависит от внутренней библиотеки gdhelper
	for i, d := range owned {
		ti, hasType := typeIdx[d.entity]
		switch d.kind {
		case kindCtor:
			if hasType && i < ti {
				pass.Reportf(d.name.Pos(),
					"%s: конструктор %q размещается под объявлением типа %q", ruleID, d.name.Name, d.entity)
			}
		case kindMethod:
			if hasType && i < ti {
				pass.Reportf(d.name.Pos(),
					"%s: метод %q размещается под объявлением типа %q", ruleID, d.name.Name, d.entity)
			}
			if ci, hasCtor := ctorIdx[d.entity]; hasCtor && i < ci {
				pass.Reportf(d.name.Pos(),
					"%s: метод %q размещается под конструктором New%s", ruleID, d.name.Name, d.entity)
			}
		}
	}

	// Перемешивание: блок сущности непрерывен.
	seen := map[string]struct{}{}
	last := ""
	//nolint:gidallptr // плагин не зависит от внутренней библиотеки gdhelper
	for _, d := range owned {
		if d.entity == last {
			continue
		}
		if _, ok := seen[d.entity]; ok {
			pass.Reportf(d.name.Pos(),
				"%s: код сущности %q перемешан с кодом других сущностей — блок сущности непрерывен",
				ruleID, d.entity)
		}
		seen[last] = struct{}{}
		last = d.entity
	}
}

// ownedDecls — последовательность деклараций файла с их сущностями.
func ownedDecls(file *ast.File, typeFile map[string]*ast.File) []ownedDecl {
	var out []ownedDecl
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if _, ok := ts.Type.(*ast.StructType); ok {
					out = append(out, ownedDecl{entity: ts.Name.Name, kind: kindType, name: ts.Name})
				}
			}
		case *ast.FuncDecl:
			if d.Recv != nil {
				if recv := recvTypeName(d); recv != "" {
					out = append(out, ownedDecl{entity: recv, kind: kindMethod, name: d.Name})
				}
				continue
			}
			if entity, ok := strings.CutPrefix(d.Name.Name, "New"); ok && entity != "" {
				if _, declared := typeFile[entity]; declared {
					out = append(out, ownedDecl{entity: entity, kind: kindCtor, name: d.Name})
				}
			}
		}
	}
	return out
}

// structFiles — файл объявления каждой структуры пакета.
func structFiles(pass *analysis.Pass) map[string]*ast.File {
	out := map[string]*ast.File{}
	for _, file := range pass.Files {
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
				if _, ok := ts.Type.(*ast.StructType); ok {
					out[ts.Name.Name] = file
				}
			}
		}
	}
	return out
}

func recvTypeName(fn *ast.FuncDecl) string {
	t := fn.Recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	if ident, ok := t.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}
