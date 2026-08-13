package defaultconfig

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// The pieces every command line in this test is built from.
const (
	bin        = "custom-gcl"
	allPkgs    = "./..."
	verbose    = "-v"
	configName = ".golangci.yml"
	yamlConfig = ".golangci.yaml"
	tomlConfig = ".golangci.toml"
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
		// positive — the choice of config is left to the binary.
		{name: "run", args: []string{bin, commandRun, allPkgs}, want: true},
		{name: "run without arguments", args: []string{bin, commandRun}, want: true},
		{name: "fmt", args: []string{bin, commandFmt, allPkgs}, want: true},
		{name: "config verify", args: []string{bin, commandConfig, "verify"}, want: true},
		{name: "flag before the subcommand", args: []string{bin, verbose, commandRun}, want: true},
		// positive — a repository config does not win by simply being there:
		// custom-gcl is the gid ruleset, and a plain run applies it.
		{name: "config in the working directory", configs: []string{configName}, args: []string{bin, commandRun, allPkgs}, want: true},
		{name: "config above the working directory", configs: []string{"../" + configName}, args: []string{bin, commandRun, allPkgs}, want: true},
		{name: "config in the home directory", home: true, args: []string{bin, commandRun, allPkgs}, want: true},
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

// TestInjectSeparatesDifferentConfigs — the cache file is named after the
// content, so a different config gets a different file instead of overwriting
// the one already there. That matters because the binary now carries two
// configs (the full ruleset and the --gid-rules-only one) and
// run.allow-parallel-runners lets agents lint at the same time: sharing a path
// would swap the config out from under a run that already pointed --config at
// it. Each file keeps its own content, and neither run is disturbed.
func TestInjectSeparatesDifferentConfigs(t *testing.T) {
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

	if nextPath == path {
		t.Fatalf("both configs went to the same file %q — a parallel run of the other one would read "+
			"a config it did not ask for", path)
	}

	assertContent(t, path, builtIn)
	assertContent(t, nextPath, next)

	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatalf("read the cache directory: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("the cache directory holds %d files, want one per config", len(entries))
	}
}

// TestInjectSweepsStaleConfigs — a content-addressed name means an upgraded
// binary writes a new file and leaves the old one behind; nothing ever reads it
// again, and seven revisions had piled up in a real cache before anybody
// looked. A run collects them, and leaves alone everything it must: the config
// it is about to use, a config of the same age (the --gid-rules-only one, in
// daily use next to it) and files that are not ours.
func TestInjectSweepsStaleConfigs(t *testing.T) {
	sandbox(t)

	_, path, err := Inject([]string{bin, commandRun}, builtIn)
	if err != nil {
		t.Fatalf("Inject error: %v", err)
	}

	dir := filepath.Dir(path)
	stale := filepath.Join(dir, cachePrefix+"0123456789abcdef.yml")
	leftover := filepath.Join(dir, cachePrefix+"0123456789abcdef.yml.tmp42")
	foreign := filepath.Join(dir, "notes.txt")

	for _, name := range []string{stale, leftover, foreign} {
		if writeErr := os.WriteFile(name, builtIn, 0o600); writeErr != nil {
			t.Fatalf("seed %s: %v", name, writeErr)
		}

		//nolint:gidtimenow // the plugin does not depend on the internal gdhelper library
		old := time.Now().Add(-cacheGrace - time.Hour)
		if chErr := os.Chtimes(name, old, old); chErr != nil {
			t.Fatalf("age %s: %v", name, chErr)
		}
	}

	// A second config of the same binary, materialized just now: fresh, and it
	// must survive the sweep the first one performs.
	fresh := []byte("version: \"2\"\nlinters:\n  enable:\n    - gidtimenow\n")

	_, freshPath, err := Inject([]string{bin, commandRun}, fresh)
	if err != nil {
		t.Fatalf("Inject of the second config error: %v", err)
	}

	if _, statErr := os.Stat(stale); !os.IsNotExist(statErr) {
		t.Errorf("the config of an earlier binary survived the sweep: %v", statErr)
	}

	if _, statErr := os.Stat(leftover); !os.IsNotExist(statErr) {
		t.Errorf("a temporary of an interrupted write survived the sweep: %v", statErr)
	}

	for _, name := range []string{path, freshPath, foreign} {
		if _, statErr := os.Stat(name); statErr != nil {
			t.Errorf("the sweep removed %s, which is in use or not ours: %v", name, statErr)
		}
	}
}

// TestInjectKeepsConfigInDailyUse — a config older than the grace period is
// still the one this run needs, so it is kept and its age reset. Without the
// reset a binary in daily use would have its own config swept out from under
// the next run the moment it crossed the grace period.
func TestInjectKeepsConfigInDailyUse(t *testing.T) {
	sandbox(t)

	_, path, err := Inject([]string{bin, commandRun}, builtIn)
	if err != nil {
		t.Fatalf("first Inject error: %v", err)
	}

	//nolint:gidtimenow // the plugin does not depend on the internal gdhelper library
	old := time.Now().Add(-cacheGrace - time.Hour)
	if chErr := os.Chtimes(path, old, old); chErr != nil {
		t.Fatalf("age the config: %v", chErr)
	}

	_, again, err := Inject([]string{bin, commandRun}, builtIn)
	if err != nil {
		t.Fatalf("second Inject error: %v", err)
	}

	if again != path {
		t.Fatalf("the config path changed between runs: %q then %q", path, again)
	}

	assertContent(t, path, builtIn)

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat the config: %v", err)
	}

	if !info.ModTime().After(old) {
		t.Error("the config in use kept its old timestamp — the next sweep would delete it")
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

// TestLocalConfig — the notice names the repository config the run steps over,
// and says nothing when there is none.
func TestLocalConfig(t *testing.T) {
	tests := []struct {
		name    string
		created string
		want    string
	}{
		{name: "yml", created: configName, want: configName},
		{name: "yaml", created: yamlConfig, want: yamlConfig},
		{name: "toml", created: tomlConfig, want: tomlConfig},
		{name: "none", created: "", want: ""},
		{name: "only above the working directory", created: "../" + configName, want: ""},
	}

	for _, tt := range tests { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		t.Run(tt.name, func(t *testing.T) {
			sandbox(t)

			if tt.created != "" {
				write(t, tt.created)
			}

			if got := LocalConfig(); got != tt.want {
				t.Errorf("LocalConfig() = %q, want %q", got, tt.want)
			}
		})
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

	flag := flagConfig + "=" + path
	if !slices.Contains(args, flag) {
		t.Errorf("%v carries no %s", args, flag)
	}
}
