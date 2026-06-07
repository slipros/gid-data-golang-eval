// Package servicesingle реализует правило GID-148: domain-сервис посвящён
// одной сущности и не зависит от других сервисов. Оркестрация бизнес-логики
// нескольких сущностей — задача usecase, который может использовать
// несколько сервисов.
//
// Детерминированная проверка: в корне /domain/service поле структуры,
// чей тип — другая структура из того же пакета (кроме *Options), означает
// зависимость сервиса от сервиса.
package servicesingle

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-148"

// Analyzer — правило GID-148: сервис не зависит от другого сервиса — оркестрация сущностей выполняется в usecase.
var Analyzer = &analysis.Analyzer{
	Name: "gidservicesingle",
	Doc:  ruleID + ": сервис не зависит от другого сервиса — оркестрация сущностей выполняется в usecase",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if !pathseg.EndsWith(pass.Pkg.Path(), "domain", "service") {
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
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				checkServiceStruct(pass, ts.Name.Name, st)
			}
		}
	}
	return nil, nil
}

func checkServiceStruct(pass *analysis.Pass, owner string, st *ast.StructType) {
	for _, field := range st.Fields.List {
		dep, ok := samePackageStruct(pass, field.Type)
		if !ok {
			continue
		}
		pass.Reportf(field.Pos(),
			"%s: сервис %q зависит от сервиса %q — сервис посвящён одной сущности, "+
				"оркестрация нескольких сервисов выполняется в usecase",
			ruleID, owner, dep)
	}
}

// samePackageStruct возвращает имя типа, если тип поля — структура
// (или указатель на структуру) из этого же пакета и не Options-тип.
func samePackageStruct(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	t := pass.TypesInfo.TypeOf(expr)
	if t == nil {
		return "", false
	}
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return "", false
	}
	namedObj := named.Obj()
	if namedObj.Pkg() != pass.Pkg {
		return "", false
	}
	if _, ok := named.Underlying().(*types.Struct); !ok {
		return "", false
	}
	name := namedObj.Name()
	if strings.HasSuffix(name, "Options") {
		return "", false
	}
	return name, true
}
