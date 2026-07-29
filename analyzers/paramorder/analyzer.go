// Package paramorder implements the canonical order of function and method
// parameters: ctx -> opts -> logger -> metrics -> the rest.
//
//   - GID-110: context.Context is always the first parameter;
//   - GID-113: opts (a type with the Options suffix) comes first after ctx
//     (or first if there is no ctx);
//   - GID-153: the logger (logrus or slog — see internal/lgr) goes after opts
//     when opts exist;
//   - GID-252: in a constructor (a New* function) the known parameters keep
//     that exact relative order and all of them precede the entity's own
//     dependencies — ctx, opts, logger, metrics, then the rest. Any of them may
//     be absent; what is present is ordered. A metrics parameter is a named
//     type with the Metric/Metrics suffix (the shape GID-174 describes).
//     GID-110/113/153 already cover ctx and opts, so GID-252 reports only what
//     they do not: a logger or a metrics parameter buried among the
//     dependencies, and metrics placed before the logger.
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
	kindMetrics
)

// Analyzer — the GID rule: see Doc.
var Analyzer = &analysis.Analyzer{
	Name: "gidparamorder",
	Doc:  "GID-110/113/153/252: parameter order is ctx, opts, logger, metrics, then the rest. Fix: reorder parameters",
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
			params := flatten(pass, fn.Type.Params)
			checkParams(pass, params)
			if isConstructor(fn) {
				checkConstructorOrder(pass, params)
			}
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

// isConstructor reports whether fn is a constructor — a package-level New*
// function. GID-252 is scoped to those: an ordinary function may legitimately
// take a logger last (a helper enriching what the caller passes in).
func isConstructor(fn *ast.FuncDecl) bool {
	return fn.Recv == nil && strings.HasPrefix(fn.Name.Name, "New")
}

// checkConstructorOrder reports the part of the canonical order that
// GID-110/113/153 do not cover: a logger or a metrics parameter standing after
// the entity's own dependencies, and metrics standing before the logger.
func checkConstructorOrder(pass *analysis.Pass, params []param) {
	firstOther, loggerIdx, metricsIdx := -1, -1, -1
	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
	for i, p := range params {
		switch {
		case p.kind == kindOther && firstOther < 0:
			firstOther = i
		case p.kind == kindLogger && loggerIdx < 0:
			loggerIdx = i
		case p.kind == kindMetrics && metricsIdx < 0:
			metricsIdx = i
		}
	}
	if firstOther >= 0 && loggerIdx > firstOther {
		pass.Reportf(params[loggerIdx].pos.Pos(),
			"GID-252: the logger comes after the entity's own dependencies. "+
				"Fix: order the parameters ctx, opts, logger, metrics, then the rest")
	}
	if firstOther >= 0 && metricsIdx > firstOther {
		pass.Reportf(params[metricsIdx].pos.Pos(),
			"GID-252: metrics come after the entity's own dependencies. "+
				"Fix: order the parameters ctx, opts, logger, metrics, then the rest")
	}
	if loggerIdx >= 0 && metricsIdx >= 0 && metricsIdx < loggerIdx {
		pass.Reportf(params[metricsIdx].pos.Pos(),
			"GID-252: metrics come before the logger. "+
				"Fix: order the parameters ctx, opts, logger, metrics, then the rest")
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
	if strings.HasSuffix(obj.Name(), "Metrics") || strings.HasSuffix(obj.Name(), "Metric") {
		return kindMetrics
	}
	return kindOther
}
