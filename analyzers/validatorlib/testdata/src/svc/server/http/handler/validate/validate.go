// Positive: a validate package with manual validation and no validator.go.
package validate // want `GID-164: validate package "svc/server/http/handler/validate" must use github\.com/raoptimus/validator\.go/v2\. Fix: import it \(exceptions: nolint or settings\.exclude\)`

import "strings"

func CreateSnapshot(name string) bool {
	return strings.TrimSpace(name) != ""
}
