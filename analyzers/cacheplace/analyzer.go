// Package cacheplace реализует правило GID-159: кэш живёт в repository.
// Если сущности нужен кэш — он оформляется кэширующим репозиторием в
// /dal/repository, который оборачивает основной (прямой ссылкой, без
// интерфейса). Domain-слой (service, usecase, model) про кэш не знает —
// импорт кэш-библиотек там запрещён. Кэш может быть любым: in-memory LRU,
// redis и т.п.
//
// Список кэш-библиотек настраивается через settings.packages
// (заменяет дефолтный).
package cacheplace

import (
	"go/ast"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-159"

// defaultPackages — известные кэш-библиотеки (матчинг по префиксу пути,
// версионные суффиксы /v2, /v9 покрываются автоматически).
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

// Analyzer — вариант с дефолтным списком кэш-библиотек.
var Analyzer = NewAnalyzer(Settings{})

// Settings — настройки линтера из .golangci.yml.
type Settings struct {
	// Packages — кэш-библиотеки (префиксы import-путей).
	// Заменяет дефолтный список.
	Packages []string `json:"packages"`
}

// NewAnalyzer строит анализатор GID-159 из настроек линтера (.golangci.yml).
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
