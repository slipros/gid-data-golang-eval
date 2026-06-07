// Package dirtree реализует правило GID-158: контроль дерева папок.
// Для каждой папки из настроек задаётся перечень разрешённых подпапок;
// появление чужой папки — предупреждение (например, новая папка в
// internal/; perhaps it should be a service or usecase).
//
// Дерево настраивается в .golangci.yml (settings.tree), ключ — путь папки
// (сегменты через /, матчится в любом месте import-пути), значение —
// разрешённые подпапки. Заданное в settings дерево заменяет дефолтное.
package dirtree

import (
	"go/ast"
	"slices"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-158"

// defaultTree — каноничная структура сервиса (ARCHITECTURE.md).
// Папки, не указанные ключом, не ограничиваются.
var defaultTree = map[string][]string{
	"internal":                {"app", "client", "dal", "domain", "event", "metric", "server"},
	"internal/dal":            {"entity", "repository"},
	"internal/dal/repository": {"convert", "build"},
	"internal/domain":         {"model", "service", "usecase"},
	"internal/domain/service": {"convert"},
	"internal/server":         {"grpc", "http"},
}

// Analyzer — вариант с деревом по умолчанию.
var Analyzer = NewAnalyzer(Settings{})

// Settings — настройки линтера из .golangci.yml.
type Settings struct {
	// Tree: "папка" -> разрешённые подпапки. Заменяет дефолтное дерево.
	Tree map[string][]string `json:"tree"`
}

// NewAnalyzer строит анализатор GID-158 из настроек линтера (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	tree := s.Tree
	if len(tree) == 0 {
		tree = defaultTree
	}
	return &analysis.Analyzer{
		Name: "giddirtree",
		Doc:  ruleID + ": a folder may contain only allowed subfolders (settings.tree). Fix: move the folder or add it to settings.tree",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, tree)
		},
	}
}

func run(pass *analysis.Pass, tree map[string][]string) (any, error) {
	pkgPath := pass.Pkg.Path()
	segs := pathseg.Segments(pkgPath)

	keys := make([]string, 0, len(tree))
	for key := range tree {
		keys = append(keys, key)
	}
	sort.Strings(keys) // детерминированный порядок диагностик

	for _, key := range keys {
		seq := pathseg.Segments(key)
		idx := pathseg.Index(pkgPath, seq...)
		if idx < 0 {
			continue
		}
		next := idx + len(seq)
		if next >= len(segs) {
			continue // пакет — сама папка-ключ
		}
		if slices.Contains(tree[key], segs[next]) {
			continue
		}
		report(pass, key, segs[next], tree[key])
	}
	return nil, nil
}

func report(pass *analysis.Pass, key, dir string, allowed []string) {
	hint := ""
	if key == "internal" {
		hint = "; perhaps it should be a service or usecase"
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		pass.Reportf(file.Name.Pos(),
			"%s: folder %q is not allowed in %s/ (allowed: %s)%s; configure the tree via settings.tree",
			ruleID, dir, key, strings.Join(allowed, ", "), hint)
	}
}
