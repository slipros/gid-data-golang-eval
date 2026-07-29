package modlayout

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	svcModule  = "example.com/svc"
	svcPkgPath = "svc/domain/service"

	dirDomain = "internal/domain"
	dirLogger = "pkg/logger"
)

func TestHasServiceDirs(t *testing.T) {
	tests := []struct {
		name string
		dirs []string
		want bool
	}{
		{name: "service under internal", dirs: []string{dirDomain + "/service", "internal/app/api"}, want: true},
		{name: "service at the top level", dirs: []string{"domain/model", "dal/repository"}, want: true},
		{name: "only the composition root", dirs: []string{"internal/app"}, want: true},
		{name: "domain and dal without an app", dirs: []string{dirDomain, "internal/dal"}, want: true},
		{name: "flat library", dirs: []string{dirLogger, "pkg/prometheus", "example"}, want: false},
		{name: "library with an internal of its own", dirs: []string{"internal/pool", "internal/retry"}, want: false},
		{name: "transport library with a server and a client", dirs: []string{"server/middleware", "client/serde", dirLogger}, want: false},
		{name: "library with a domain but no dal", dirs: []string{dirDomain, "client/interceptor"}, want: false},
		{name: "library publishing its own app package", dirs: []string{"app", "errors", "mapper"}, want: false},
		{name: "empty module", want: false},
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
