package defaultconfig

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// The pieces every command line in this test is built from.
const (
	bin        = "custom-gcl"
	allPkgs    = "./..."
	verbose    = "-v"
	configName = ".golangci.yml"
	userConfig = "my.yml"
)

// builtIn — a config small enough to compare byte by byte; the real one is the
// embedded gid-golangci.yml, and Inject never looks inside it.
var builtIn = []byte("version: \"2\"\n")

func TestInject(t *testing.T) {
	const (
		commandVersion = "version"
		commandCache   = "cache"
		commandHelp    = "help"
	)

	tests := []struct {
		name string
		// configs — config files created before the run, relative to the
		// working directory ("../" places one in the parent).
		configs []string
		// home — a config file created in $HOME.
		home bool
		args []string
		want bool
	}{
		// positive — nothing tells golangci-lint where its config is.
		{name: "run without a config", args: []string{bin, commandRun, allPkgs}, want: true},
		{name: "run without arguments", args: []string{bin, commandRun}, want: true},
		{name: "fmt without a config", args: []string{bin, commandFmt, allPkgs}, want: true},
		{name: "config verify without a config", args: []string{bin, commandConfig, "verify"}, want: true},
		{name: "flag before the subcommand", args: []string{bin, verbose, commandRun}, want: true},
		// negative — a config file is already there.
		{name: "config in the working directory", configs: []string{configName}, args: []string{bin, commandRun, allPkgs}},
		{name: "yaml extension", configs: []string{".golangci.yaml"}, args: []string{bin, commandRun, allPkgs}},
		{name: "toml extension", configs: []string{".golangci.toml"}, args: []string{bin, commandRun, allPkgs}},
		{name: "config above the working directory", configs: []string{"../" + configName}, args: []string{bin, commandRun, allPkgs}},
		// The search starts from the path arguments too, not only from the
		// working directory — that is where golangci-lint would start.
		{name: "config above the path argument", configs: []string{"../other/" + configName}, args: []string{bin, commandRun, "../other/..."}},
		{name: "config in the home directory", home: true, args: []string{bin, commandRun, allPkgs}},
		// negative — the user decided about the config themselves.
		{name: "explicit config", args: []string{bin, commandRun, flagConfig, userConfig}},
		{name: "explicit config with an equals sign", args: []string{bin, commandRun, flagConfig + "=" + userConfig}},
		{name: "short config flag", args: []string{bin, commandRun, flagConfigShort, userConfig}},
		{name: "config disabled", args: []string{bin, commandRun, flagNoConfig}},
		// non-applicability — subcommands that read no config file.
		{name: commandVersion, args: []string{bin, commandVersion}},
		{name: "cache clean", args: []string{bin, commandCache, "clean"}},
		{name: commandHelp, args: []string{bin, commandHelp}},
		{name: "no subcommand", args: []string{bin}},
	}

	for _, tt := range tests { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		t.Run(tt.name, func(t *testing.T) {
			home := sandbox(t)
			if tt.home {
				write(t, filepath.Join(home, configName))
			}

			for _, config := range tt.configs {
				write(t, config)
			}

			args, path, err := Inject(tt.args, builtIn)
			if err != nil {
				t.Fatalf("Inject(%v) error: %v", tt.args, err)
			}

			if injected := path != ""; injected != tt.want {
				t.Fatalf("Inject(%v) injected = %v, want %v (path %q)", tt.args, injected, tt.want, path)
			}

			if !tt.want {
				if len(args) != len(tt.args) {
					t.Errorf("Inject(%v) changed the arguments: %v", tt.args, args)
				}

				return
			}

			assertConfigFlag(t, args, path)
		})
	}
}

// TestInjectPlacesFlagAfterSubcommand — --config belongs to the subcommand, so
// it has to sit after it and leave the remaining arguments in order.
func TestInjectPlacesFlagAfterSubcommand(t *testing.T) {
	sandbox(t)

	args, path, err := Inject([]string{bin, commandRun, verbose, allPkgs}, builtIn)
	if err != nil {
		t.Fatalf("Inject error: %v", err)
	}

	want := []string{bin, commandRun, flagConfig + "=" + path, verbose, allPkgs}
	if strings.Join(args, " ") != strings.Join(want, " ") {
		t.Errorf("Inject = %v, want %v", args, want)
	}
}

// TestInjectWritesConfigOnce — a run that already has the right config in the
// cache reuses the file instead of rewriting it.
func TestInjectWritesConfigOnce(t *testing.T) {
	sandbox(t)

	_, first, err := Inject([]string{bin, commandRun}, builtIn)
	if err != nil {
		t.Fatalf("first Inject error: %v", err)
	}

	_, second, err := Inject([]string{bin, commandRun}, builtIn)
	if err != nil {
		t.Fatalf("second Inject error: %v", err)
	}

	if first != second {
		t.Errorf("the config path changed between runs: %q then %q", first, second)
	}

	assertContent(t, first, builtIn)
}

// TestInjectReplacesOldConfig — a new binary carries a new config; the file in
// the cache is the previous one and has to be rewritten in place, not joined by
// a second copy.
func TestInjectReplacesOldConfig(t *testing.T) {
	sandbox(t)

	_, path, err := Inject([]string{bin, commandRun}, builtIn)
	if err != nil {
		t.Fatalf("first Inject error: %v", err)
	}

	next := []byte("version: \"2\"\nlinters:\n  enable:\n    - gidtimenow\n")

	_, nextPath, err := Inject([]string{bin, commandRun}, next)
	if err != nil {
		t.Fatalf("second Inject error: %v", err)
	}

	if nextPath != path {
		t.Errorf("the new config went to %q instead of %q", nextPath, path)
	}

	assertContent(t, path, next)

	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatalf("read the cache directory: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("the cache directory holds %d files, want the single config", len(entries))
	}
}

// TestInjectWithoutConfig — a binary built without the embedded config keeps
// the stock golangci-lint behaviour instead of writing an empty file.
func TestInjectWithoutConfig(t *testing.T) {
	sandbox(t)

	args, path, err := Inject([]string{bin, commandRun}, nil)
	if err != nil {
		t.Fatalf("Inject error: %v", err)
	}

	if path != "" || len(args) != 2 {
		t.Errorf("Inject with an empty config = %v, %q, want the arguments unchanged", args, path)
	}
}

// sandbox isolates the run from the developer's own files: the working
// directory moves into a temporary tree with no config above it, and both the
// home and the cache directory point inside that tree. It returns the home
// directory.
func sandbox(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	work := filepath.Join(root, "work", "module")

	for _, dir := range []string{home, work} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create %s: %v", dir, err)
		}
	}

	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(root, "cache"))
	t.Chdir(work)

	return home
}

func write(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create the directory of %s: %v", path, err)
	}

	if err := os.WriteFile(path, builtIn, 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertContent(t *testing.T, path string, want []byte) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read the written config: %v", err)
	}

	if !bytes.Equal(data, want) {
		t.Errorf("the written config is %q, want %q", data, want)
	}
}

func assertConfigFlag(t *testing.T, args []string, path string) {
	t.Helper()

	flag := "--config=" + path
	if !slices.Contains(args, flag) {
		t.Errorf("%v carries no %s", args, flag)
	}
}
