// Package storageplace implements rule GID-249 (slug storage-in-repository,
// linter gidstorageplace). Source: ARCHITECTURE.md / dal.md — a data store is
// reached through the repository layer, a client is for an EXTERNAL API
// (HTTP/gRPC/SMTP/queue), not for a database.
//
// A storage driver — a SQL driver/pool (pgx, database/sql, clickhouse), a
// key-value or NoSQL client (go-redis, valkey, rueidis, mongo, etcd, badger,
// bbolt, gocql, elasticsearch) or the corresponding gid.team library — is
// imported ONLY from:
//
//   - /dal/** — the repository layer itself, including /dal/entity (a pgtype
//     column type in an entity struct is legitimate) and its build/convert
//     subpackages;
//   - the composition root /app/** — the place where the pool/connection is
//     opened and injected into repositories.
//
// Anywhere else — /client/**, /domain/**, /event/**, /server, /schedule,
// /job, /metric, /validate — an import of a driver means the store is being
// reached past the repository. The typical smell this rule was written for is
// a "client" that is in fact a key-value repository: internal/client/ratelimit
// on go-redis, internal/client/{dedup,magiclink,devicelogin} on eredis
// (govorun-server, 2026-07-25). Such a package belongs in /dal/repository, its
// stored shapes in /dal/entity.
//
// Driver detection is by import path prefix. settings.packages EXTENDS the
// default list (unlike gidconvpure, where it replaces it): the defaults are
// well-known drivers that stay relevant, while a service adds its own
// in-house wrapper (an internal eredis-style library) on top. An import path
// that is a false positive is silenced pinpoint with //nolint:gidstorageplace.
//
// Layers where the driver is allowed can be redefined with settings.allow
// (path segments, e.g. "dal" or "app"); the default is dal + app.
// Test files (_test.go) are skipped: a container-based integration test or a
// miniredis fixture legitimately touches the driver from any layer.
package storageplace

import (
	"go/ast"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-249"

// defaultDrivers — import path prefixes of storage drivers: SQL pools and
// drivers, key-value and NoSQL clients, and the gid.team libraries wrapping
// them. Extended (not replaced) by settings.packages.
var defaultDrivers = []string{
	// SQL
	"database/sql",
	"github.com/jackc/pgx",
	"github.com/lib/pq",
	"github.com/go-sql-driver/mysql",
	"github.com/mattn/go-sqlite3",
	"github.com/jmoiron/sqlx",
	"github.com/ClickHouse/clickhouse-go",
	"gid.team/gid-data/tech/golang/libs/postgres.git",
	"gid.team/gid-data/tech/golang/libs/clickhouse.git",
	// key-value / NoSQL
	"github.com/redis/go-redis",
	"github.com/redis/rueidis",
	"github.com/go-redis/redis",
	"github.com/gomodule/redigo",
	"github.com/valkey-io/valkey-go",
	"go.mongodb.org/mongo-driver",
	"go.etcd.io/etcd",
	"go.etcd.io/bbolt",
	"github.com/dgraph-io/badger",
	"github.com/gocql/gocql",
	"github.com/elastic/go-elasticsearch",
	"gid.team/gid-data/tech/golang/libs/redis.git",
}

// observabilitySubpaths — subpackages of a driver library that carry no
// storage access at all: the Prometheus collector or the logger adapter of an
// in-house wrapper (eredis/pkg/prometheus, epgx/pkg/prometheus,
// postgres.git/pkg/logrus). They are wired in /metric and /app, so a prefix
// match on the driver must not flag them.
var observabilitySubpaths = []string{
	"pkg/prometheus",
	"pkg/metrics",
	"pkg/logrus",
	"pkg/logger",
	"pkg/otel",
	"pkg/tracing",
}

// valueOnlySymbols — drivers with a second, harmless role: a package that also
// hosts value types used far from any connection. database/sql is the only
// such case — sql.NullString and friends legitimately appear in an entity and
// in a converter over it. An import used exclusively for these symbols is not
// storage access; the moment sql.DB/sql.Open shows up, the file is flagged.
var valueOnlySymbols = map[string]map[string]bool{
	"database/sql": {
		"NullBool": true, "NullByte": true, "NullFloat64": true, "NullInt16": true,
		"NullInt32": true, "NullInt64": true, "NullString": true, "NullTime": true,
		"Null": true, "RawBytes": true, "Scanner": true, "NamedArg": true, "Named": true,
		"ErrNoRows": true,
	},
}

// defaultAllowedLayers — layers allowed to import a storage driver: the whole
// dal layer (repository, entity, build, convert) and the composition root,
// which opens the pool and injects it into repositories.
var defaultAllowedLayers = [][]string{
	{"dal"},
	{"app"},
}

// Analyzer — GID-249 with the default driver list and layers.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Packages — additional storage drivers (import path prefixes), e.g. an
	// in-house redis wrapper. Extends the default list, does not replace it.
	Packages []string `json:"packages"`
	// Allow — layers (path segments) allowed to import a driver on top of
	// the default dal + app; a nested layer is written as "dal/repository".
	Allow []string `json:"allow"`
	// ExcludePackages — import path prefixes that must NOT count as a driver,
	// even when a default prefix matches them: an observability subpackage the
	// built-in list does not know, a pure query builder, a driver's own
	// type-only package. The escape hatch that keeps the driver list itself
	// untouched.
	ExcludePackages []string `json:"exclude-packages"`
	// ExcludePaths — "/"-joined path-segment sequences; a package whose import
	// path contains such a sequence is not checked at all (e.g.
	// "internal/legacy/cache"). An escape hatch for a concrete proven case,
	// not a blessing for a whole layer — for a layer there is Allow.
	ExcludePaths []string `json:"exclude-paths"`
}

