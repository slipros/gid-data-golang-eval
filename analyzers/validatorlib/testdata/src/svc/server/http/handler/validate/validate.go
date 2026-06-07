// Позитив: validate-пакет с ручной валидацией без validator.go.
package validate // want `GID-164: validate-пакет "svc/server/http/handler/validate" обязан использовать github\.com/raoptimus/validator\.go/v2 \(исключения: nolint или settings\.exclude\)`

import "strings"

func CreateSnapshot(name string) bool {
	return strings.TrimSpace(name) != ""
}
