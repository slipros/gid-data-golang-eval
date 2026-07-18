// Package errname implements rule GID-234 (giderrname): every error in
// /domain/model is concrete and bound to its entity.
//
// The styleguide (model.md, layer errors section) names errors as
// Err<Entity><Reason> — ErrSnapshotNotFound, ErrSnapshotAlreadyExists.
// Generic names like ErrNotFound are forbidden in the domain model:
// concrete errors are essential for handling on the upper layers.
// Generic names are the convention of /dal/entity (ErrNoResult,
// ErrAlreadyExists), so the dal layer is out of scope of this rule.
//
// Deterministic core: in /domain/model (root package and subpackages,
// matched via pathseg) a package-level variable of type error whose name
// is in the banned generic list (settings.names) is reported. Names that
// carry an entity prefix (ErrSnapshotNotFound) pass. Point exceptions —
// //nolint:giderrname; centralized — settings.exclude.
package errname

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-234"

// Analyzer — the variant with the default list of generic names.
var Analyzer = NewAnalyzer(Settings{})

// Settings — rule GID-234 settings from .golangci.yml.
type Settings struct {
	// Names — forbidden "generic" error names.
	// Empty → default list (ErrNotFound, ErrAlreadyExists, ...).
	Names []string `json:"names"`
	// Exclude — names of exception variables that are not reported.
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-234 analyzer with the given list of generic names.
func NewAnalyzer(s Settings) *analysis.Analyzer {
	names := resolveNames(s)
	banned := make(map[string]struct{}, len(names))
	for _, n := range names {
		banned[n] = struct{}{}
	}
	return &analysis.Analyzer{
		Name: "giderrname",
		Doc:  ruleID + ": generic error names (ErrNotFound) are forbidden in /domain/model. Fix: bind it to the entity: ErrSnapshotNotFound",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, banned, s.Exclude)
		},
	}
}

func resolveNames(s Settings) []string {
	if len(s.Names) == 0 {
		return []string{
			"ErrNotFound",
			"ErrAlreadyExists",
			"ErrExists",
			"ErrInvalid",
			"ErrInvalidInput",
			"ErrInternal",
			"ErrConflict",
			"ErrBadRequest",
			"ErrFailed",
			"ErrUnknown",
			"ErrNoResult",
			"ErrForbidden",
			"ErrUnauthorized",
		}
	}
	return s.Names
}

func run(pass *analysis.Pass, banned map[string]struct{}, excluded []string) (any, error) {
	// Layer root and subpackages: /domain/model, /domain/model/filter, ...
	if !pathseg.HasLayer(pass.Pkg.Path(), "domain", "model") {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		fname := pass.Fset.Position(file.Pos()).Filename
		if strings.HasSuffix(fname, "_test.go") {
			continue
		}
		checkFile(pass, banned, excluded, file)
	}
	return nil, nil
}

// checkFile reports package-level error variables with a generic name.
func checkFile(pass *analysis.Pass, banned map[string]struct{}, excluded []string, file *ast.File) {
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range vs.Names {
				if name.Name == "_" {
					continue
				}
				if _, ok := banned[name.Name]; !ok {
					continue // the name carries an entity prefix — ok
				}
				if exclude.Match(excluded, name.Name, name.Name) {
					continue
				}
				obj := pass.TypesInfo.Defs[name]
				if obj == nil || !implementsError(obj.Type()) {
					continue // not an error — the rule does not apply
				}
				pass.Reportf(name.Pos(),
					"%s: generic error name %q in domain model. Fix: bind it to the entity: %s",
					ruleID, name.Name, entityBoundExample(name.Name))
			}
		}
	}
}

// entityBoundExample builds a fix example: ErrNotFound → ErrSnapshotNotFound.
func entityBoundExample(name string) string {
	if rest, ok := strings.CutPrefix(name, "Err"); ok && rest != "" {
		return "ErrSnapshot" + rest
	}
	return "ErrSnapshotNotFound"
}

func implementsError(t types.Type) bool {
	errObj := types.Universe.Lookup("error")
	errType := errObj.Type()
	errIface, ok := errType.Underlying().(*types.Interface)
	if !ok {
		return false
	}
	return types.Implements(t, errIface)
}
