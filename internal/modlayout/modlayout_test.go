package modlayout

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis"
)

const (
	svcModule  = "example.com/svc"
	svcPkgPath = "svc/domain/service"

	dirDomain        = "internal/domain"
	dirDomainModel   = "domain/model"
	dirLogger        = "pkg/logger"
	dirPrometheus    = "pkg/prometheus"
	dirAppAPI        = "internal/app/api"
	dirDALRepository = "internal/dal/repository"
	dirServerHTTP    = "internal/server/http"

	caseFlatLibrary = "flat library"
	caseEmptyModule = "empty module"
	caseOtherModule = "package of another module than the go.mod found above it"
)

// goModContents — built once: converting the same string to []byte inside the
// table loop repeats the allocation (GID-182).
var goModContents = []byte("module " + svcModule + "\n\ngo 1.24\n")

func TestHasServiceDirs(t *testing.T) {
	tests := []struct {
		name string
		dirs []string
		want bool
	}{
		{name: "service under internal", dirs: []string{dirDomain + "/service", dirAppAPI}, want: true},
		{name: "service at the top level", dirs: []string{dirDomainModel, "dal/repository"}, want: true},
		{name: "only the composition root", dirs: []string{"internal/app"}, want: true},
		{name: "domain and dal without an app", dirs: []string{dirDomain, "internal/dal"}, want: true},
		{name: "flat library", dirs: []string{dirLogger, "pkg/prometheus", "example"}, want: false},
		{name: "library with an internal of its own", dirs: []string{"internal/pool", "internal/retry"}, want: false},
		{name: "transport library with a server and a client", dirs: []string{"server/middleware", "client/serde", dirLogger}, want: false},
		{name: "library with a domain but no dal", dirs: []string{dirDomain, "client/interceptor"}, want: false},
		{name: "library publishing its own app package", dirs: []string{"app", "errors", "mapper"}, want: false},
		{name: caseEmptyModule, want: false},
	}

	for _, tt := range tests { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			for _, dir := range tt.dirs {
				if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755); err != nil {
					t.Fatalf("mkdir %s: %v", dir, err)
				}
			}

			if got := hasServiceDirs(root); got != tt.want {
				t.Errorf("hasServiceDirs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasDataDirs(t *testing.T) {
	tests := []struct {
		name string
		dirs []string
		want bool
	}{
		{name: "dal under internal", dirs: []string{dirDALRepository, dirDomain + "/service"}, want: true},
		{name: "dal at the top level", dirs: []string{"dal/entity", dirDomainModel}, want: true},
		{name: "a dal holding entities only", dirs: []string{"internal/dal/entity"}, want: true},
		{name: "a repository without a dal", dirs: []string{"internal/repository", dirDomain}, want: true},
		{name: "bff: a service and a transport, no data layer", dirs: []string{dirDomain + "/service", dirServerHTTP, dirAppAPI}, want: false},
		{name: "a repository nested below another layer is not the data layer", dirs: []string{"internal/server/grpc/repository"}, want: false},
		{name: caseFlatLibrary, dirs: []string{dirLogger, dirPrometheus}, want: false},
		{name: caseEmptyModule, want: false},
	}

	for _, tt := range tests { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			for _, dir := range tt.dirs {
				if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755); err != nil {
					t.Fatalf("mkdir %s: %v", dir, err)
				}
			}

			if got := hasDataDirs(root); got != tt.want {
				t.Errorf("hasDataDirs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBelongsTo(t *testing.T) {
	tests := []struct {
		name   string
		pkg    string
		module string
		want   bool
	}{
		{name: "the module root package", pkg: svcModule, module: svcModule, want: true},
		{name: "a package of the module", pkg: svcModule + "/internal/dal", module: svcModule, want: true},
		{name: "an analysistest fixture inside another module", pkg: svcPkgPath, module: "github.com/slipros/gid-data-golang-eval", want: false},
		{name: "a shared prefix that is not a path prefix", pkg: "example.com/svc-two/dal", module: svcModule, want: false},
		{name: "an unknown module path", pkg: svcPkgPath, module: "", want: true},
		{name: "no type information, so no package path", pkg: "", module: svcModule, want: true},
	}

	for _, tt := range tests { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		t.Run(tt.name, func(t *testing.T) {
			if got := belongsTo(tt.pkg, tt.module); got != tt.want {
				t.Errorf("belongsTo(%q, %q) = %v, want %v", tt.pkg, tt.module, got, tt.want)
			}
		})
	}
}

func TestModuleRoot(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "internal", "domain", "service")

	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module "+svcModule+"\n"), 0o600); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	t.Run("walks_up_to_the_go_mod", func(t *testing.T) {
		got, modPath, ok := moduleRoot(nested)
		if !ok || got != root {
			t.Errorf("moduleRoot() = %q, %v; want %q, true", got, ok, root)
		}
		if modPath != svcModule {
			t.Errorf("moduleRoot() module path = %q, want %q", modPath, svcModule)
		}
	})

	t.Run("a_submodule_wins_over_its_parent", func(t *testing.T) {
		sub := filepath.Join(root, "pkg", "logger")
		if err := os.MkdirAll(sub, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(sub, "go.mod"), []byte("module example.com/svc/pkg/logger\n"), 0o600); err != nil {
			t.Fatalf("write go.mod: %v", err)
		}

		got, modPath, ok := moduleRoot(sub)
		if !ok || got != sub {
			t.Errorf("moduleRoot() = %q, %v; want %q, true", got, ok, sub)
		}
		if modPath != "example.com/svc/pkg/logger" {
			t.Errorf("moduleRoot() module path = %q, want the submodule's own", modPath)
		}
	})

	t.Run("no_go_mod_above_the_directory", func(t *testing.T) {
		orphan := t.TempDir()
		if _, _, ok := moduleRoot(orphan); ok {
			t.Errorf("moduleRoot() reported a root for a directory outside any module")
		}
	})
}

// TestIsServiceModule builds a pass over a real directory tree: the verdict is
// read off the module root found above the package, so the go.mod and the layer
// directories have to exist on disk.
func TestIsServiceModule(t *testing.T) {
	tests := []struct {
		name    string
		dirs    []string
		pkgDir  string
		pkgPath string
		want    bool
	}{
		{
			name:    "service module",
			dirs:    []string{dirDomain + "/service", dirDALRepository},
			pkgDir:  dirDomain + "/service",
			pkgPath: svcModule + "/" + dirDomain + "/service",
			want:    true,
		},
		{
			name:    caseFlatLibrary,
			dirs:    []string{dirLogger, dirPrometheus},
			pkgDir:  dirLogger,
			pkgPath: svcModule + "/" + dirLogger,
			want:    false,
		},
		{
			name:    caseOtherModule,
			dirs:    []string{dirLogger},
			pkgDir:  dirLogger,
			pkgPath: "other.example.com/lib/logger",
			want:    true, // nothing reliable to read: the pre-existing behaviour is kept
		},
	}

	for _, tt := range tests { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			for _, dir := range tt.dirs {
				if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755); err != nil {
					t.Fatalf("mkdir %s: %v", dir, err)
				}
			}
			if err := os.WriteFile(filepath.Join(root, "go.mod"), goModContents, 0o600); err != nil {
				t.Fatalf("write go.mod: %v", err)
			}

			pass := newPass(t, filepath.Join(root, filepath.FromSlash(tt.pkgDir)), tt.pkgPath)
			if got := IsServiceModule(pass); got != tt.want {
				t.Errorf("IsServiceModule() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsServiceModuleNoFiles — a pass with no files at all has no directory to
// inspect, and the rule that asks must not be switched off by that.
func TestIsServiceModuleNoFiles(t *testing.T) {
	pass := &analysis.Pass{
		Fset: token.NewFileSet(),
		Pkg:  types.NewPackage(svcPkgPath, "service"),
	}

	if !IsServiceModule(pass) {
		t.Error("IsServiceModule() = false, want true for a pass with no files")
	}
}

// TestHasDataLayer builds a pass the same way: a BFF module — a service and a
// transport and no data layer — gets false, and a module with a dal gets true.
func TestHasDataLayer(t *testing.T) {
	tests := []struct {
		name    string
		dirs    []string
		pkgDir  string
		pkgPath string
		want    bool
	}{
		{
			name:    "service module with a dal",
			dirs:    []string{dirDomain + "/service", dirDALRepository},
			pkgDir:  dirDomain + "/service",
			pkgPath: svcModule + "/" + dirDomain + "/service",
			want:    true,
		},
		{
			name:    "bff without a data layer",
			dirs:    []string{dirDomain + "/service", dirServerHTTP, dirAppAPI},
			pkgDir:  dirDomain + "/service",
			pkgPath: svcModule + "/" + dirDomain + "/service",
			want:    false,
		},
		{
			name:    caseOtherModule,
			dirs:    []string{dirDomain + "/service"},
			pkgDir:  dirDomain + "/service",
			pkgPath: "other.example.com/lib/service",
			want:    true, // nothing reliable to read: the rule keeps judging
		},
	}

	for _, tt := range tests { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			for _, dir := range tt.dirs {
				if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(dir)), 0o755); err != nil {
					t.Fatalf("mkdir %s: %v", dir, err)
				}
			}
			if err := os.WriteFile(filepath.Join(root, "go.mod"), goModContents, 0o600); err != nil {
				t.Fatalf("write go.mod: %v", err)
			}

			pass := newPass(t, filepath.Join(root, filepath.FromSlash(tt.pkgDir)), tt.pkgPath)
			if got := HasDataLayer(pass); got != tt.want {
				t.Errorf("HasDataLayer() = %v, want %v", got, tt.want)
			}
		})
	}
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