// NewAnalyzer builds the GID-249 analyzer from the linter settings.
//
//nolint:gocritic // hugeParam: the plugin contract takes Settings by value (newConfigurablePlugin)
func NewAnalyzer(s Settings) *analysis.Analyzer {
	drivers := append(append([]string{}, defaultDrivers...), s.Packages...)
	allowed := append([][]string{}, defaultAllowedLayers...)
	for _, layer := range s.Allow {
		if segs := splitLayer(layer); len(segs) > 0 {
			allowed = append(allowed, segs)
		}
	}
	return &analysis.Analyzer{
		Name: "gidstorageplace",
		Doc: ruleID + ": a storage driver (SQL/key-value/NoSQL) is imported only by /dal/** and the " +
			"composition root /app/** — a data store is reached through a repository, a client is for an " +
			"external API",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, drivers, allowed, s.ExcludePaths, s.ExcludePackages)
		},
	}
}

func run(pass *analysis.Pass, drivers []string, allowed [][]string, excludePaths, excludePackages []string) (any, error) {
	pkgPath := pass.Pkg.Path()
	if inAllowedLayer(pkgPath, allowed) || excludedPath(pkgPath, excludePaths) {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) || isTestFile(pass, file) {
			continue
		}
		checkImports(pass, file, drivers, excludePackages)
	}
	return nil, nil
}

func checkImports(pass *analysis.Pass, file *ast.File, drivers, excluded []string) {
	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		if hasPrefixPath(path, excluded) || !isDriver(path, drivers) || usesValueSymbolsOnly(pass, file, path) {
			continue
		}
		pass.Reportf(imp.Pos(),
			"%s: package %q reaches a data store directly — driver %q belongs to the repository layer. "+
				"Fix: move the storage code to /dal/repository (its stored shapes to /dal/entity) and "+
				"inject the repository through an interface; the pool itself is opened in /app. "+
				"A /client package is for an external API (HTTP/gRPC/SMTP), not for a database",
			ruleID, pass.Pkg.Path(), path)
	}
}

func isDriver(path string, drivers []string) bool {
	if isObservabilitySubpath(path) {
		return false
	}
	return hasPrefixPath(path, drivers)
}

// hasPrefixPath reports whether path equals one of the prefixes or lies under
// it as a path segment ("github.com/redis/go-redis" matches
// ".../go-redis/v9" but not ".../go-redis-extra-tools").
func hasPrefixPath(path string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}

// excludedPath reports whether the package under analysis is exempted through
// settings.exclude-paths (a "/"-joined sequence of path segments).
func excludedPath(pkgPath string, excludes []string) bool {
	for _, e := range excludes {
		if pathseg.Contains(pkgPath, pathseg.Segments(strings.Trim(e, "/"))...) {
			return true
		}
	}
	return false
}

// isObservabilitySubpath reports whether path is the metrics/logging adapter of
// a driver library rather than the driver itself.
func isObservabilitySubpath(path string) bool {
	for _, sub := range observabilitySubpaths {
		if strings.HasSuffix(path, "/"+sub) {
			return true
		}
	}
	return false
}

// usesValueSymbolsOnly reports whether the file touches the imported package
// exclusively through its harmless value symbols (see valueOnlySymbols) — an
// entity converter over sql.NullString is not storage access.
func usesValueSymbolsOnly(pass *analysis.Pass, file *ast.File, path string) bool {
	allowed, ok := valueOnlySymbols[path]
	if !ok {
		return false
	}
	onlyValues := true
	ast.Inspect(file, func(n ast.Node) bool {
		id, isIdent := n.(*ast.Ident)
		if !isIdent {
			return true
		}
		obj := pass.TypesInfo.Uses[id]
		if obj == nil {
			return true
		}
		objPkg := obj.Pkg()
		if objPkg == nil || objPkg.Path() != path {
			return true
		}
		// Struct fields and methods (sql.NullString.Valid) belong to the package
		// too — only package-level symbols say what the import is used for.
		if obj.Parent() != objPkg.Scope() {
			return true
		}
		if !allowed[obj.Name()] {
			onlyValues = false
		}
		return true
	})
	return onlyValues
}

func inAllowedLayer(pkgPath string, allowed [][]string) bool {
	for _, layer := range allowed {
		if pathseg.HasLayer(pkgPath, layer...) {
			return true
		}
	}
	return false
}

// isTestFile reports whether the file is a _test.go one: an integration test
// or a fixture legitimately talks to the driver from any layer.
func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	return strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, "_test.go")
}

// splitLayer turns a "dal/repository" setting into path segments.
func splitLayer(layer string) []string {
	var out []string
	for seg := range strings.SplitSeq(layer, "/") {
		if seg != "" {
			out = append(out, seg)
		}
	}
	return out
}
