// Package importalias implements rule GID-275 (redundant-import-alias, linter
// gidimportalias): an import of the module's own package carries an alias only
// when the file needs it. Without a reason the alias is noise a reader has to
// decode — the package is called convert, the file calls it
// domainserviceconvert, and nothing else in the file is called convert.
//
// Scope: imports of the module under analysis (pass.Module). Aliases of other
// modules follow conventions of their own — the gd prefix of gid.team
// libraries (gdhelper, gderror), the pb suffix of generated packages — and are
// not judged. Without a module path (GOPATH mode) nothing is the module's own,
// so the rule stays silent.
//
// An alias is needed when dropping it would not compile or would change what a
// name refers to — the package name N (the real one, from type information)
// would then collide with:
//   - the name of another import of the same file, aliased or not (two
//     convert packages: each may keep its alias);
//   - a package-level declaration named N (a file-scope import and a
//     package-scope object cannot share a name);
//   - a predeclared identifier (a package named error would hide the
//     built-in error type from the whole file);
//   - a local declaration visible at a place the file refers to the import
//     (func New(config Config) shadows the config package inside New).
//
// Also not reported:
//   - an alias spelling the real package name where the import path does not
//     suggest it — the name goimports itself would write;
//   - inside pkg/<module>, an import of the repository's shared internal/** —
//     GID-240 makes its common-prefixed alias mandatory;
//   - inside pkg/<module>, a common-prefixed alias of a package outside that
//     module (a nested module borrowing its parent's model) — the same
//     own-versus-borrowed marker GID-240 sets for internal/**;
//   - blank and dot imports — they bind no package name to shorten or clarify;
//   - generated files.
//
// A _test.go file is judged: nothing forces a test to rename a package.
// Settings.Prefix overrides the "common" marker, as in GID-240.
// Escape hatch: //nolint:gidimportalias.
package importalias

import (
	"go/ast"
	"go/types"
	"path"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-275"

// Analyzer — GID-275 with the default "common" borrowed-package prefix.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Prefix — the alias prefix marking a package a pkg/<module> borrows from
	// outside itself. Defaults to "common" (GID-240).
	Prefix string `json:"prefix"`
}

// NewAnalyzer builds the GID-275 analyzer with the given borrowed-package prefix.
func NewAnalyzer(s Settings) *analysis.Analyzer {
	const defaultPrefix = "common"
	prefix := s.Prefix
	if prefix == "" {
		prefix = defaultPrefix
	}

	return &analysis.Analyzer{
		Name: "gidimportalias",
		Doc: ruleID + ": an import of the module's own package is aliased only " +
			"when the file needs it — the package name collides with another " +
			"import, a package-level declaration or a local name where the " +
			`import is used. Fix: drop the alias, import ".../domain/service/convert" ` +
			"and refer to it as convert",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, prefix)
		},
	}
}

// aliasedImport — a judged import with a name of its own and the package it binds.
type aliasedImport struct {
	spec *ast.ImportSpec
	name *types.PkgName
	path string
}

func run(pass *analysis.Pass, prefix string) (any, error) {
	if pass.Module == nil {
		return nil, nil
	}
	var uses map[*types.PkgName][]*ast.Ident
	usesOf := func(pn *types.PkgName) []*ast.Ident {
		if uses == nil {
			uses = usesByPkgName(pass)
		}

		return uses[pn]
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		checkFile(pass, file, usesOf, prefix)
	}

	return nil, nil
}

// usesByPkgName groups the identifiers referring to an imported package by
// the import that binds them. Built on the first alias that survives the
// cheaper checks — most packages never need it.
func usesByPkgName(pass *analysis.Pass) map[*types.PkgName][]*ast.Ident {
	uses := make(map[*types.PkgName][]*ast.Ident)
	for id, obj := range pass.TypesInfo.Uses {
		if pn, ok := obj.(*types.PkgName); ok {
			uses[pn] = append(uses[pn], id)
		}
	}

	return uses
}

func checkFile(pass *analysis.Pass, file *ast.File, usesOf func(*types.PkgName) []*ast.Ident, prefix string) {
	fileScope := pass.TypesInfo.Scopes[file]
	if fileScope == nil {
		return
	}
	occupied := occupiedNames(pass, file, prefix)
	pkgScope := pass.Pkg.Scope()
	for _, imp := range judgedImports(pass, file, prefix) {
		imported := imp.name.Imported()
		pkgName := imported.Name()
		if imp.spec.Name.Name == pkgName {
			if pkgName == assumedName(imp.path) {
				report(pass, imp, pkgName)
			}
			continue
		}
		if occupied[pkgName] > 1 || pkgScope.Lookup(pkgName) != nil ||
			shadowedAtUse(fileScope, usesOf(imp.name), pkgName) {
			continue
		}
		report(pass, imp, pkgName)
	}
}

