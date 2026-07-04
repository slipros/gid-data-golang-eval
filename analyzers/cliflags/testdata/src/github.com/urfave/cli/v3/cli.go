// Package cli is a minimal stub of github.com/urfave/cli/v3 for the eval:
// just enough of the flag surface (Name, Required, Value, EnvVars, Sources)
// to exercise the analyzer.
package cli

// ValueSourceChain is the type of the Sources field in v3.
type ValueSourceChain struct {
	envVars []string
}

// EnvVars builds a ValueSourceChain from environment variable names — the
// v3 replacement for the v2 "EnvVars []string" field.
func EnvVars(vars ...string) ValueSourceChain {
	return ValueSourceChain{envVars: vars}
}

// Flag is the common interface every flag type satisfies.
type Flag interface {
	flagName() string
}

type StringFlag struct {
	Name     string
	Usage    string
	Required bool
	Value    string
	EnvVars  []string
	Sources  ValueSourceChain
}

func (f *StringFlag) flagName() string { return f.Name }

type IntFlag struct {
	Name     string
	Usage    string
	Required bool
	Value    int
	EnvVars  []string
	Sources  ValueSourceChain
}

func (f *IntFlag) flagName() string { return f.Name }

type BoolFlag struct {
	Name     string
	Usage    string
	Required bool
	Value    bool
	EnvVars  []string
	Sources  ValueSourceChain
}

func (f *BoolFlag) flagName() string { return f.Name }

// Command is a minimal stand-in for the app/command wiring that consumes flags.
type Command struct {
	Name  string
	Flags []Flag
}
