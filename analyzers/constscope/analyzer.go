// Package constscope implements rule GID-194: constants are declared where
// they are directly used — package-level constants outside the
// model/entity layers are forbidden.
//
// Allowed: a const inside a function; a package-level const in /domain/model/**
// and /dal/entity/** (the home of shared constants); an unexported package-level
// const used by several functions of the package or by a package-level
// declaration (var, type, another const). Violations: an exported const
// outside model/entity and an unexported const used by exactly one
// function — its place is inside that function.
package constscope

import (
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-194"

// allowedScopes — the layers where package-level constants are legal,
// including subpackages (model/filter, entity/filter, etc.).
var allowedScopes = [][]string{
	{"domain", "model"},
	{"dal", "entity"},
}

// Analyzer — rule GID-194 with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings — settings of rule GID-194 from .golangci.yml.
type Settings struct {
	// Exclude — names of constants the rule skips.
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-194 analyzer with the given exclusions.
func NewAnalyzer(s Settings) *analysis.Analyzer {
	excluded := make(map[string]struct{}, len(s.Exclude))
	for _, name := range s.Exclude {
		excluded[name] = struct{}{}
	}
	return &analysis.Analyzer{
		Name: "gidconstscope",
		Doc:  ruleID + ": constants are declared where they are used; package-level const outside model/entity is forbidden. Fix: move it into the using function, or into /domain/model or /dal/entity if shared",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, excluded)
		},
	}
}

// constGroup — one package-level const block.
type constGroup struct {
	decl *ast.GenDecl
	// names — unexported candidates for localization.
	names []*ast.Ident
	// grouped — the values are tied via iota: the block can only be moved as a whole.
	grouped bool
	// skipLocal — the block contains an exported or excluded name,
	// so we do not suggest localizing the iota block.
	skipLocal bool
}

func run(pass *analysis.Pass, excluded map[string]struct{}) (any, error) {
	if inAllowedScope(pass.Pkg.Path()) {
		return nil, nil
	}
	groups := collectGroups(pass, excluded)
	if len(groups) == 0 {
		return nil, nil
	}
	usage := collectUsage(pass, candidateObjects(pass, groups))
	for _, g := range groups {
		reportGroup(pass, g, usage)
	}
	return nil, nil
}

// collectGroups collects package-level const blocks and immediately reports
// exported constants — shared constants live in model/entity.
func collectGroups(pass *analysis.Pass, excluded map[string]struct{}) []*constGroup {
	var out []*constGroup
	for _, file := range pass.Files {
		if ast.IsGenerated(file) || isTestFile(pass, file) {
			continue
		}
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.CONST {
				continue
			}
			g := &constGroup{decl: gd, grouped: iotaDependent(gd)}
			for _, spec := range gd.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range vs.Names {
					switch {
					case name.Name == "_":
					case isExcluded(excluded, name.Name):
						g.skipLocal = true
					case name.IsExported():
						g.skipLocal = true
						pass.Reportf(name.Pos(),
							"%s: exported constant %q is declared outside model/entity. "+
								"Fix: keep shared constants in /domain/model or /dal/entity, and declare local ones where they are used",
							ruleID, name.Name)
					default:
						g.names = append(g.names, name)
					}
				}
			}
			if len(g.names) > 0 {
				out = append(out, g)
			}
		}
	}
	return out
}

func candidateObjects(pass *analysis.Pass, groups []*constGroup) map[types.Object]struct{} {
	objs := map[types.Object]struct{}{}
	for _, g := range groups {
		for _, name := range g.names {
			if obj := pass.TypesInfo.Defs[name]; obj != nil {
				objs[obj] = struct{}{}
			}
		}
	}
	return objs
}

// useInfo — where the constant is used.
type useInfo struct {
	funcs map[*ast.FuncDecl]struct{}
	// nonLocal — a use outside a function body (a package-level var/const/type,
	// a signature) or from a test/generated file: such a constant
	// cannot be moved inside a function.
	nonLocal bool
}

