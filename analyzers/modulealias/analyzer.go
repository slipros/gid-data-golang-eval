// Package modulealias implements rule GID-240 (module-common-alias, linter
// gidmodulealias): inside the pkg/<module> application-module layout
// (module.md — pkg/<module> repeats the internal/ layered structure, scoped
// to one module), an import of a shared internal/** entity from the same
// repository must carry an alias prefixed with a fixed marker (default
// "common"), e.g. commonservice, commonmodel — so a reader can immediately
// tell an application module's own types (service, model, ...) apart from
// the common entities it borrows from internal/.
//
// Scope: packages whose import path contains a /pkg/<module> segment — the
// same module boundary as analyzers/layerimports (module.md). A package
// outside pkg/<module> is out of scope: internal/** importing internal/**
// is an ordinary same-module import, not cross-module borrowing, and needs
// no alias.
//
// Checked import: same-repository internal/** — the repository prefix is
// everything before /pkg/<module>, so "<prefix>/internal/..." is in scope.
// Third-party imports and imports of other modules are not affected.
//
// Diagnostics (GID-240):
//   - no alias (import "prefix/internal/domain/service");
//   - an alias without the required prefix (import svc "...");
//   - a dot-import (import . "...") — an alias is required, "." does not qualify.
//
// Not flagged:
//   - a blank import (import _ "...") — a side-effect-only import, not a
//     reference to the entity, so no alias is meaningful;
//   - an alias that already carries the required prefix (commonservice).
//
// Settings.Prefix overrides the default "common" alias prefix.
// Escape hatch: //nolint:gidmodulealias.
package modulealias

import (
	"go/ast"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-240"

// Analyzer — GID-240 with the default "common" alias prefix.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Prefix — the required alias prefix for imports of shared internal
	// entities inside pkg/<module>. Defaults to "common".
	Prefix string `json:"prefix"`
}

// NewAnalyzer builds the GID-240 analyzer with the given alias prefix.
func NewAnalyzer(s Settings) *analysis.Analyzer {
	const defaultPrefix = "common"
	prefix := s.Prefix
	if prefix == "" {
		prefix = defaultPrefix
	}
	return &analysis.Analyzer{
		Name: "gidmodulealias",
		Doc: ruleID + ": inside pkg/<module>, an import of a shared " +
			"internal/** entity must carry a " + prefix + "-prefixed alias " +
			"(e.g. " + prefix + "service). Fix: alias the import, for example " +
			prefix + `service "<repo>/internal/domain/service"`,
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, prefix)
		},
	}
}

func run(pass *analysis.Pass, prefix string) (any, error) {
	repoPrefix, ok := pkgModuleBoundary(pass.Pkg.Path())
	if !ok {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		checkImports(pass, file, repoPrefix, prefix)
	}
	return nil, nil
}

func checkImports(pass *analysis.Pass, file *ast.File, repoPrefix, prefix string) {
	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		if !isSharedInternal(path, repoPrefix) {
			continue
		}
		checkAlias(pass, imp, path, prefix)
	}
}

// isSharedInternal reports whether path is a same-repository internal/**
// import: repoPrefix + "/internal/...".
func isSharedInternal(path, repoPrefix string) bool {
	const internalSeg = "/internal/"
	return strings.HasPrefix(path, repoPrefix+internalSeg)
}

// checkAlias reports GID-240 unless the import already carries a
// prefix-prefixed alias. A blank import ("_") is a side-effect-only import
// and is skipped — there is no entity reference to alias.
func checkAlias(pass *analysis.Pass, imp *ast.ImportSpec, path, prefix string) {
	if imp.Name == nil {
		report(pass, imp, path, prefix)
		return
	}
	name := imp.Name.Name
	if name == "_" {
		return
	}
	if name == "." || !strings.HasPrefix(name, prefix) {
		report(pass, imp, path, prefix)
	}
}

func report(pass *analysis.Pass, imp *ast.ImportSpec, path, prefix string) {
	pass.Reportf(imp.Pos(),
		"%s: import %q of shared internal entities must carry a %s-prefixed alias (e.g. %sservice)",
		ruleID, path, prefix, prefix)
}

// pkgModuleBoundary resolves the pkg/<module> application-module layout
// boundary for pkgPath (module.md, mirrors analyzers/layerimports):
// repoPrefix is the repository prefix before /pkg/<module>, used to
// recognize same-repository internal/** imports. ok is false if pkgPath has
// no /pkg/<module> segment — the package is then out of GID-240's scope.
func pkgModuleBoundary(pkgPath string) (repoPrefix string, ok bool) {
	// The module.md application-module layout marker (mirrors
	// analyzers/layerimports): pkg/<module>/ repeats the same layered
	// structure (dal/, domain/, server/) as internal/.
	const pkgSeg = "/pkg/"
	before, rest, cut := strings.Cut(pkgPath, pkgSeg)
	if !cut || rest == "" {
		return "", false
	}
	if module, _, _ := strings.Cut(rest, "/"); module == "" {
		return "", false
	}
	return before, true
}
