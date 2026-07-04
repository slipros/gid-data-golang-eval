// Stub of an in-house side-effect-bearing library for the settings.packages eval.
package somelib

type Value struct {
	Data string
}

func (v Value) String() string { return v.Data }
