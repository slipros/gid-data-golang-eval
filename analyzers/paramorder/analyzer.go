// Package paramorder реализует канонический порядок параметров функций
// и методов: ctx -> opts -> logger -> остальные.
//
//   - GID-110: context.Context — всегда первый параметр;
//   - GID-113: opts (тип с постфиксом Options) — первый после ctx
//     (или первый, если ctx нет);
//   - GID-153: logger (logrus) идёт после opts, если opts существуют.
package paramorder

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/lgr"
)

const (
	kindOther paramKind = iota
	kindCtx
	kindOpts
	kindLogger
)

// Analyzer — правило GID: см. Doc.
var Analyzer = &analysis.Analyzer{
	Name: "gidparamorder",
	Doc:  "GID-110/113/153: порядок параметров — ctx, opts, logger, остальные",
	Run:  run,
}

type paramKind int

type param struct {
	kind paramKind
	pos  ast.Node
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Type.Params == nil {
				continue
			}
			checkParams(pass, flatten(pass, fn.Type.Params))
		}
	}
	return nil, nil
}

func flatten(pass *analysis.Pass, fields *ast.FieldList) []param {
	var params []param
	for _, field := range fields.List {
		kind := classify(pass.TypesInfo.TypeOf(field.Type))
		n := max(len(field.Names), 1)
		for range n {
			params = append(params, param{kind: kind, pos: field})
		}
	}
	return params
}

func checkParams(pass *analysis.Pass, params []param) {
	ctxIdx, optsIdx, loggerIdx := -1, -1, -1
	//nolint:gidallptr // плагин не зависит от внутренней библиотеки gdhelper
	for i, p := range params {
		switch {
		case p.kind == kindCtx && ctxIdx < 0:
			ctxIdx = i
		case p.kind == kindOpts && optsIdx < 0:
			optsIdx = i
		case p.kind == kindLogger && loggerIdx < 0:
			loggerIdx = i
		}
	}
	if ctxIdx > 0 {
		pass.Reportf(params[ctxIdx].pos.Pos(),
			"GID-110: context.Context должен быть первым параметром")
	}
	if optsIdx >= 0 {
		want := 0
		if ctxIdx == 0 {
			want = 1
		}
		if optsIdx != want {
			pass.Reportf(params[optsIdx].pos.Pos(),
				"GID-113: opts идёт первым параметром после ctx, не последним")
		}
	}
	if loggerIdx >= 0 && optsIdx >= 0 && loggerIdx < optsIdx {
		pass.Reportf(params[loggerIdx].pos.Pos(),
			"GID-153: logger идёт после opts сущности")
	}
}

func classify(t types.Type) paramKind {
	if t == nil {
		return kindOther
	}
	if lgr.IsType(t) {
		return kindLogger
	}
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return kindOther
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	if pkg != nil && pkg.Path() == "context" && obj.Name() == "Context" {
		return kindCtx
	}
	if strings.HasSuffix(obj.Name(), "Options") {
		return kindOpts
	}
	return kindOther
}
