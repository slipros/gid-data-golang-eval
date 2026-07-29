// Package modlayout answers one question for a rule: does the module under
// analysis follow the service layout at all?
//
// The layer rules of this styleguide describe a gid.team service — /domain,
// /dal, /server, /app and their neighbours. A library module (libs/trino,
// libs/logger) is flat by nature: it has no layers to place code into, so a
// rule saying "move this to /dal/repository" has nothing to point at. Such
// rules ask IsServiceModule first and stay silent in a library.
//
// The verdict is per module, not per package: the module root is found by
// walking up from the package's own directory to its go.mod, and the layout is
// read from the directories that sit there (directly or under internal/). The
// answer is cached per root — the walk happens once per module per run.
//
// When no go.mod is found — an analysistest fixture in GOPATH style, a package
// compiled outside a module — the answer is "service", so the rules keep the
// behaviour they had before this package existed.
package modlayout

import (
	"go/ast"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"
)

// serviceDirs — the directories whose presence marks a service layout. Any one
// of them is enough: a service always has at least a composition root and one
// business layer, while a library has none of them.
var serviceDirs = []string{"domain", "dal", "server", "app", "usecase", "repository"}

// cache — module root -> verdict, so the directory walk runs once per module.
var cache sync.Map

// IsServiceModule reports whether the package under analysis belongs to a
// module laid out as a service. A library module — flat, with no layer
// directories — gets false, and the layer rules skip it.
func IsServiceModule(pass *analysis.Pass) bool {
	dir := packageDir(pass)
	if dir == "" {
		return true // nothing to inspect: keep the pre-existing behaviour
	}

	root, modPath, ok := moduleRoot(dir)
	if !ok || !belongsTo(pass.Pkg.Path(), modPath) {
		// No go.mod above the package, or the go.mod found belongs to another
		// module than the package under analysis (a GOPATH-style analysistest
		// fixture sitting inside this repository). Nothing reliable to read.
		return true
	}

	if v, cached := cache.Load(root); cached {
		if verdict, isBool := v.(bool); isBool {
			return verdict
		}
	}

	verdict := hasServiceDirs(root)
	cache.Store(root, verdict)

	return verdict
}

// packageDir — the directory holding the package's first file.
func packageDir(pass *analysis.Pass) string {
	for _, file := range pass.Files {
		if name := fileName(pass, file); name != "" {
			return filepath.Dir(name)
		}
	}

	return ""
}

func fileName(pass *analysis.Pass, file *ast.File) string {
	tokenFile := pass.Fset.File(file.Pos())
	if tokenFile == nil {
		return ""
	}

	return tokenFile.Name()
}

// moduleRoot walks up from dir to the nearest directory holding a go.mod,
// returning that directory and the module path declared inside it.
func moduleRoot(dir string) (root, modPath string, ok bool) {
	for {
		if path, found := modulePath(filepath.Join(dir, "go.mod")); found {
			return dir, path, true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", false
		}

		dir = parent
	}
}

// modulePath reads the module path from a go.mod file.
func modulePath(goMod string) (string, bool) {
	data, err := os.ReadFile(goMod)
	if err != nil {
		return "", false
	}

	for line := range strings.SplitSeq(string(data), "\n") {
		rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module")
		if !ok {
			continue
		}

		path := strings.TrimSpace(rest)
		if path != "" {
			return path, true
		}
	}

	return "", true // a go.mod without a module line: the root is still found
}

// belongsTo reports whether the package under analysis is part of the module.
func belongsTo(pkgPath, modPath string) bool {
	if modPath == "" {
		return true // the module path is unknown — do not second-guess the root
	}

	return pkgPath == modPath || strings.HasPrefix(pkgPath, modPath+"/")
}

// hasServiceDirs reports whether the module root holds a layer directory —
// either at the top level or under internal/, the two layouts the styleguide
// describes.
func hasServiceDirs(root string) bool {
	for _, layer := range serviceDirs {
		if isDir(filepath.Join(root, layer)) || isDir(filepath.Join(root, "internal", layer)) {
			return true
		}
	}

	return false
}

func isDir(path string) bool {
	info, err := os.Stat(path)

	return err == nil && info.IsDir()
}