func collectUsage(pass *analysis.Pass, candidates map[types.Object]struct{}) map[types.Object]*useInfo {
	usage := map[types.Object]*useInfo{}
	record := func(n ast.Node, fn *ast.FuncDecl) {
		ast.Inspect(n, func(node ast.Node) bool {
			id, ok := node.(*ast.Ident)
			if !ok {
				return true
			}
			obj := pass.TypesInfo.Uses[id]
			if obj == nil {
				return true
			}
			if _, ok := candidates[obj]; !ok {
				return true
			}
			info := usage[obj]
			if info == nil {
				info = &useInfo{funcs: map[*ast.FuncDecl]struct{}{}}
				usage[obj] = info
			}
			if fn == nil {
				info.nonLocal = true
			} else {
				info.funcs[fn] = struct{}{}
			}
			return true
		})
	}
	for _, file := range pass.Files {
		// A use from a test or generated file makes the constant
		// immovable — those files are not "fixed up" after the edit.
		immovable := ast.IsGenerated(file) || isTestFile(pass, file)
		for _, decl := range file.Decls {
			fn, isFunc := decl.(*ast.FuncDecl)
			if !isFunc || immovable {
				record(decl, nil)
				continue
			}
			// The receiver and the signature are evaluated outside the body — a use
			// there (e.g. an array length) prevents moving the constant inside.
			if fn.Recv != nil {
				record(fn.Recv, nil)
			}
			record(fn.Type, nil)
			if fn.Body != nil {
				record(fn.Body, fn)
			}
		}
	}
	return usage
}

func reportGroup(pass *analysis.Pass, g *constGroup, usage map[types.Object]*useInfo) {
	if g.grouped {
		reportIotaGroup(pass, g, usage)
		return
	}
	for _, name := range g.names {
		info := usage[pass.TypesInfo.Defs[name]]
		// An unused constant is the domain of unused; nonLocal and ≥2 functions are fine.
		if info == nil || info.nonLocal || len(info.funcs) != 1 {
			continue
		}
		pass.Reportf(name.Pos(),
			"%s: constant %q is used only in %q. Fix: declare it inside that function",
			ruleID, name.Name, funcDisplayName(soleFunc(info.funcs)))
	}
}

// reportIotaGroup: an iota block can only be moved as a whole — the diagnostic
// is emitted when all its constants are used by one and the same function.
func reportIotaGroup(pass *analysis.Pass, g *constGroup, usage map[types.Object]*useInfo) {
	if g.skipLocal {
		return
	}
	funcs := map[*ast.FuncDecl]struct{}{}
	used := false
	for _, name := range g.names {
		info := usage[pass.TypesInfo.Defs[name]]
		if info == nil {
			continue
		}
		if info.nonLocal {
			return
		}
		used = true
		for fn := range info.funcs {
			funcs[fn] = struct{}{}
		}
	}
	if !used || len(funcs) != 1 {
		return
	}
	pass.Reportf(g.decl.Pos(),
		"%s: this constant group is used only in %q. Fix: declare it inside that function",
		ruleID, funcDisplayName(soleFunc(funcs)))
}

func inAllowedScope(pkgPath string) bool {
	for _, scope := range allowedScopes {
		if pathseg.HasLayer(pkgPath, scope...) {
			return true
		}
	}
	return false
}

func isExcluded(excluded map[string]struct{}, name string) bool {
	_, ok := excluded[name]
	return ok
}

// iotaDependent reports whether the const block's values are tied via iota
// (an explicit use or inheriting the value of the previous spec).
func iotaDependent(gd *ast.GenDecl) bool {
	for _, spec := range gd.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		if len(vs.Values) == 0 || slices.ContainsFunc(vs.Values, usesIota) {
			return true
		}
	}
	return false
}

func usesIota(e ast.Expr) bool {
	found := false
	ast.Inspect(e, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && id.Name == "iota" {
			found = true
		}
		return !found
	})
	return found
}

// funcDisplayName — the function name for the diagnostic: a method is "Type.Method".
func funcDisplayName(fn *ast.FuncDecl) string {
	if fn.Recv != nil {
		if recv := recvTypeName(fn); recv != "" {
			return recv + "." + fn.Name.Name
		}
	}
	return fn.Name.Name
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

func soleFunc(m map[*ast.FuncDecl]struct{}) *ast.FuncDecl {
	for fn := range m {
		return fn
	}
	return nil
}

func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	return strings.HasSuffix(filepath.Base(pass.Fset.Position(file.Pos()).Filename), "_test.go")
}
