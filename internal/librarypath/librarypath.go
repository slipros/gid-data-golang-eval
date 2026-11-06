package librarypath

import (
	"cmp"
	"reflect"
	"slices"
	"strings"

	"github.com/pkg/errors"

	"golang.org/x/mod/module"
	"golang.org/x/tools/go/analysis"
)

type replacement struct {
	original string
	fork     string
}

type mapping []replacement

func Configure(analyzer *analysis.Analyzer, libraries map[string]string) (*analysis.Analyzer, error) {
	if len(libraries) == 0 {
		return analyzer, nil
	}

	paths := make(mapping, 0, len(libraries))
	targets := make(map[string]string, len(libraries))
	for original, fork := range libraries {
		if err := module.CheckImportPath(original); err != nil {
			return nil, errors.Wrapf(err, "library source %q", original)
		}

		if err := module.CheckImportPath(fork); err != nil {
			return nil, errors.Wrapf(err, "library replacement %q", fork)
		}

		if previous, exists := targets[fork]; exists {
			return nil, errors.Errorf("library replacement %q is shared by %q and %q", fork, previous, original)
		}

		targets[fork] = original
		paths = append(paths, replacement{original: original, fork: fork})
	}

	slices.SortFunc(paths, func(a, b replacement) int {
		return cmp.Compare(a.original, b.original)
	})

	resolver := &analysis.Analyzer{
		Name:       analyzer.Name + "libraries",
		Doc:        "Resolve configured library import paths.",
		ResultType: reflect.TypeFor[mapping](),
		Run: func(_ *analysis.Pass) (any, error) {
			return paths, nil
		},
	}

	configured := *analyzer
	configured.Requires = append(slices.Clone(analyzer.Requires), resolver)
	configured.Doc = paths.rewrite(analyzer.Doc)
	configured.Run = func(pass *analysis.Pass) (any, error) {
		local := *pass
		local.Report = func(diagnostic analysis.Diagnostic) {
			diagnostic.Message = paths.rewrite(diagnostic.Message)
			pass.Report(diagnostic)
		}

		return analyzer.Run(&local)
	}

	return &configured, nil
}

func Canonical(pass *analysis.Pass, path string) string {
	return resolve(pass, path, false)
}

func Replacement(pass *analysis.Pass, path string) string {
	return resolve(pass, path, true)
}

func resolve(pass *analysis.Pass, path string, toFork bool) string {
	var best replacement
	longest := 0

	for _, result := range pass.ResultOf {
		paths, ok := result.(mapping)
		if !ok {
			continue
		}

		//nolint:gidallptr // the standalone plugin does not depend on the corporate helper library
		for _, entry := range paths {
			prefix := entry.fork
			if toFork {
				prefix = entry.original
			}

			if len(prefix) > longest && (path == prefix || strings.HasPrefix(path, prefix+"/")) {
				best, longest = entry, len(prefix)
			}
		}
	}

	if longest == 0 {
		return path
	}

	if toFork {
		return best.fork + path[longest:]
	}

	return best.original + path[longest:]
}

func (mm mapping) rewrite(message string) string {
	pairs := make([]string, 0, len(mm)+len(mm))
	ordered := slices.Clone(mm)
	slices.SortFunc(ordered, func(a, b replacement) int {
		return cmp.Compare(len(b.original), len(a.original))
	})

	//nolint:gidallptr // the standalone plugin does not depend on the corporate helper library
	for _, entry := range ordered {
		pairs = append(pairs, entry.original, entry.fork)
	}

	replacer := strings.NewReplacer(pairs...)

	return replacer.Replace(message)
}
