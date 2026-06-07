// Не исключён — репортится даже при exclude для kafka.
package validate // want `GID-164: validate-пакет "excluded/server/http/handler/validate" обязан использовать github\.com/raoptimus/validator\.go/v2`

func Request(raw []byte) bool { return len(raw) > 0 }
