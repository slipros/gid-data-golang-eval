package sqlstack

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
)

const (
	repoFile   = "internal/dal/repository/user.go"
	entityFile = "internal/dal/entity/user.go"
	appFile    = "internal/app/app.go"

	caseGRPCBacked = "a dal backed by grpc, no sql anywhere"

	libSQLX    = "github.com/jmoiron/sqlx"
	libStdSQL  = "database/sql"
	libPGX     = "github.com/jackc/pgx"
	libPGXPool = libPGX + "/v5/pgxpool"
	libGRPC    = "google.golang.org/grpc"
	libWrapper = "gid.team/libs/pgstore"

	svcModule  = "example.com/svc"
	svcPkgPath = svcModule + "/internal/dal/entity"
)

// file — one source file of the fixture module: its path and the import it
// carries.
type file struct {
	path    string
	imports []string
}

func TestScan(t *testing.T) {
	tests := []struct {
		name    string
		files   []file
		imports []string
		want    bool
	}{
		{
			name:  "sqlx in the repository",
			files: []file{{path: repoFile, imports: []string{libSQLX}}, {path: entityFile}},
			want:  true,
		},
		{
			name:  "database/sql in the composition root only",
			files: []file{{path: appFile, imports: []string{libStdSQL}}, {path: entityFile}},
			want:  true,
		},
		{
			name:  "pgx at a major version, inside a subpackage",
			files: []file{{path: repoFile, imports: []string{libPGXPool}}},
			want:  true,
		},
		{
			name:  "the clickhouse driver",
			files: []file{{path: repoFile, imports: []string{"github.com/ClickHouse/clickhouse-go/v2"}}},
			want:  true,
		},
		{
			name:  caseGRPCBacked,
			files: []file{{path: repoFile, imports: []string{libGRPC, "svc/genproto/consentapi"}}, {path: entityFile}},
			want:  false,
		},
		{
			name:  "sql reached for by a test file only",
			files: []file{{path: "internal/dal/repository/user_test.go", imports: []string{libSQLX}}, {path: entityFile}},
			want:  false,
		},
		{
			name:  "sql inside vendor — another module's sources",
			files: []file{{path: "vendor/github.com/jmoiron/sqlx/sqlx.go", imports: []string{libStdSQL}}, {path: entityFile}},
			want:  false,
		},
		{
			name:  "sql inside testdata — a fixture imports anything",
			files: []file{{path: "internal/dal/testdata/src/db/db.go", imports: []string{libStdSQL}}, {path: entityFile}},
			want:  false,
		},
		{
			name:    "an in-house wrapper listed in the settings",
			files:   []file{{path: repoFile, imports: []string{libWrapper}}},
			imports: []string{libWrapper},
			want:    true,
		},
		{
			name:  "an in-house wrapper the settings do not list",
			files: []file{{path: repoFile, imports: []string{libWrapper}}},
			want:  false,
		},
		{name: "an empty module", want: false},
	}

	for _, tt := range tests { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		t.Run(tt.name, func(t *testing.T) {
			root := writeModule(t, tt.files)

			imports := tt.imports
			if len(imports) == 0 {
				imports = Default
			}

			if got := scan(root, imports); got != tt.want {
				t.Errorf("scan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsLibrary(t *testing.T) {
	tests := []struct {
		name       string
		importPath string
		library    string
		want       bool
	}{
		{name: "the library itself", importPath: libSQLX, library: libSQLX, want: true},
		{name: "a major version of it", importPath: libPGX + "/v5", library: libPGX, want: true},
		{name: "a subpackage", importPath: libStdSQL + "/driver", library: libStdSQL, want: true},
		{name: "a subpackage of a major version", importPath: libPGXPool, library: libPGX, want: true},
		{name: "a shared prefix that is not a path prefix", importPath: "database/sqlite", library: libStdSQL, want: false},
		{name: "another library", importPath: libGRPC, library: libSQLX, want: false},
	}

	for _, tt := range tests { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		t.Run(tt.name, func(t *testing.T) {
			if got := isLibrary(tt.importPath, tt.library); got != tt.want {
				t.Errorf("isLibrary(%q, %q) = %v, want %v", tt.importPath, tt.library, got, tt.want)
			}
		})
	}
}

// writeModule lays the fixture files out under a temporary module root.
func writeModule(t *testing.T, files []file) string {
	t.Helper()

	root := t.TempDir()
	for _, f := range files { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		path := filepath.Join(root, filepath.FromSlash(f.path))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", f.path, err)
		}
		if err := os.WriteFile(path, []byte(source(f.imports)), 0o600); err != nil {
			t.Fatalf("write %s: %v", f.path, err)
		}
	}

	return root
}

// source builds a minimal Go file carrying the given imports.
func source(imports []string) string {
	var out strings.Builder

	out.WriteString("package fixture\n")
	for _, imp := range imports {
		out.WriteString("\nimport \"" + imp + "\"\n")
	}

	return out.String()
}

// TestHasStack — the entry point the rule calls: the verdict is read from the
// module the package belongs to, and a package with no module of its own keeps
// the pre-gate behaviour.
func TestHasStack(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, filepath.FromSlash("internal/dal/entity"))

	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module "+svcModule+"\n"), 0o600); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	writeFile(t, filepath.Join(root, filepath.FromSlash(repoFile)), source([]string{libGRPC}))

	t.Run("a_module_without_the_sql_stack", func(t *testing.T) {
		if HasStack(newPass(t, pkgDir, svcPkgPath), nil) {
			t.Error("HasStack() = true, want false for a module importing no sql")
		}
	})

	t.Run("the_wrapper_named_in_the_settings", func(t *testing.T) {
		writeFile(t, filepath.Join(root, filepath.FromSlash("internal/dal/repository/store.go")), source([]string{libWrapper}))

		if !HasStack(newPass(t, pkgDir, svcPkgPath), []string{libWrapper}) {
			t.Error("HasStack() = false, want true with the wrapper on the list")
		}
	})

	t.Run("a_package_outside_any_module", func(t *testing.T) {
		outside := t.TempDir()
		if !HasStack(newPass(t, outside, "svc/dal/entity"), nil) {
			t.Error("HasStack() = false, want true without a module root to read")
		}
	})
}

// newPass builds a minimal pass whose single file sits in pkgDir.
func newPass(t *testing.T, pkgDir, pkgPath string) *analysis.Pass {
	t.Helper()

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, filepath.Join(pkgDir, "pkg.go"), "package pkg\n", parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	return &analysis.Pass{
		Fset:  fset,
		Files: []*ast.File{file},
		Pkg:   types.NewPackage(pkgPath, "pkg"),
	}
}

// writeFile lays one source file down, creating its directory.
func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
