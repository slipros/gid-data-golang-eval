// Eval of GID-272 settings.exclude: the same violations as in the main
// fixture, cleared only because the functions are on the exclusion list
// ("legacyConvert" — a bare Function name, "Converter.legacyMethod" —
// Type.Method). The rest of the package is still judged.
package model

func legacyConvert(a, b, c, d int) {}

// Converter owns the excluded method.
type Converter struct{}

func (conv *Converter) legacyMethod(a, b, c, d string) {}

func stillJudged(a, b, c, d int) {} // want `GID-272: function stillJudged takes 4 substantive arguments \(allowed: 3\)`
