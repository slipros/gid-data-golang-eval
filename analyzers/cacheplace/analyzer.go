// Package cacheplace implements rule GID-159: the cache lives in a repository.
// If an entity needs a cache, it is implemented as a caching repository in
// /dal/repository that wraps the main one (by a direct reference, without an
// interface). The domain layer (service, usecase, model) knows nothing about
// the cache — importing cache libraries there is forbidden. The cache can be
// anything: in-memory LRU, redis, etc.
//
// The list of cache libraries is configured via settings.packages
// (it replaces the default one).
package cacheplace

import (
	"go/ast"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-159"

// defaultPackages — well-known cache libraries (matched by path prefix,
// versioned suffixes /v2, /v9 are covered automatically).
var defaultPackages = []string{
	"github.com/redis/go-redis",
	"github.com/go-redis/redis",
	"github.com/hashicorp/golang-lru",
	"github.com/dgraph-io/ristretto",
	"github.com/allegro/bigcache",
	"github.com/coocood/freecache",
	"github.com/patrickmn/go-cache",
	"github.com/bradfitz/gomemcache",
}

// Analyzer — the variant with the default list of cache libraries.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Packages — cache libraries (import path prefixes).
	// Replaces the default list.
	Packages []string `json:"packages"`
}

// NewAnalyzer builds the GID-159 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	pkgs := s.Packages
	if len(pkgs) == 0 {
		pkgs = defaultPackages
	}
	return &analysis.Analyzer{
		Name: "gidcacheplace",
		Doc:  ruleID + ": caching lives in a repository decorator; the domain layer knows nothing about cache. Fix: wrap caching in /dal/repository",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, pkgs)
		},
	}
}

func run(pass *analysis.Pass, cachePkgs []string) (any, error) {
	if !pathseg.Contains(pass.Pkg.Path(), "domain") {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, imp := range file.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			if !isCachePkg(path, cachePkgs) {
				continue
			}
			pass.Reportf(imp.Pos(),
				"%s: importing the cache library %q in the domain layer is forbidden. Fix: implement caching "+
					"as a caching repository in /dal/repository that wraps the main one",
				ruleID, path)
		}
	}
	return nil, nil
}

func isCachePkg(path string, cachePkgs []string) bool {
	for _, pkg := range cachePkgs {
		if path == pkg || strings.HasPrefix(path, pkg+"/") {
			return true
		}
	}
	return false
}
