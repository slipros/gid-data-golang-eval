// Package flatlayout реализует правило GID-138: репозитории и сервисы
// живут в корне своего слоя, без группирующих подпакетов. Репозиторий,
// работающий с redis, лежит в /dal/repository, а не в /dal/repository/redis.
//
// Легитимные подпакеты из стайлгайда: convert/ и build/ у репозитория,
// convert/ у сервиса.
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

// Analyzer — правило GID-138: репозитории и сервисы живут в корне /dal/repository и /domain/service, без подпапок.
var Analyzer = &analysis.Analyzer{
	Name: "gidflatlayout",
	Doc:  ruleID + ": репозитории и сервисы живут в корне /dal/repository и /domain/service, без подпапок",
	Run:  run,
}

// layerRoot — корень слоя и разрешённые в нём подпакеты.
type layerRoot struct {
	seq     []string
	allowed map[string]struct{}
}

func run(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()
	//nolint:gidallptr // плагин не зависит от внутренней библиотеки gdhelper
	for _, root := range layerRoots {
		idx := pathseg.Index(pkgPath, root.seq...)
		if idx < 0 {
			continue
		}
		segs := pathseg.Segments(pkgPath)
		next := idx + len(root.seq)
		if next >= len(segs) {
			continue // сам корень слоя — ок
		}
		if _, ok := root.allowed[segs[next]]; ok {
			continue
		}
		rootPath := strings.Join(root.seq, "/")
		for _, file := range pass.Files {
			pass.Reportf(file.Name.Pos(),
				"%s: пакет %q — группирующие подпакеты в /%s запрещены, сущности слоя живут в его корне",
				ruleID, pkgPath, rootPath)
		}
	}
	return nil, nil
}
