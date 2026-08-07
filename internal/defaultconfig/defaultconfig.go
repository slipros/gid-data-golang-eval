// Package defaultconfig makes the gid ruleset the default of custom-gcl: the
// binary carries its own config and runs with it unless told otherwise.
//
// custom-gcl is not a drop-in golangci-lint — it is golangci-lint plus the gid
// rules, and those rules only exist as the config that enables them. Left to
// itself, golangci-lint would read whatever .golangci.yml happens to sit in the
// repository and quietly run without a single gid linter. So the distributable
// config is embedded into the binary, written into the user cache directory on
// demand and passed as --config on every run.
//
// The user stays in charge: --config picks another file (a repository config is
// used exactly this way), --no-config drops to the stock golangci-lint set.
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

// Inject returns the command line to execute: the built-in config written into
// the user cache directory and --config pointing at it, inserted after the
// subcommand. args comes back unchanged when the user has decided about the
// config themselves (--config, --no-config) or the subcommand reads no config
// at all; path is the file written, and stays empty in those cases. On an
// error out is nil: the caller runs with the command line it already has.
func Inject(args []string, config []byte) (out []string, path string, err error) {
	if len(config) == 0 || !wantsDefault(args) {
		return args, "", nil
	}

	path, err = materialize(config)
	if err != nil {
		return nil, "", err
	}

	return insertConfigFlag(args, path), path, nil
}

// wantsDefault reports whether this command line leaves the choice of config to
// the binary.
func wantsDefault(args []string) bool {
	command, rest := split(args)
	if _, ok := configAwareCommands[command]; !ok {
		return false
	}

	return !hasConfigFlag(rest)
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

// LocalConfig returns the .golangci.* file of the working directory, or an
// empty string when there is none. That file is NOT what a plain run uses — it
// is named in the notice so that nobody mistakes the built-in config for it.
func LocalConfig() string {
	// The name golangci-lint looks for, without an extension
	// (pkg/config: viper.SetConfigName(".golangci")).
	const configBaseName = ".golangci"

	for _, ext := range configExts {
		path := configBaseName + ext
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
