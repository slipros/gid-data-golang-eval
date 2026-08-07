// Package defaultconfig makes custom-gcl usable in a repository that has no
// golangci-lint config of its own: the gid ruleset ships inside the binary and
// is used when — and only when — golangci-lint would otherwise find no config.
//
// golangci-lint reads exactly one config file and has no notion of a built-in
// default: without .golangci.yml the gid* linters are simply not enabled, which
// is why the distributable config had to be copied next to the binary and kept
// in sync with it by hand. Here the same file is embedded into the binary,
// written into the user cache directory on demand and passed as --config.
//
// The rule is conservative: anything that looks like an existing config wins.
// The search repeats what pkg/config does (.golangci.* from the target
// directory up to the filesystem root, then $HOME), and it starts not only from
// the working directory but from every path argument on the command line — an
// over-eager find only means the binary keeps its stock behaviour, while a
// missed one would override the config the repository actually has.
package defaultconfig

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// The subcommands that read a config file. For the rest (version, cache,
// custom, help) an injected --config would be an unknown flag.
const (
	commandRun        = "run"
	commandFmt        = "fmt"
	commandConfig     = "config"
	commandLinters    = "linters"
	commandFormatters = "formatters"
)

// The flags through which the user decides about the config themselves.
const (
	flagConfig      = "--config"
	flagConfigShort = "-c"
	flagNoConfig    = "--no-config"
)

// configExts — the extensions golangci-lint accepts for its config file.
var configExts = []string{".yml", ".yaml", ".toml", ".json"}

// configAwareCommands — the subcommands Inject may add --config to.
var configAwareCommands = map[string]struct{}{
	commandRun:        {},
	commandFmt:        {},
	commandConfig:     {},
	commandLinters:    {},
	commandFormatters: {},
}

// Inject returns the command line to execute. It is args unchanged whenever
// golangci-lint has a config to read — an explicit --config, --no-config, a
// .golangci.* file above the checked packages, a subcommand that reads no
// config at all. Otherwise the built-in config is written into the user cache
// directory and --config pointing at it is inserted after the subcommand;
// path is the file written, and stays empty when nothing was injected. On an
// error out is nil: the caller runs with the command line it already has.
func Inject(args []string, config []byte) (out []string, path string, err error) {
	if len(config) == 0 || !needsConfig(args) {
		return args, "", nil
	}

	path, err = materialize(config)
	if err != nil {
		return nil, "", err
	}

	return insertConfigFlag(args, path), path, nil
}

// needsConfig reports whether this command line would run without any config.
func needsConfig(args []string) bool {
	command, rest := split(args)
	if _, ok := configAwareCommands[command]; !ok {
		return false
	}

	return !hasConfigFlag(rest) && findConfig(searchDirs(rest)) == ""
}

// split separates the subcommand from its arguments. The program name and any
// flag preceding the subcommand (--verbose linters) are dropped.
func split(args []string) (command string, rest []string) {
	if len(args) > 0 {
		args = args[1:]
	}

	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			return arg, args[i+1:]
		}
	}

	return "", nil
}

// hasConfigFlag reports whether the user already decided where the config comes
// from: --config/-c names a file, --no-config forbids one.
func hasConfigFlag(args []string) bool {
	for _, arg := range args {
		name, _, _ := strings.Cut(arg, "=")

		switch name {
		case flagConfig, flagConfigShort, flagNoConfig:
			return true
		}
	}

	return false
}

// searchDirs — the directories the config search starts from: the working
// directory plus every path argument, resolved the way pkg/config resolves its
// first argument (./... is not a directory, so its parent is taken).
func searchDirs(args []string) []string {
	dirs := []string{"."}

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}

		dirs = append(dirs, packageDir(arg))
	}

	return dirs
}

func packageDir(arg string) string {
	abs, err := filepath.Abs(arg)
	if err != nil {
		return filepath.Clean(arg)
	}

	if info, statErr := os.Stat(abs); statErr == nil && info.IsDir() {
		return abs
	}

	return filepath.Dir(abs)
}

// findConfig returns the first .golangci.* file golangci-lint would read,
// walking each start directory up to the filesystem root and then $HOME.
func findConfig(dirs []string) string {
	for _, dir := range dirs {
		if path := findUp(dir); path != "" {
			return path
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return configIn(home)
}

func findUp(dir string) string {
	current, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}

	for {
		if path := configIn(current); path != "" {
			return path
		}

		parent := filepath.Dir(current)
		if parent == current {
			return ""
		}

		current = parent
	}
}

func configIn(dir string) string {
	// The name golangci-lint looks for, without an extension
	// (pkg/config: viper.SetConfigName(".golangci")).
	const configBaseName = ".golangci"

	for _, ext := range configExts {
		path := filepath.Join(dir, configBaseName+ext)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}

	return ""
}

// insertConfigFlag puts --config right after the subcommand, where it is a flag
// of that subcommand rather than of the root command.
func insertConfigFlag(args []string, path string) []string {
	flag := flagConfig + "=" + path

	for i, arg := range args {
		if i > 0 && !strings.HasPrefix(arg, "-") {
			out := make([]string, 0, len(args)+1)
			out = append(out, args[:i+1]...)
			out = append(out, flag)

			return append(out, args[i+1:]...)
		}
	}

	return append(append([]string{}, args...), flag)
}

// materialize writes the built-in config into the user cache directory and
// returns its path. One fixed file, rewritten whenever the binary carries a
// different config — an upgraded binary replaces it rather than leaving the
// previous revision behind. Anyone who needs a particular config passes
// --config with its own path and never touches this one.
func materialize(config []byte) (string, error) {
	const (
		// cacheDirName — the subdirectory of the user cache directory holding
		// the materialized config, and the name of the file itself.
		cacheDirName = "gid-golangci"

		dirPerm = 0o755
	)

	dir, err := os.UserCacheDir()
	if err != nil {
		dir = os.TempDir()
	}

	dir = filepath.Join(dir, cacheDirName)
	if mkErr := os.MkdirAll(dir, dirPerm); mkErr != nil {
		return "", errors.Wrap(mkErr, "create the cache directory")
	}

	path := filepath.Join(dir, cacheDirName+".yml")

	if current, readErr := os.ReadFile(path); readErr == nil && bytes.Equal(current, config) {
		return path, nil
	}

	if writeErr := writeAtomic(path, config); writeErr != nil {
		return "", writeErr
	}

	return path, nil
}

// writeAtomic writes through a temporary neighbour, so a parallel run never
// reads a half-written config.
func writeAtomic(path string, config []byte) error {
	const filePerm = 0o644

	file, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*")
	if err != nil {
		return errors.Wrap(err, "create the temporary config")
	}

	tmp := file.Name()

	defer func() {
		_ = file.Close()
		_ = os.Remove(tmp)
	}()

	if _, err = file.Write(config); err != nil {
		return errors.Wrap(err, "write the temporary config")
	}

	if err = file.Close(); err != nil {
		return errors.Wrap(err, "close the temporary config")
	}

	if err = os.Chmod(tmp, filePerm); err != nil {
		return errors.Wrap(err, "set the config permissions")
	}

	if err = os.Rename(tmp, path); err != nil {
		return errors.Wrap(err, "move the config into place")
	}

	return nil
}
