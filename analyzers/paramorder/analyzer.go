// Package paramorder implements the canonical order of function and method
// parameters: ctx -> opts -> logger -> the rest.
//
//   - GID-110: context.Context is always the first parameter;
//   - GID-113: opts (a type with the Options suffix) comes first after ctx
//     (or first if there is no ctx);
//   - GID-153: logger (logrus) goes after opts when opts exist.
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

// Analyzer — the GID rule: see Doc.
var Analyzer = &analysis.Analyzer{
	Name: "gidparamorder",
	Doc:  "GID-110/113/153: parameter order is ctx, opts, logger, then the rest. Fix: reorder parameters",
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
	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
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
			"GID-110: context.Context must be the first parameter. Fix: move ctx first")
	}
	if optsIdx >= 0 {
		want := 0
		if ctxIdx == 0 {
			want = 1
		}
		if optsIdx != want {
			pass.Reportf(params[optsIdx].pos.Pos(),
				"GID-113: opts must come right after ctx, not last. Fix: move opts after ctx")
		}
	}
	if loggerIdx >= 0 && optsIdx >= 0 && loggerIdx < optsIdx {
		pass.Reportf(params[loggerIdx].pos.Pos(),
			"GID-153: logger must come after the entity opts. Fix: move logger after opts")
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
