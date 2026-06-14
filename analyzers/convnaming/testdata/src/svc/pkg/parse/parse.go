// Not applicable: outside the dal/domain/server/event layers the pattern is not checked.
package parse

type Config struct{ Raw string }

func ConfigFromEnv(raw string) Config { return Config{Raw: raw} }
