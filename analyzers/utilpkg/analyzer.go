// Package utilpkg implements rule GID-187 (no-util-package): package names
// like util/utils/common/helper/helpers/shared/misc/lib/base are forbidden.
// Such a package is a junk drawer with no area of responsibility: the name
// does not tell what it provides. A package is named after what it contains
// (e.g. parse, retry, money).
//
// The package name is checked (pass.Pkg.Name(), which matches the last path
// segment) — case-insensitively. The _test suffix of a test package is
// normalized (utils_test → utils). One report per package: on the package
// clause of the first (non-generated) file. The name dictionary is configured
// via settings.names and fully replaces the default list.
//
// LoadMode — Syntax: no types needed, the package name and the AST suffice.
package utilpkg

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-187"

// Analyzer — the variant with the default name blacklist.
var Analyzer = NewAnalyzer(Settings{})

// Settings — the linter settings from .golangci.yml.
type Settings struct {
	// Names — the blacklist of package names (replaces the default list).
	Names []string `json:"names"`
}

// NewAnalyzer builds the GID-187 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	names := resolveNames(s)
	blacklist := make(map[string]struct{}, len(names))
	for _, n := range names {
		blacklist[strings.ToLower(n)] = struct{}{}
	}
	return &analysis.Analyzer{
		Name: "gidutilpkg",
		Doc: ruleID + ": forbid junk-drawer packages (util, utils, common, helper, …). " +
			"Fix: name the package after what it provides",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, blacklist)
		},
	}
}

func resolveNames(s Settings) []string {
	if len(s.Names) == 0 {
		return []string{
			"util",
			"utils",
			"common",
			"helper",
			"helpers",
			"shared",
			"misc",
			"lib",
			"base",
		}
	}
	return s.Names
}

func run(pass *analysis.Pass, blacklist map[string]struct{}) (any, error) {
	name := normalize(pass.Pkg.Name())
	if _, banned := blacklist[name]; !banned {
		return nil, nil
	}

	// One report per package — on the package clause of the first non-generated file.
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		pass.Reportf(file.Name.Pos(),
			"%s: package %q is a junk drawer with no responsibility. Fix: name the package after what it provides",
			ruleID, pass.Pkg.Name())
		return nil, nil
	}
	return nil, nil
}

// normalize lowercases the package name and strips the _test suffix
// of test packages (utils_test → utils).
func normalize(name string) string {
	name = strings.ToLower(name)
	return strings.TrimSuffix(name, "_test")
}