// judgedImports lists the aliased imports of the module's own packages in
// file, leaving out blank and dot imports and the aliases the module layout
// asks for.
func judgedImports(pass *analysis.Pass, file *ast.File, prefix string) []*aliasedImport {
	var out []*aliasedImport
	for _, spec := range file.Imports {
		if spec.Name == nil || spec.Name.Name == "_" || spec.Name.Name == "." {
			continue
		}
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil || !ownModule(pass, importPath) || layoutAlias(pass, importPath, spec.Name.Name, prefix) {
			continue
		}
		pn, ok := pass.TypesInfo.Defs[spec.Name].(*types.PkgName)
		if !ok {
			continue
		}
		out = append(out, &aliasedImport{spec: spec, name: pn, path: importPath})
	}

	return out
}

func ownModule(pass *analysis.Pass, importPath string) bool {
	modulePath := pass.Module.Path

	return importPath == modulePath || strings.HasPrefix(importPath, modulePath+"/")
}

// layoutAlias reports whether the pkg/<module> layout itself asks for the
// alias: a shared internal/** import (GID-240), or a prefix-marked package
// from outside the importing module.
func layoutAlias(pass *analysis.Pass, importPath, alias, prefix string) bool {
	pkgPath := pass.Pkg.Path()
	if pathseg.SharedInternalImport(pkgPath, importPath) {
		return true
	}
	root, ok := pathseg.PkgModuleRoot(pkgPath)
	if !ok || !strings.HasPrefix(alias, prefix) {
		return false
	}
	importRoot, importOK := pathseg.PkgModuleRoot(importPath)

	return !importOK || importRoot != root
}

// occupiedNames counts, per name, the imports of file that hold it: the name a
// package is referred to by, and also the real package name of an aliased
// import — two aliased convert packages each justify the other's alias, or
// dropping one alias would leave the other one looking redundant. An import
// whose alias the layout asks for holds only that alias: the mandatory
// commonservice frees service for the module's own package.
func occupiedNames(pass *analysis.Pass, file *ast.File, prefix string) map[string]int {
	occupied := make(map[string]int)
	for _, spec := range file.Imports {
		pn := importedName(pass, spec)
		if pn == nil {
			continue
		}
		occupied[pn.Name()]++
		imported := pn.Imported()
		pkgName := imported.Name()
		if pkgName == pn.Name() {
			continue
		}
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err == nil && ownModule(pass, importPath) && layoutAlias(pass, importPath, pn.Name(), prefix) {
			continue
		}
		occupied[pkgName]++
	}

	return occupied
}

// importedName returns the package name spec binds in the file scope, or nil
// for a blank or dot import.
func importedName(pass *analysis.Pass, spec *ast.ImportSpec) *types.PkgName {
	obj := pass.TypesInfo.Implicits[spec]
	if spec.Name != nil {
		obj = pass.TypesInfo.Defs[spec.Name]
	}
	pn, ok := obj.(*types.PkgName)
	if !ok {
		return nil
	}

	return pn
}

// shadowedAtUse reports whether, at any place the file refers to the import,
// the real package name would resolve to another object instead — a local
// declaration or a predeclared identifier (imports and package-level names are
// weighed before it).
func shadowedAtUse(fileScope *types.Scope, uses []*ast.Ident, pkgName string) bool {
	for _, id := range uses {
		inner := fileScope.Innermost(id.Pos())
		if inner == nil {
			continue
		}
		if _, obj := inner.LookupParent(pkgName, id.Pos()); obj != nil {
			return true
		}
	}

	return false
}

// assumedName is the package name a reader — and goimports — infers from an
// import path: the last element, skipping a /vN major-version suffix, without
// a go- prefix and cut at the first character that cannot be in an identifier
// (http.git -> http, go-redis/v9 -> redis, yaml.v3 -> yaml).
func assumedName(importPath string) string {
	base := path.Base(importPath)
	if strings.HasPrefix(base, "v") && isDigits(base[1:]) {
		base = path.Base(path.Dir(importPath))
	}
	base = strings.TrimPrefix(base, "go-")
	if i := strings.IndexFunc(base, notIdentifier); i >= 0 {
		base = base[:i]
	}

	return base
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

func notIdentifier(r rune) bool {
	return r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r)
}

func report(pass *analysis.Pass, imp *aliasedImport, pkgName string) {
	fix := strconv.Quote(imp.path) + " and refer to it as " + pkgName
	if pkgName != assumedName(imp.path) {
		fix = pkgName + " " + strconv.Quote(imp.path) + ", the real package name"
	}
	pass.Reportf(imp.spec.Pos(),
		"%s: import alias %s is not needed — nothing else in the file is called %s. Fix: import it as %s",
		ruleID, imp.spec.Name.Name, pkgName, fix)
}
