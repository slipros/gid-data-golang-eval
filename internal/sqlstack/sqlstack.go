// Package sqlstack answers one question for a rule: does the module under
// analysis actually speak SQL?
//
// GID-125 demands a db column tag on every exported field of an entity. The
// layer path alone does not earn that demand: a /dal says where a repository
// WOULD live, not that there is a database behind it. consent-webhook-trigger
// implements /dal/repository over gRPC (genproto/consentapi) and fills its
// entities from protobuf — the 28 diagnostics on entity.Document, Profile and
// Webhook asked for tags documenting a mapping that never reaches a database.
//
// The verdict is per module, not per package: at least one of the module's own
// files must import the SQL stack. It is read from the IMPORTS IN CODE, not
// from go.mod — a dependency there can be transitive, pulled in by a library
// the service never calls. A SQL import in any package turns the rule on for
// every /dal/entity of that module: the repository is what talks to the
// database, and it sits next door to the entity it maps.
//
// _test.go files are not counted — a fixture may reach for sqlmock in a
// service that stores nothing. Generated code IS counted: sqlc and sqlboiler
// generate exactly the SQL layer the rule exists for.
//
// The answer is cached per module root (like modlayout), so the walk happens
// once per module per run. When there is no go.mod above the package — an
// analysistest fixture, a package compiled outside a module — the answer is
// "speaks SQL", so the rule keeps the behaviour it had before this gate.
package sqlstack

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/modlayout"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

// Default — the SQL stack of a gid.team service: the standard library, the
// drivers, the query builders and the ORMs. A module importing none of them
// has no database to map an entity onto. An in-house wrapper hiding the stack
// is added through settings (sql-imports).
var Default = []string{
	"database/sql",
	"github.com/jmoiron/sqlx",
	"github.com/jackc/pgx",
	"github.com/lib/pq",
	"github.com/go-sql-driver/mysql",
	"github.com/Masterminds/squirrel",
	"github.com/ClickHouse/clickhouse-go",
	"github.com/uptrace/bun",
	"gorm.io/gorm",
	"entgo.io/ent",
}

// skipDirs — directories holding no code of the module itself: another
// module's sources (vendor), fixtures deliberately importing anything
// (testdata), and the tooling directories.
var skipDirs = map[string]struct{}{
	"vendor":       {},
	"testdata":     {},
	"node_modules": {},
}

// cache — module root + the import list -> verdict.
var cache sync.Map

// HasStack reports whether the module under analysis imports the SQL stack in
// at least one of its own non-test files. An empty imports list means the
// default stack.
func HasStack(pass *analysis.Pass, imports []string) bool {
	if len(imports) == 0 {
		imports = Default
	}

	root, ok := modlayout.Root(pass)
	if !ok {
		return true // nothing to read: keep the pre-gate behaviour
	}

	key := root + "\x00" + strings.Join(imports, ",")
	if v, cached := cache.Load(key); cached {
		if verdict, isBool := v.(bool); isBool {
			return verdict
		}
	}

	verdict := scan(root, imports)
	cache.Store(key, verdict)

	return verdict
}

// scan walks the module for the first file importing the stack — the walk
// stops on the first hit, a module that speaks SQL usually says so early.
// A walk that fails outright leaves the module unread, and the answer falls
// back to "speaks SQL": the behaviour the rule had before this gate.
//
// The sentinels fs.SkipDir and fs.SkipAll are returned bare on purpose —
// WalkDir compares them by identity, and wrapping them (GID-177) would turn
// the control flow into a real error.
func scan(root string, imports []string) bool {
	found := false

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fs.SkipDir //nolint:gidstaticerr // a WalkDir sentinel: an unreadable directory is stepped over, not reported
		}
		if entry.IsDir() {
			if path != root && skipDir(entry.Name()) {
				return fs.SkipDir //nolint:gidstaticerr // a WalkDir sentinel, compared by identity
			}

			return nil
		}
		if !isSource(entry.Name()) || !fileImports(path, imports) {
			return nil
		}
		found = true

		return fs.SkipAll //nolint:gidstaticerr // a WalkDir sentinel: the first hit ends the walk
	})
	if err != nil {
		return true
	}

	return found
}

func skipDir(name string) bool {
	if _, ok := skipDirs[name]; ok {
		return true
	}

	return strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

// isSource reports whether the file is a Go source file of the module itself —
// a _test.go file is scaffolding and does not make the module speak SQL.
func isSource(name string) bool {
	return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
}

// fileImports parses the import block alone (parser.ImportsOnly) and reports
// whether one of them is the SQL stack.
func fileImports(path string, imports []string) bool {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		return false // a file that does not parse says nothing about the module
	}

	for _, spec := range file.Imports {
		imported, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		for _, lib := range imports {
			if isLibrary(imported, lib) {
				return true
			}
		}
	}

	return false
}

// isLibrary reports whether the import path is the library or a package of it:
// pgx is imported as github.com/jackc/pgx/v5/pgxpool, sqlx as itself, the
// ClickHouse driver as .../clickhouse-go/v2 — the major-version suffix is
// handled by pathseg.SameLibrary, the subpackage by the prefix.
func isLibrary(importPath, library string) bool {
	return pathseg.SameLibrary(importPath, library) || strings.HasPrefix(importPath, library+"/")
}
