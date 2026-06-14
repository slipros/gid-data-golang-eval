// Helper package for the GID-196 eval: a conversion via a selector.
package sub

type Code string

func (c Code) Upper() string { return string(c) }
