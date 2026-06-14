// Package flatlayout implements rule GID-138: repositories and services
// live at the root of their layer, without grouping subpackages. A repository
// working with redis lives in /dal/repository, not in /dal/repository/redis.
//
// Legitimate subpackages from the styleguide: convert/ and build/ for a
// repository, convert/ for a service.
package flatlayout

import (
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-138"

var layerRoots = []layerRoot{
	{seq: []string{"dal", "repository"}, allowed: map[string]struct{}{"convert": {}, "build": {}}},
	{seq: []string{"domain", "service"}, allowed: map[string]struct{}{"convert": {}}},
}

// Analyzer — rule GID-138: repositories and services live at the root of /dal/repository and /domain/service, without subfolders. Fix: move the entity to the layer root.
var Analyzer = &analysis.Analyzer{
	Name: "gidflatlayout",
	Doc:  ruleID + ": repositories and services live at the root of /dal/repository and /domain/service, without subfolders. Fix: move the entity to the layer root",
	Run:  run,
}

// layerRoot — a layer root and the subpackages allowed in it.
type layerRoot struct {
	seq     []string
	allowed map[string]struct{}
}

func run(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()
	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
	for _, root := range layerRoots {
		idx := pathseg.Index(pkgPath, root.seq...)
		if idx < 0 {
			continue
		}
		segs := pathseg.Segments(pkgPath)
		next := idx + len(root.seq)
		if next >= len(segs) {
			continue // the layer root itself — fine
		}
		if _, ok := root.allowed[segs[next]]; ok {
			continue
		}
		rootPath := strings.Join(root.seq, "/")
		for _, file := range pass.Files {
			pass.Reportf(file.Name.Pos(),
				"%s: package %q. Fix: grouping subpackages in /%s are forbidden, keep layer entities at its root",
				ruleID, pkgPath, rootPath)
		}
	}
	return nil, nil
}
