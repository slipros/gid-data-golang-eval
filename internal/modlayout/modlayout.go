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

// The marker of a service layout is the composition root internal/app: by the
// service template every gid.team service has one, and no library does — a
// library that publishes an app/ package (libs/helper) puts it at the root, not
// under internal/. The fallback marker is the pair domain + dal — a service
// owns both a business layer and a data layer, while a library holding one of
// them (libs/grpc has internal/domain, libs/http has server/) never holds both.
//
// Single directories are deliberately NOT markers: server/, client/ or domain/
// on their own are ordinary package names in a transport library.
// serviceCache — module root -> verdict, so the directory walk runs once per module.
var serviceCache sync.Map

// dataCache — module root -> "the module owns a data layer", cached the same way.
var dataCache sync.Map

// rootOf — package directory -> the module root above it. Without this the
// walk re-ran for every package of every rule that asks: eight rules across a
// module of sixty packages meant hundreds of repeated stat/read syscalls, and
// re-reading the whole go.mod each time was the single largest allocation of
// those rules. Every directory the walk passes through is memoized, so the
// ancestors shared by sibling packages are resolved once.
var rootOf sync.Map

// IsServiceModule reports whether the package under analysis belongs to a
// module laid out as a service: one holding a composition root (app/), or both
// a business and a data layer (domain/ + dal/). A library module gets false,
// and the layer rules skip it.
func IsServiceModule(pass *analysis.Pass) bool {
	return moduleVerdict(pass, &serviceCache, hasServiceDirs)
}

// HasDataLayer reports whether the module under analysis owns a data layer at
// all — /dal (the styleguide layout) or a bare /repository (a service that
// never grew a dal). A rule whose fix is "move this into a repository" has
// nothing to point at in a module without one, and must ask before reporting.
//
// The case this exists for is a BFF (lk-api): its whole job is to call other
// services over gRPC and shape the answer for the frontend, so it holds
// /domain/service and /server/http and no data layer whatsoever. Demanding a
// repository there is demanding a layer the service does not have — the ~106
// GID-160 diagnostics it produced were all of that kind.
//
// A module with no go.mod above it (an analysistest fixture, a package built
// outside a module) gets true, the same fallback as IsServiceModule: without a
// root to read, the rule keeps the behaviour it had before this check existed.
func HasDataLayer(pass *analysis.Pass) bool {
	return moduleVerdict(pass, &dataCache, hasDataDirs)
}

// moduleVerdict answers a question about the layout of the module the package
// belongs to, caching the answer per module root.
func moduleVerdict(pass *analysis.Pass, cache *sync.Map, inspect func(root string) bool) bool {
	dir := packageDir(pass)
	if dir == "" {
		return true // nothing to inspect: keep the pre-existing behaviour
	}

	root, modPath, ok := moduleRoot(dir)
	if !ok || !belongsTo(pkgPath(pass), modPath) {
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

	verdict := inspect(root)
	cache.Store(root, verdict)

	return verdict
}

// pkgPath — the import path of the package under analysis, empty when the
// analyzer runs without type information (a syntax-only load mode leaves
// pass.Pkg nil, and reading it would panic).
func pkgPath(pass *analysis.Pass) string {
	if pass.Pkg == nil {
		return ""
	}

	return pass.Pkg.Path()
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

// moduleResult — the memoized answer of moduleRoot for one directory.
type moduleResult struct {
	root    string
	modPath string
	ok      bool
}

// moduleRoot walks up from dir to the nearest directory holding a go.mod,
// returning that directory and the module path declared inside it.
func moduleRoot(dir string) (root, modPath string, ok bool) {
	// visited — the directories walked past before the answer was known; they
	// all share it, so one walk fills the cache for the whole branch.
	var visited []string

	for {
		if v, cached := rootOf.Load(dir); cached {
			res, isResult := v.(moduleResult)
			if isResult {
				storeAll(visited, res)

				return res.root, res.modPath, res.ok
			}
		}

		if path, found := modulePath(filepath.Join(dir, "go.mod")); found {
			res := moduleResult{root: dir, modPath: path, ok: true}
			rootOf.Store(dir, res)
			storeAll(visited, res)

			return res.root, res.modPath, res.ok
		}

		visited = append(visited, dir)

		parent := filepath.Dir(dir)
		if parent == dir {
			storeAll(visited, moduleResult{})

			return "", "", false
		}

		dir = parent
	}
}

func storeAll(dirs []string, res moduleResult) {
	for _, dir := range dirs {
		rootOf.Store(dir, res)
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
	if modPath == "" || pkgPath == "" {
		return true // nothing to compare — do not second-guess the root found
	}

	return pkgPath == modPath || strings.HasPrefix(pkgPath, modPath+"/")
}

// hasServiceDirs reports whether the module root holds a layer directory —
// either at the top level or under internal/, the two layouts the styleguide
// describes.
func hasServiceDirs(root string) bool {
	const (
		rootApp    = "app"
		rootDomain = "domain"
		rootDAL    = "dal"
	)

	// The composition root counts only under internal/: a library may well
	// publish an app/ package of its own (libs/helper does).
	if isDir(filepath.Join(root, "internal", rootApp)) {
		return true
	}

	return hasLayerDir(root, rootDomain) && hasLayerDir(root, rootDAL)
}

// hasDataDirs reports whether the module root holds a data layer: /dal, the
// layout of the styleguide, or a bare /repository for a service laid out
// without one. The presence of the layer is what matters, not what is inside
// it — a dal holding only entities today is where the repository goes
// tomorrow, so the rules that ask for a repository keep working there.
func hasDataDirs(root string) bool {
	const (
		rootDAL        = "dal"
		rootRepository = "repository"
	)

	return hasLayerDir(root, rootDAL) || hasLayerDir(root, rootRepository)
}

// hasLayerDir reports whether the layer directory sits at the module root or
// under its internal/.
func hasLayerDir(root, layer string) bool {
	return isDir(filepath.Join(root, layer)) || isDir(filepath.Join(root, "internal", layer))
}

func isDir(path string) bool {
	info, err := os.Stat(path)

	return err == nil && info.IsDir()
}
