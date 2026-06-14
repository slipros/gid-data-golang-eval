// Eval GID-180: negative and boundary cases — no diagnostic is emitted.
package good

import "os"

var cfg = map[string]string{}

// Negative: constructing a map in init — deterministic, ok.
func init() {
	cfg["a"] = "1"
	cfg["b"] = "2"
}

// Negative: reading env in init is allowed (it is not I/O).
func init() {
	cfg["host"] = os.Getenv("HOST")
	if v, ok := os.LookupEnv("PORT"); ok {
		cfg["port"] = v
	}
}

// Negative: a go statement in an ordinary function — the rule does not apply to non-init.
func StartWorker() {
	go func() {}()
}

// Boundary (limitation): a helper outside init with os.Open, called from init,
// is NOT matched — the analysis is intraprocedural, the helper's body is not walked as init.
func loadFile() {
	_, _ = os.Open("/etc/hosts")
}

func init() {
	loadFile()
}
